package gnet

// TODO: Gnet 客户端 压测有bug。1k-3k之间，请求会异常断开
// import (
// 	"encoding/binary"
// 	"time"

// 	"github.com/aggronmagi/walle/net/iface"
// 	"github.com/aggronmagi/walle/net/packet"
// 	process "github.com/aggronmagi/walle/net/process"
// 	zaplog "github.com/aggronmagi/walle/zaplog"
// 	"github.com/panjf2000/gnet"
// 	"go.uber.org/atomic"
// 	"go.uber.org/zap"
// )

// // ClientOption
// // go:generate gogen option -n ClientOption -f Client -o option.client.go
// func walleClient() interface{} {
// 	return map[string]interface{}{
// 		"Network": string("tcp"),
// 		// Addr Server Addr
// 		"Addr": string("localhost:8080"),
// 		// Process Options
// 		"ProcessOptions": []process.ProcessOption{},
// 		// process router
// 		"Router": Router(nil),
// 		// log interface
// 		"Logger": (*zaplog.Logger)(zaplog.Default),
// 		// AutoReconnect auto reconnect server. zero means not reconnect!
// 		"AutoReconnectTime": int(5),
// 		// StopImmediately when session finish,business finish immediately.
// 		"StopImmediately": false,
// 		// WithMulticore sets up multi-cores in gnet server.
// 		"Multicore": false,
// 		// WithLockOSThread sets up LockOSThread mode for I/O event-loops.
// 		"LockOSThread": false,
// 		// WithLoadBalancing sets up the load-balancing algorithm in gnet server.
// 		"LoadBalancing": gnet.LoadBalancing(gnet.SourceAddrHash),
// 		// WithNumEventLoop sets up NumEventLoop in gnet server.
// 		"NumEventLoop": int(0),
// 		// WithReusePort sets up SO_REUSEPORT socket option.
// 		"ReusePort": false,
// 		// WithTCPKeepAlive sets up the SO_KEEPALIVE socket option with duration.
// 		"TCPKeepAlive": time.Duration(0),
// 		// WithTCPNoDelay enable/disable the TCP_NODELAY socket option.
// 		"TCPNoDelay": gnet.TCPSocketOpt(gnet.TCPNoDelay),
// 		// WithReadBufferCap sets up ReadBufferCap for reading bytes.
// 		"ReadBufferCap": int(0),
// 		// WithSocketRecvBuffer sets the maximum socket receive buffer in bytes.
// 		"SocketRecvBuffer": int(0),
// 		// WithSocketSendBuffer sets the maximum socket send buffer in bytes.
// 		"SocketSendBuffer": int(0),
// 		// WithTicker indicates that a ticker is set.
// 		"Ticker": time.Duration(0),
// 		// WithCodec sets up a codec to handle TCP stream.
// 		"Codec": gnet.ICodec(gnet.ICodec(DefaultGNetCodec)),
// 	}
// }

// // GNetClient gnet.Client 封装
// type GNetClient struct {
// 	*process.Process
// 	opts  *ClientOptions
// 	cli   *gnet.Client
// 	conn  gnet.Conn
// 	close atomic.Bool
// 	udp   bool
// }

// var _ iface.Link = &GNetClient{}

// // NewClientEx 创建客户端
// // inner *process.InnerOptions 选项应该由上层ClientProxy去决定如何设置。
// // svr 内部应该设置链接相关的参数。比如读写超时，如何发送数据
// // opts 业务方决定
// func NewClientEx(inner *process.InnerOptions, cc *ClientOptions, opts ...process.ProcessOption) (c *GNetClient, err error) {
// 	popt := process.NewProcessOptions(opts...)
// 	inner.Router = cc.Router
// 	inner.NewContext = c.newContext
// 	c = &GNetClient{}
// 	inner.Output = c
// 	c.opts = cc
// 	c.Process = process.NewProcess(inner, popt)
// 	c.cli, err = gnet.NewClient(c, convertClientOptions(c.opts)...)
// 	if err != nil {
// 		return
// 	}
// 	c.conn, err = c.cli.Dial(c.opts.Network, c.opts.Addr)
// 	if err != nil {
// 		return
// 	}
// 	err = c.cli.Start()
// 	return
// }

// func NewClient(cc *ClientOptions,
// 	opts ...process.ProcessOption) (_ Client, err error) {
// 	return NewClientEx(process.NewInnerOptions(), cc, opts...)
// }

// func (c *GNetClient) Write(in []byte) (n int, err error) {
// 	if c.close.Load() {
// 		err = packet.ErrSessionClosed
// 		return
// 	}

// 	if c.conn == nil {
// 		// TODO: 断线重连处理
// 		zaplog.Default.Info5("on reconncet")
// 		return
// 	}

