package gotcp

import (
	"context"
	"encoding/binary"
	"io"
	"math"
	net "net"
	"sync"
	"time"

	"github.com/aggronmagi/walle/network/rpc"
	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/process/errcode"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// ClientOption
//
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
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
		// AutoReconnect auto reconnect server. zero means not reconnect! -1 means always reconnect, >0 : reconnect times
		"AutoReconnectTime": int(-1),
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
				return errcode.ErrPacketsizeInvalid
			}
			binary.LittleEndian.PutUint32(head, uint32(size))
			return
		},
		// ReadBufferSize 一定要大于最大消息的大小.每个链接一个缓冲区。
		"ReadBufferSize": int(65535),
		// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
		// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
		// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
		"ReuseReadBuffer": true,
		// MaxMessageSizeLimit limit message size
		"MaxMessageSizeLimit": int(0),
		// BlockConnect 创建客户端时候，是否阻塞等待链接服务器
		"BlockConnect": true,
	}
}

// go client
type GoClient struct {
	// process
	*rpc.RPCProcess
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

	if copts.Router != nil {
		inner.Router = copts.Router
	}

	cli = &GoClient{
		RPCProcess: rpc.NewRPCProcess(
			inner,
			process.NewProcessOptions(copts.ProcessOptions...),
		),
	}
	cli.Inner.ApplyOption(
		process.WithInnerOptionOutput(cli),
		process.WithInnerOptionBindData(cli),
		process.WithInnerOptionContextPool(GoClientContextPool),
	)
	cli.opts = copts
	cli.ctx = context.Background()
	cli.cancel = func() {}
	if copts.StopImmediately {
		cli.ctx, cli.cancel = context.WithCancel(context.Background())
	}

	// block connect to server
	if copts.BlockConnect {
		cli.conn, err = copts.Dialer(copts.Network, copts.Addr)
		if err != nil {
			return
		}
	}

	// async write
	if cli.opts.WriteMethods == WriteAsync {
		cli.send = make(chan []byte, cli.opts.SendQueueSize)
	}

	go cli.Run()
	return cli, nil
}

func NewClient(opts ...ClientOption) (_ Client, err error) {
	return NewClientEx(process.NewInnerOptions(), NewClientOptions(opts...))
}

// NewClientForProxy new client for client proxy.
// NOTE: you should rewrite this function for custom set option
func NewClientForProxy(net, addr string, inner *process.InnerOptions) (Client, error) {
	return NewClientEx(inner, NewClientOptions(
		WithClientOptionsNetwork(net),
		WithClientOptionsAddr(addr),
	))
}

func (sess *GoClient) logger(fname string) *zaplog.LogEntities {
	return sess.opts.FrameLogger.New(fname)
}

func (sess *GoClient) Write(in []byte) (n int, err error) {
	log := sess.logger("goclient.Write")
	if len(in) >= sess.opts.MaxMessageSizeLimit {
		err = errcode.ErrPacketsizeInvalid
		log.Error("write msg too big", zap.Int("size", len(in)), zap.Int("limit", sess.opts.MaxMessageSizeLimit))
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
		log.Warn("client closed")
		err = errcode.ErrSessionClosed
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
		log.Error("write message failed", zap.Error(err))
		return
	}

	return
}

func (sess *GoClient) Close() (err error) {
	if !sess.close.CAS(false, true) {
		return
	}
	sess.cancel()
	// if sess.send != nil {
	// 	close(sess.send)
	// }

	return
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (sess *GoClient) GetConn() interface{} {
	return sess.conn
}

// Run run client
func (sess *GoClient) Run() {
	log := sess.logger("goclient.Run")

	// conn
	if !sess.opts.BlockConnect && !sess.reconnLoop() {
		return
	}

	for {

		wg := sync.WaitGroup{}
		// async write
		if sess.opts.WriteMethods == WriteAsync {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sess.writeLoop()
			}()
		}
		sess.readLoop()
		wg.Wait()

		sess.RPCProcess.Clean()
		// closed
		if sess.close.Load() {
			break
		}
		// 重连失败
		if !sess.reconnLoop() {
			break
		}
	}
	for _, ntf := range sess.closeChain {
		ntf(sess)
	}
	log.Debug("connect closed")
}

func (sess *GoClient) reconnLoop() (ok bool) {
	log := sess.logger("goclient.reconnLoop")
	if sess.opts.AutoReconnectTime == 0 {
		log.Debug("disable reconnect")
		return
	}

	if sess.close.Load() {
		log.Debug("client has closed")
		return
	}

	if sess.conn != nil {
		sess.conn.Close()
	}
	sess.conn = nil
	sess.reconn.Store(true)
	reconnectTimeLimit := sess.opts.AutoReconnectTime
	index := 0
	for {
		// connect to server
		conn, err := sess.opts.Dialer(sess.opts.Network, sess.opts.Addr)
		if err == nil {
			sess.conn = conn
			break
		}

		log.Error("reconnect server failed",
			zap.String("network", sess.opts.Network),
			zap.String("addr", sess.opts.Addr),
			zap.Error(err),
		)
		if sess.close.Load() {
			log.Debug("client has closed")
			return
		}
		time.Sleep(sess.opts.AutoReconnectWait)
		if sess.close.Load() {
			log.Debug("client has closed")
			return
		}
		index++
		if reconnectTimeLimit > 0 && index >= reconnectTimeLimit {
			log.Warn("reconn failed")
			break
		}
	}

	sess.reconn.Store(false)
	// check
	if sess.close.Load() {
		if sess.conn != nil {
			sess.conn.Close()
			sess.conn = nil
		}
		return
	}
	return sess.conn != nil
}

