package gotcp

import (
	"context"
	"encoding/binary"
	"io"
	"math"
	net "net"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// ClientOption
//go:generate gogen option -n ClientOption -f Client -o option.client.go
func walleClient() interface{} {
	return map[string]interface{}{
		// Network tcp/tcp4/tcp6/unix
		"Network": "tcp",
		// Addr Server Addr
		"Addr": string("localhost:8080"),
		// Dialer config net dialer
		"Dialer": func(network, addr string) (conn net.Conn, err error) {
			return net.Dial(network, addr)
		},
		// Process Options
		"ProcessOptions": []process.ProcessOption{},
		// process router
		"Router": Router(nil),
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.Default),
		// AutoReconnect auto reconnect server. zero means not reconnect!
		"AutoReconnectTime": int(5),
		// AutoReconnectWait reconnect wait time
		"AutoReconnectWait": time.Duration(time.Millisecond * 500),
		// StopImmediately when session finish,business finish immediately.
		"StopImmediately": false,
		// ReadTimeout read timeout
		"ReadTimeout": time.Duration(0),
		// WriteTimeout write timeout
		"WriteTimeout": time.Duration(0),
		// Write network data method.
		"WriteMethods": WriteMethod(WriteAsync),
		// SendQueueSize async send queue size
		"SendQueueSize": int(1024),
		// Heartbeat use websocket ping/pong.
		"Heartbeat": time.Duration(0),
		// tcp packet head
		"PacketHeadBuf": func() []byte {
			return make([]byte, 4)
		},
		// read tcp packet head size
		"ReadSize": func(head []byte) (size int) {
			size = int(binary.LittleEndian.Uint32(head))
			return
		},
		// write tcp packet head size
		"WriteSize": func(head []byte, size int) (err error) {
			if size >= math.MaxUint32 {
				return packet.ErrPacketTooLarge
			}
			binary.LittleEndian.PutUint32(head, uint32(size))
			return
		},
		// ReadBufferSize 一定要大于最大消息的大小.每个链接一个缓冲区。
		"ReadBufferSize": int(65535),
		// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
		// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
		// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
		// 默认为false,是为了防止错误的配置导致bug。
		"ReuseReadBuffer": false,
		// MaxMessageSizeLimit limit message size
		"MaxMessageSizeLimit": int(0),
	}
}

// go client
type GoClient struct {
	// process
	*process.Process
	// net conn
	conn net.Conn
	// session context
	ctx    context.Context
	cancel func()
	// flag
	close  atomic.Bool
	reconn atomic.Bool
	send   chan []byte
	mux    sync.Mutex
	//
	writeMethod WriteMethod
	opts        *ClientOptions
	// close call back
	closeChain []func(Client)
}

// NewClientEx 创建客户端
// inner *process.InnerOptions 选项应该由上层ClientProxy去决定如何设置。
// copts 内部应该设置链接相关的参数。比如读写超时，如何发送数据
// opts 业务方决定
func NewClientEx(inner *process.InnerOptions, copts *ClientOptions) (cli *GoClient, err error) {
	// check option limit
	if copts.MaxMessageSizeLimit > copts.ReadBufferSize {
		copts.ReadBufferSize = copts.MaxMessageSizeLimit
	}
	if copts.MaxMessageSizeLimit == 0 {
		copts.MaxMessageSizeLimit = copts.ReadBufferSize
	}
	// modify limit for write check
	copts.MaxMessageSizeLimit -= len(copts.PacketHeadBuf())

	cli = &GoClient{
		Process: process.NewProcess(
			inner,
			process.NewProcessOptions(copts.ProcessOptions...),
		),
	}
	cli.Process.Inner.ApplyOption(
		process.WithInnerOptionsOutput(cli),
		process.WithInnerOptionsBindData(cli),
		process.WithInnerOptionsNewContext(cli.newContext),
	)
	cli.opts = copts
	cli.ctx = context.Background()
	cli.cancel = func() {}
	if copts.StopImmediately {
		cli.ctx, cli.cancel = context.WithCancel(context.Background())
	}

	// connect to server
	cli.conn, err = copts.Dialer(copts.Network, copts.Addr)
	if err != nil {
		return
	}

	go cli.Run()
	return cli, nil
}

func NewClient(opts ...ClientOption) (_ Client, err error) {
	return NewClientEx(process.NewInnerOptions(), NewClientOptions(opts...))
}

func (sess *GoClient) Write(in []byte) (n int, err error) {
	log := sess.Process.Opts.Logger
	if len(in) >= sess.opts.MaxMessageSizeLimit {
		err = packet.ErrPacketTooLarge
		return
	}
	// select {
	// case <-sess.ctx.Done():
	// 	err = packet.ErrSessionClosed
	// 	return
	// default:
	// 	// ok
	// }

	if sess.close.Load() || sess.reconn.Load() {
		log.Error3("client closed")
		err = packet.ErrSessionClosed
		return
	}
	// async write
	if sess.opts.WriteMethods == WriteAsync {
		sess.send <- in
		n = len(in)
		return
	}
	// sync write
	sess.mux.Lock()
	sess.mux.Unlock()
	if sess.opts.WriteTimeout > 0 {
		sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
	}
	n, err = sess.conn.Write(in)
	if err != nil {
		log.Error3("write message failed", zap.Error(err))
		return
	}

	return
}

