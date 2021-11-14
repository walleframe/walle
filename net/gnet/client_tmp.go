package gnet

import (
	"encoding/binary"
	"io"
	net "net"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"github.com/smallnest/goframe"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// ClientOption
//go:generate gogen option -n ClientOption -f Client -o option.client.go
func walleClient() interface{} {
	return map[string]interface{}{
		"Network": string("tcp"),
		// Addr Server Addr
		"Addr": string("localhost:8080"),
		// Process Options
		"ProcessOptions": []process.ProcessOption{},
		// process router
		"Router": Router(nil),
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.Default),
		// AutoReconnect auto reconnect server. zero means not reconnect!
		"AutoReconnectTime": int(5),
		// StopImmediately when session finish,business finish immediately.
		"StopImmediately": false,
		//
		"EncodeConfig": (*goframe.EncoderConfig)(DefaultClientEncodeConfig),
		"DecodeConfig": (*goframe.DecoderConfig)(DefaultClientDecodeConfig),
	}
}

var DefaultClientEncodeConfig = &goframe.EncoderConfig{
	ByteOrder:                       binary.LittleEndian,
	LengthFieldLength:               4,
	LengthAdjustment:                0,
	LengthIncludesLengthFieldLength: false,
}
var DefaultClientDecodeConfig = &goframe.DecoderConfig{
	ByteOrder:           binary.LittleEndian,
	LengthFieldOffset:   0,
	LengthFieldLength:   4,
	LengthAdjustment:    0,
	InitialBytesToStrip: 4,
}

type GoFrameClient struct {
	*process.Process
	opts  *ClientOptions
	conn  net.Conn
	fc    goframe.FrameConn
	close atomic.Bool
	send  chan []byte
}

func NewClientEx(inner *process.InnerOptions, cc *ClientOptions) (cli *GoFrameClient, err error) {
	cli = &GoFrameClient{}
	inner.Output = cli
	cli.opts = cc
	cli.Process = process.NewProcess(
		inner,
		process.NewProcessOptions(cc.ProcessOptions...),
	)
	cli.Process.Inner.ApplyOption(
		process.WithInnerOptionsOutput(cli),
		process.WithInnerOptionsBindData(cli),
		process.WithInnerOptionsNewContext(cli.newContext),
	)
	cli.conn, err = net.Dial("tcp", cli.opts.Addr)
	if err != nil {
		return
	}

	cli.fc = goframe.NewLengthFieldBasedFrameConn(
		*cli.opts.EncodeConfig,
		*cli.opts.DecodeConfig,
		cli.conn,
	)

	go cli.readLoop()
	go cli.writeLoop()

	return
}

func NewClient(opts ...ClientOption) (_ Client, err error) {
	return NewClientEx(process.NewInnerOptions(), NewClientOptions(opts...))
}

func (c *GoFrameClient) Write(in []byte) (n int, err error) {
	if c.close.Load() {
		err = packet.ErrSessionClosed
		return
	}

	if c.conn == nil {
		// TODO: 断线重连处理
		zaplog.Default.Info5("on reconncet")
		return
	}

	c.send <- in

	// // write msg
	// log := c.Process.Opts.Logger
	// n = len(in)
	// err = c.fc.WriteFrame(in)
	// if err != nil {
	// 	log.Error3("write message failed", zap.Error(err))
	// }

	return
}

func (c *GoFrameClient) Close() (err error) {
	if !c.close.CAS(false, true) {
		return
	}
	zaplog.Default.Info5("hand close conn")
	close(c.send)
	// c.cancel()
	// c.conn.Close()
	return c.fc.Close()
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (c *GoFrameClient) GetConn() interface{} {
	return c.conn
}

func (c *GoFrameClient) writeLoop() {
	c.send = make(chan []byte, 1024)
	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				return
			}
			c.fc.WriteFrame(data)
		}
	}
}

func (c *GoFrameClient) readLoop() {
	head := make([]byte, c.opts.DecodeConfig.LengthFieldLength)
	size := uint32(0)
	defer c.Close()
	for {
		_, err := io.ReadFull(c.conn, head)
		if err != nil {
			c.opts.Logger.Error3("read head error", zap.Error(err))
			c.Close()
			return
		}
		switch c.opts.DecodeConfig.LengthFieldLength {
		case 2:
			size = uint32(c.opts.DecodeConfig.ByteOrder.Uint16(head))
		case 4:
			size = c.opts.DecodeConfig.ByteOrder.Uint32(head)
		case 8:
			size = uint32(c.opts.DecodeConfig.ByteOrder.Uint64(head))
		default:
			c.opts.Logger.Error3("invalid head length", zap.Int("l", c.opts.DecodeConfig.LengthFieldLength))
			c.Close()
			return
		}
		// TODO 内存复用
		buf := make([]byte, size)
		_, err = io.ReadFull(c.conn, buf)
		if err != nil {
			c.opts.Logger.Error3("read body error", zap.Error(err))
			c.Close()
			return
		}
		c.Process.OnRead(buf)

		// data, err := c.fc.ReadFrame()
		// if err != nil {
		// 	fmt.Println("read error", err)
		// 	c.Close()
		// 	panic("xxxx")
		// 	return
		// }

		// c.Process.OnRead(data)
	}
}

func (c *GoFrameClient) newContext(ctx process.Context, ud interface{}) process.Context {
	return &clientCtx{
		Context:       ctx,
		GoFrameClient: c,
	}
}

type clientCtx struct {
	process.Context
	*GoFrameClient
}

var _ ClientContext = &clientCtx{}