func (sess *GoClient) ClientValid() bool {
	return !sess.reconn.Load() && !sess.close.Load()
}

func (sess *GoClient) writeLoop() {
	log := sess.logger("goclient.writeLoop")
	var err error

	var buf net.Buffers = make([][]byte, 0, 32*2)
	// defer sess.Close()
	for {
		select {
		case <-sess.ctx.Done():
			for range sess.send {
				// TODO drop message notify
			}
			return
		case data, ok := <-sess.send:
			if !ok {
				sess.conn.Close()
				return
			}
			buf := buf[:0]
			buf = append(buf, data)
			for k := 0; k < len(sess.send); k++ {
				data := <-sess.send
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
				log.Error("write data failed", zap.Error(err))
				return
			}
		}
	}
}

// func (sess *GoClient) readLoop() {
// 	log := sess.logger("goclient.readLoop")
// 	headSize := len(sess.opts.PacketHeadBuf())
// 	buf := make([]byte, sess.opts.ReadBufferSize)
// 	bufSize := 0
// 	// defer sess.Close()
// 	for {
// 		if sess.opts.ReadTimeout > 0 {
// 			sess.conn.SetReadDeadline(time.Now().Add(sess.opts.ReadTimeout))
// 		}
// 		read, err := sess.conn.Read(buf[bufSize:])
// 		if err != nil {
// 			if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
// 				// heatbeat
// 				continue
// 			}
// 			if err == io.EOF {
// 				log.Debug("io.EOF. connect close")
// 				return
// 			}
// 			log.Error("read head error", zap.Error(err))
// 			return
// 		}
// 		bufSize += read

// 		for {
// 			if bufSize < headSize {
// 				// wait left data
// 				break
// 			}
// 			size := sess.opts.ReadSize(buf[:headSize])
// 			if size > sess.opts.MaxMessageSizeLimit {
// 				log.Error("invalid packet", zap.Int("size", size))
// 				return
// 			}
// 			if bufSize < size+headSize {
// 				// wait left data
// 				break
// 			}
// 			// NOTE: 复用缓冲区，不能设置Process.DispatchFilter（开启新协程）,或者应该在内部手动复制内存。
// 			if sess.opts.ReuseReadBuffer {
// 				sess.Process.OnRead(buf[headSize : headSize+size])
// 			} else {
// 				cache := make([]byte, size)
// 				copy(cache, buf[headSize:headSize+size])
// 				sess.Process.OnRead(cache)
// 			}
// 			copy(buf, buf[headSize+size:])
// 			bufSize -= headSize + size
// 		}
// 	}
// }

func (sess *GoClient) readLoop() {
	log := sess.logger("goclient.readLoop")
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
				log.Debug("io.EOF. connect close")
				return
			}
			log.Error("read head error", zap.Error(err))
			return
		}
		bufSize += read

		for {
			if bufSize < 4 {
				// wait left data
				break
			}
			size := int(binary.BigEndian.Uint32(buf[:4]))
			if size > sess.opts.MaxMessageSizeLimit {
				log.Error("invalid packet", zap.Int("size", size))
				return
			}
			if bufSize < size+4 {
				// wait left data
				break
			}
			// NOTE: 复用缓冲区，不能设置Process.DispatchFilter（开启新协程）,或者应该在内部手动复制内存。
			if sess.opts.ReuseReadBuffer {
				sess.Process.OnRead(buf[:4+size])
			} else {
				cache := make([]byte, size+4)
				copy(cache, buf[:4+size])
				sess.Process.OnRead(cache)
			}
			copy(buf, buf[4+size:])
			bufSize -= 4 + size
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

func (sess *GoClient) AddCloseClientFunc(f func(sess Client)) {
	sess.closeChain = append(sess.closeChain, f)
}

type clientCtx struct {
	process.WrapContext
	*GoClient
}

var _ ClientContext = &clientCtx{}

// process.ContextPool interface
type goClientContextPool struct {
	sync.Pool
}

func (p *goClientContextPool) NewContext(inner *process.InnerOptions, opts *process.ProcessOptions, inPkg interface{}, handlers []process.MiddlewareFunc, loadFlag bool) process.Context {
	ctx := p.Get().(*clientCtx)
	ctx.Inner = inner
	ctx.Opts = opts
	ctx.SrcContext = inner.ParentCtx
	ctx.Index = 0
	ctx.Handlers = handlers
	ctx.InPkg = inPkg
	ctx.LoadFlag = loadFlag
	ctx.Log = opts.Logger
	ctx.FreeContext = ctx
	ctx.GoClient = inner.BindData.(*GoClient)
	return ctx
}

func (p *goClientContextPool) FreeContext(ctx process.Context) {
	p.Put(ctx)
}

var GoClientContextPool process.ContextPool = &goClientContextPool{
	Pool: sync.Pool{
		New: func() interface{} {
			return &clientCtx{}
		},
	},
}