func (sess *GoClient) Close() (err error) {
	if !sess.close.CAS(false, true) {
		return
	}
	sess.cancel()
	if sess.send != nil {
		close(sess.send)
	}

	return
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (sess *GoClient) GetConn() interface{} {
	return sess.conn
}

// Run run client
func (sess *GoClient) Run() {
	log := sess.opts.Logger
	for {
		reconnectTimeLimit := sess.opts.AutoReconnectTime
		wg := sync.WaitGroup{}
		// async write
		if sess.opts.WriteMethods == WriteAsync {
			sess.send = make(chan []byte, sess.opts.SendQueueSize)
			wg.Add(1)
			go func() {
				defer wg.Done()
				sess.writeLoop()
			}()
		}
		sess.readLoop()
		wg.Wait()

		sess.Process.Clean()
		// enable reconnect server
		if reconnectTimeLimit == 0 {
			break
		}
		if sess.close.Load() {
			break
		}
		sess.conn.Close()
		sess.conn = nil
		sess.reconn.Store(true)
		for k := 0; k < reconnectTimeLimit; k++ {
			time.Sleep(sess.opts.AutoReconnectWait)
			if sess.close.Load() {
				break
			}
			conn, err := sess.opts.Dialer(sess.opts.Network, sess.opts.Addr)
			if err != nil {
				log.Error3("reconnect server failed",
					zap.String("network", sess.opts.Network),
					zap.String("addr", sess.opts.Addr),
					zap.Error(err),
				)
				continue
			}
			sess.conn = conn
		}
		sess.reconn.Store(false)
		// check
		if sess.close.Load() {
			if sess.conn != nil {
				sess.conn.Close()
			}
			break
		}
		if sess.conn == nil {
			break
		}
	}
	for _, ntf := range sess.closeChain {
		ntf(sess)
	}
	log.Develop8("connect closed")
}

func (sess *GoClient) writeLoop() {
	log := sess.Process.Opts.Logger
	var err error
	hb := make([][]byte, 0, 16)
	hb = append(hb, sess.opts.PacketHeadBuf())
	// defer sess.Close()
	for {
		select {
		case <-sess.ctx.Done():
			for _ = range sess.send {
				// TODO drop message notify
			}
			return
		case data, ok := <-sess.send:
			if !ok {
				sess.conn.Close()
				return
			}
			err = sess.opts.WriteSize(hb[0], len(data))
			if err != nil {
				log.Error3("write message size failed", zap.Error(err))
				return
			}
			buf := net.Buffers{hb[0], data}
			for k := 0; k < len(sess.send); k++ {
				data := <-sess.send
				if k+1 <= len(hb) {
					hb = append(hb, sess.opts.PacketHeadBuf())
				}
				err = sess.opts.WriteSize(hb[k+1], len(data))
				if err != nil {
					log.Error3("write message size failed", zap.Error(err))
					return
				}
				buf = append(buf, hb[k+1])
				buf = append(buf, data)
			}
			if sess.opts.WriteTimeout > 0 {
				sess.conn.SetWriteDeadline(time.Now().Add(sess.opts.WriteTimeout))
			}
			// writev
			_, err = buf.WriteTo(sess.conn)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
					// retry
				}
				log.Error3("write data failed", zap.Error(err))
				return
			}
		}
	}
}

func (sess *GoClient) readLoop() {
	log := sess.Process.Opts.Logger
	headSize := len(sess.opts.PacketHeadBuf())
	buf := make([]byte, sess.opts.ReadBufferSize)
	bufSize := 0
	// defer sess.Close()
	for {
		if sess.opts.ReadTimeout > 0 {
			sess.conn.SetReadDeadline(time.Now().Add(sess.opts.ReadTimeout))
		}
		read, err := sess.conn.Read(buf[bufSize:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
				// heatbeat
				continue
			}
			if err == io.EOF {
				log.Develop8("io.EOF. connect close")
				return
			}
			log.Error3("read head error", zap.Error(err))
			return
		}
		bufSize += read

		for {
			if bufSize < headSize {
				// wait left data
				break
			}
			size := sess.opts.ReadSize(buf[:headSize])
			if size > sess.opts.MaxMessageSizeLimit {
				log.Error3("invalid packet", zap.Int("size", size))
				return
			}
			if bufSize < size+headSize {
				// wait left data
				break
			}
			// NOTE: 复用缓冲区，不能设置Process.DispatchFilter（开启新协程）,或者应该在内部手动复制内存。
			if sess.opts.ReuseReadBuffer {
				sess.Process.OnRead(buf[headSize : headSize+size])
			} else {
				cache := make([]byte, size)
				copy(cache, buf[headSize:headSize+size])
				sess.Process.OnRead(cache)
			}
			copy(buf, buf[headSize+size:])
			bufSize -= headSize + size
		}
	}
}

// WithValue wrap context.WithValue
func (sess *GoClient) WithSessionValue(key, value interface{}) {
	sess.ctx = context.WithValue(sess.ctx, key, value)
	return
}

// Value wrap context.Context.Value
func (sess *GoClient) SessionValue(key interface{}) interface{} {
	return sess.ctx.Value(key)
}

func (sess *GoClient) AddCloseFunc(f func(sess Client)) {
	sess.closeChain = append(sess.closeChain, f)
}

func (sess *GoClient) newContext(ctx process.Context, ud interface{}) process.Context {
	return &clientCtx{
		Context:  ctx,
		GoClient: sess,
	}
}

type clientCtx struct {
	process.Context
	*GoClient
}

var _ ClientContext = &clientCtx{}
