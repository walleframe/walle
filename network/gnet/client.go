package gnet

import (
	"context"
	"encoding/binary"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/walleframe/walle/network"
	"github.com/walleframe/walle/network/rpc"
	process "github.com/walleframe/walle/process"
	"github.com/walleframe/walle/process/errcode"
	zaplog "github.com/walleframe/walle/zaplog"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// ClientOption
//
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
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
		// AutoReconnect auto reconnect server. zero means not reconnect!
		"AutoReconnectTime": int(5),
		// StopImmediately when session finish,business finish immediately.
		"StopImmediately": false,
		// WithMulticore sets up multi-cores in gnet server.
		"Multicore": false,
		// WithLockOSThread sets up LockOSThread mode for I/O event-loops.
		"LockOSThread": false,
		// WithLoadBalancing sets up the load-balancing algorithm in gnet server.
		"LoadBalancing": gnet.LoadBalancing(gnet.SourceAddrHash),
		// WithNumEventLoop sets up NumEventLoop in gnet server.
		"NumEventLoop": int(0),
		// WithReusePort sets up SO_REUSEPORT socket option.
		"ReusePort": false,
		// WithTCPKeepAlive sets up the SO_KEEPALIVE socket option with duration.
		"TCPKeepAlive": time.Duration(0),
		// WithTCPNoDelay enable/disable the TCP_NODELAY socket option.
		"TCPNoDelay": gnet.TCPSocketOpt(gnet.TCPNoDelay),
		// WithReadBufferCap sets up ReadBufferCap for reading bytes.
		"ReadBufferCap": int(0),
		// WithSocketRecvBuffer sets the maximum socket receive buffer in bytes.
		"SocketRecvBuffer": int(0),
		// WithSocketSendBuffer sets the maximum socket send buffer in bytes.
		"SocketSendBuffer": int(0),
		// WithTicker indicates that a ticker is set.
		"Ticker": time.Duration(0),
		// BlockConnect 创建客户端时候，是否阻塞等待链接服务器
		"BlockConnect": true,
		// Write network data method.
		"WriteMethods": WriteMethod(WriteAsync),
		// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
		// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
		// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
		"ReuseReadBuffer": true,
	}
}

// GNetClient gnet.Client 封装
type GNetClient struct {
	*rpc.RPCProcess
	opts *ClientOptions
	cli  *gnet.Client
	conn gnet.Conn
	// session context
	ctx    context.Context
	cancel func()
	close  atomic.Bool
	// close call back
	closeChain []func(Client)
}

var _ network.Link = &GNetClient{}

// NewClientEx 创建客户端
// inner *process.InnerOptions 选项应该由上层ClientProxy去决定如何设置。
// copts 内部应该设置链接相关的参数。比如读写超时，如何发送数据
// opts 业务方决定
func NewClientEx(inner *process.InnerOptions, copts *ClientOptions) (cli *GNetClient, err error) {

	if copts.Router != nil {
		inner.Router = copts.Router
	}

	cli = &GNetClient{
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

	cli.cli, err = gnet.NewClient(cli, convertClientOptions(cli.opts)...)
	if err != nil {
		return
	}
	cli.cli.Start()

	// block connect to server
	if copts.BlockConnect {
		cli.conn, err = cli.cli.Dial(copts.Network, copts.Addr)
		if err != nil {
			return
		}
		cli.conn.SetContext(cli)
	}

	return cli, nil
}

func NewClient(opts ...ClientOption) (_ Client, err error) {
	return NewClientEx(process.NewInnerOptions(), NewClientOptions(opts...))
}

// NewClientForProxy new client for client proxy.
// NOTE: you should rewrite this function for custom set option
func NewClientForProxy(net, addr string, inner *process.InnerOptions) (Client, error) {
	return NewClientEx(inner, NewClientOptions(
		WithClientOptionNetwork(net),
		WithClientOptionAddr(addr),
	))
}

func (sess *GNetClient) logger(fname string) *zaplog.LogEntities {
	return sess.opts.FrameLogger.New(fname)
}

func (c *GNetClient) Write(in []byte) (n int, err error) {
	if c.close.Load() {
		err = errcode.ErrSessionClosed
		return
	}

	if c.conn == nil {
		// TODO: 断线重连处理
		c.logger("").Error("on reconncet")
		return
	}

	// write msg
	//n = len(in)
	// if c.udp {
	// 	err = c.conn.SendTo(in)
	// } else {
	n, err = c.conn.Write(in)
	//c.logger("gnet.write").Info("client write size", zap.Int("len", n))
	// }
	if err != nil {
		c.logger("gnet.write").Error("write message failed", zap.Error(err))
	}

	return
}

func (c *GNetClient) Close() (err error) {
	if !c.close.CAS(false, true) {
		return
	}
	c.logger("gnet.close").Info("hand close conn")
	// c.cancel()
	c.conn.Close()
	return c.cli.Stop()
}

// GetConn get raw conn(net.Conn,websocket.Conn...)
func (c *GNetClient) GetConn() interface{} {
	return c.conn
}

func (sess *GNetClient) AddCloseClientFunc(f func(sess Client)) {
	sess.closeChain = append(sess.closeChain, f)
}

// OnBoot fires when the engine is ready for accepting connections.
// The parameter engine has information and various utilities.
func (c *GNetClient) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return
}