// 	// write msg
// 	log := c.Process.Opts.Logger
// 	n = len(in)
// 	if c.udp {
// 		err = c.conn.SendTo(in)
// 	} else {
// 		err = c.conn.AsyncWrite(in)
// 	}
// 	if err != nil {
// 		log.Error3("write message failed", zap.Error(err))
// 	}

// 	return
// }

// func (c *GNetClient) Close() (err error) {
// 	if !c.close.CAS(false, true) {
// 		return
// 	}
// 	zaplog.Default.Info5("hand close conn")
// 	// c.cancel()
// 	c.conn.Close()
// 	return c.cli.Stop()
// }

// // GetConn get raw conn(net.Conn,websocket.Conn...)
// func (c *GNetClient) GetConn() interface{} {
// 	return c.conn
// }

// // OnInitComplete fires when the server is ready for accepting connections.
// // The parameter:server has information and various utilities.
// func (c *GNetClient) OnInitComplete(svr gnet.Server) (action gnet.Action) {
// 	return
// }

// // OnShutdown fires when the server is being shut down, it is called right after
// // all event-loops and connections are closed.
// func (c *GNetClient) OnShutdown(svr gnet.Server) {
// 	zaplog.Default.Info5("client shutdown", zap.Any("info", svr), zap.String("addr", c.opts.Addr))
// }

// // OnOpened fires when a new connection has been opened.
// // The parameter:c has information about the connection such as it's local and remote address.
// // Parameter:out is the return value which is going to be sent back to the client.
// func (c *GNetClient) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
// 	zaplog.Default.Info5("client connect", zap.String("addr", c.opts.Addr))
// 	return
// }

// // OnClosed fires when a connection has been closed.
// // The parameter:err is the last known connection error.
// func (c *GNetClient) OnClosed(conn gnet.Conn, err error) (action gnet.Action) {
// 	zaplog.Default.Info5("client closed", zap.Error(err), zap.String("addr", c.opts.Addr))
// 	if c.close.Load() {
// 		return
// 	}
// 	c.conn = nil

// 	for k := 0; k < c.opts.AutoReconnectTime; k++ {
// 		newConn, err := c.cli.Dial(c.opts.Network, c.opts.Addr)
// 		if err != nil {
// 			// TODO 断线重连处理
// 			time.Sleep(time.Second)
// 			continue
// 		}
// 		c.conn = newConn
// 		break
// 	}
// 	return
// }

// // PreWrite fires just before a packet is written to the peer socket, this event function is usually where
// // you put some code of logging/counting/reporting or any fore operations before writing data to client.
// func (c *GNetClient) PreWrite(conn gnet.Conn) {
// }

// // AfterWrite fires right after a packet is written to the peer socket, this event function is usually where
// // you put the []byte's back to your memory pool.
// func (c *GNetClient) AfterWrite(conn gnet.Conn, b []byte) {
// }

// // React fires when a connection sends the server data.
// // Call c.Read() or c.ReadN(n) within the parameter:c to read incoming data from client.
// // Parameter:out is the return value which is going to be sent back to the client.
// func (c *GNetClient) React(packet []byte, conn gnet.Conn) (out []byte, action gnet.Action) {
// 	c.Process.OnRead(packet)
// 	return
// }

// // Tick fires immediately after the server starts and will fire again
// // following the duration specified by the delay return value.
// func (c *GNetClient) Tick() (delay time.Duration, action gnet.Action) {
// 	return
// }

// func (c *GNetClient) newContext(ctx process.Context, ud interface{}) process.Context {
// 	return &clientCtx{
// 		Context:    ctx,
// 		GNetClient: c,
// 	}
// }

// type clientCtx struct {
// 	process.Context
// 	*GNetClient
// }

// var _ ClientContext = &clientCtx{}

// func convertClientOptions(opts *ClientOptions) (cfgs []gnet.Option) {
// 	cfgs = append(cfgs,
// 		gnet.WithMulticore(opts.Multicore),
// 		gnet.WithLockOSThread(opts.LockOSThread),
// 		gnet.WithLoadBalancing(opts.LoadBalancing),
// 		gnet.WithNumEventLoop(opts.NumEventLoop),
// 		gnet.WithReusePort(opts.ReusePort),
// 		gnet.WithTCPKeepAlive(opts.TCPKeepAlive),
// 		gnet.WithTCPNoDelay(opts.TCPNoDelay),
// 		gnet.WithReadBufferCap(opts.ReadBufferCap),
// 		gnet.WithSocketRecvBuffer(opts.SocketRecvBuffer),
// 		gnet.WithSocketSendBuffer(opts.SocketSendBuffer),
// 		gnet.WithTicker(opts.Ticker > 0),
// 		gnet.WithCodec(opts.Codec),
// 		gnet.WithLogLevel(zap.DebugLevel),
// 	)
// 	// TODO: gnet log options
// 	return
// }