// OnShutdown fires when the engine is being shut down, it is called right after
// all event-loops and connections are closed.
func (c *GNetClient) OnShutdown(eng gnet.Engine) {

}

// OnOpen fires when a new connection has been opened.
//
// The Conn c has information about the connection such as its local and remote addresses.
// The parameter out is the return value which is going to be sent back to the peer.
// Sending large amounts of data back to the peer in OnOpen is usually not recommended.
func (c *GNetClient) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	return
}

// OnClose fires when a connection has been closed.
// The parameter err is the last known connection error.
func (c *GNetClient) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	if c.close.Load() {
		return
	}
	c.conn = nil

	for k := 0; k < c.opts.AutoReconnectTime; k++ {
		newConn, err := c.cli.Dial(c.opts.Network, c.opts.Addr)
		if err != nil {
			// TODO 断线重连处理
			time.Sleep(time.Second)
			continue
		}
		c.conn = newConn
		break
	}
	return
}

// OnTraffic fires when a socket receives data from the peer.
//
// Note that the []byte returned from Conn.Peek(int)/Conn.Next(int) is not allowed to be passed to a new goroutine,
// as this []byte will be reused within event-loop after OnTraffic() returns.
// If you have to use this []byte in a new goroutine, you should either make a copy of it or call Conn.Read([]byte)
// to read data into your own []byte, then pass the new []byte to the new goroutine.
func (c *GNetClient) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	src := conn.Context()
	if src == nil {
		c.opts.FrameLogger.New("gnetserver.React").Info("react conn no context")
		action = gnet.Close
		return
	}
	sess, ok := src.(*GNetClient)
	if !ok {
		c.opts.FrameLogger.New("gnetserver.React").Info("react conn context not session")
		action = gnet.Close
		return
	}
	for {
		if conn.InboundBuffered() < 4 {
			break
		}
		buf, err := conn.Peek(4)
		if err != nil {
			c.opts.FrameLogger.New("gnetserver.React").Debug("read packet head", zap.Error(err))
			break
		}
		size := binary.BigEndian.Uint32(buf)
		buf, err = conn.Peek(int(size) + 4)
		if err != nil {
			c.opts.FrameLogger.New("gnetserver.React").Debug("read packet full", zap.Error(err))
			break
		}
		if c.opts.ReuseReadBuffer {
			sess.OnRead(buf)
			conn.Discard(int(size) + 4)
		} else {
			data := make([]byte, len(buf))
			copy(data, buf)
			sess.OnRead(data)
			conn.Discard(int(size) + 4)
		}
	}
	return
}

// OnTick fires immediately after the engine starts and will fire again
// following the duration specified by the delay return value.
func (c *GNetClient) OnTick() (delay time.Duration, action gnet.Action) {
	return
}

func convertClientOptions(opts *ClientOptions) (cfgs []gnet.Option) {
	cfgs = append(cfgs,
		gnet.WithMulticore(opts.Multicore),
		gnet.WithLockOSThread(opts.LockOSThread),
		gnet.WithLoadBalancing(opts.LoadBalancing),
		gnet.WithNumEventLoop(opts.NumEventLoop),
		gnet.WithReusePort(opts.ReusePort),
		gnet.WithTCPKeepAlive(opts.TCPKeepAlive),
		gnet.WithTCPNoDelay(opts.TCPNoDelay),
		gnet.WithReadBufferCap(opts.ReadBufferCap),
		gnet.WithSocketRecvBuffer(opts.SocketRecvBuffer),
		gnet.WithSocketSendBuffer(opts.SocketSendBuffer),
		gnet.WithTicker(opts.Ticker > 0),
		gnet.WithLogLevel(zap.DebugLevel),
	)
	// TODO: gnet log options
	return
}

type clientCtx struct {
	process.WrapContext
	*GNetClient
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
	ctx.GNetClient = inner.BindData.(*GNetClient)
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
