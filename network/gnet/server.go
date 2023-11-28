package gnet

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aggronmagi/walle/network"
	"github.com/aggronmagi/walle/network/rpc"
	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/process/metadata"
	"github.com/aggronmagi/walle/process/packet"
	"github.com/aggronmagi/walle/zaplog"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// import type
type (
	Router         = process.Router
	Server         = network.Server
	Session        = network.Session
	SessionContext = network.SessionContext
	Client         = network.Client
	ClientContext  = network.ClientContext
	WriteMethod    = network.WriteMethod
)

// import const value
const (
	WriteAsync       = network.WriteAsync
	WriteImmediately = network.WriteImmediately
)

// ServerOption
//
//go:generate gogen option -n ServerOption -o option.server.go
func walleServer() interface{} {
	return map[string]interface{}{
		// Addr Server Addr
		"Addr": string("tcp://0.0.0.0:8080"),
		// NetOption modify raw options
		"NetConnOption": func(net.Conn) {},
		// accepted load limit
		"AcceptLoadLimit": func(sess Session, cnt int64) bool { return false },
		// Process Options
		"ProcessOptions": []process.ProcessOption{},
		// process router
		"Router": Router(process.GetRouter()),
		// SessionRouter custom session router
		"SessionRouter": func(sess Session, global Router) (r Router) { return global },
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
		// SessionLogger custom session logger
		"SessionLogger": func(sess Session, global *zaplog.Logger) (r *zaplog.Logger) { return global },
		// NewSession custom session
		"NewSession": func(in Session) (Session, error) { return in, nil },
		// StopImmediately when session finish,business finish immediately.
		"StopImmediately": false,
		// Heartbeat use websocket ping/pong.
		"Heartbeat": time.Duration(0),
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
		// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
		// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
		// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
		// 默认为false,是为了防止错误的配置导致bug。
		"ReuseReadBuffer": false,
	}
}

// GNetServer impletion gnet.EventHandler
type GNetServer struct {
	acceptLoad atomic.Int64
	pkgLoad    atomic.Int64
	sequence   atomic.Int64
	opts       *ServerOptions
	procOpts   *process.ProcessOptions
	mux        sync.RWMutex
	clients    map[*GNetSession]Session
	udp        bool
	initNotify chan error
}

func NewServer(opts ...ServerOption) *GNetServer {
	s := &GNetServer{
		opts:    NewServerOptions(opts...),
		clients: make(map[*GNetSession]network.Session),
	}
	s.procOpts = process.NewProcessOptions(s.opts.ProcessOptions...)
	return s
}

func (svr *GNetServer) Run(addr string) (err error) {
	if addr == "" {
		addr = svr.opts.Addr
	} else {
		svr.opts.Addr = addr
	}
	if strings.HasPrefix(strings.ToLower(addr), "udp") {
		svr.udp = true
	}
	opts := convertServerOptions(svr.opts)
	return gnet.Run(svr, addr, opts...)
}

func (s *GNetServer) Broadcast(uri interface{}, msg interface{}, md metadata.MD) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return nil
	}

	ntf := s.procOpts.PacketPool.Get()
	defer s.procOpts.PacketPool.Put(ntf)
	err := s.procOpts.PacketWraper.NewPacket(ntf, packet.CmdNotify, uri, md)
	if err != nil {
		return err
	}
	err = s.procOpts.PacketWraper.PayloadMarshal(ntf, s.procOpts.MsgCodec, msg)
	if err != nil {
		return err
	}
	data, err := s.procOpts.PacketCodec.Marshal(ntf)
	if err != nil {
		return err
	}

	data = s.procOpts.PacketEncode.Encode(data)

	for cli := range s.clients {
		cli.Write(data)
	}
	return nil
}

func (s *GNetServer) BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, md metadata.MD) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return nil
	}

	ntf := s.procOpts.PacketPool.Get()
	defer s.procOpts.PacketPool.Put(ntf)
	err := s.procOpts.PacketWraper.NewPacket(ntf, packet.CmdNotify, uri, md)
	if err != nil {
		return err
	}
	err = s.procOpts.PacketWraper.PayloadMarshal(ntf, s.procOpts.MsgCodec, msg)
	if err != nil {
		return err
	}
	data, err := s.procOpts.PacketCodec.Marshal(ntf)
	if err != nil {
		return err
	}

	data = s.procOpts.PacketEncode.Encode(data)
	for cli := range s.clients {
		if filter(cli) {
			continue
		}
		cli.Write(data)
	}
	return nil
}

func (s *GNetServer) ForEach(f func(Session)) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return
	}
	for cli := range s.clients {
		f(cli)
	}
}

func (s *GNetServer) Shutdown(ctx context.Context) (err error) {
	gnet.Stop(ctx, s.opts.Addr)
	s.mux.Lock()
	defer s.mux.Unlock()
	for cli := range s.clients {
		cli.Close()
	}
	s.clients = nil
	return
}

// OnBoot fires when the engine is ready for accepting connections.
// The parameter engine has information and various utilities.
func (svr *GNetServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	if svr.initNotify != nil {
		svr.initNotify <- nil
	}
	return
}

// OnShutdown fires when the engine is being shut down, it is called right after
// all event-loops and connections are closed.

func (svr *GNetServer) OnShutdown(eng gnet.Engine) {
	if svr.initNotify != nil {
		svr.initNotify <- errors.New("init failed")
	}
}

// OnOpen fires when a new connection has been opened.
//
// The Conn c has information about the connection such as its local and remote addresses.
// The parameter out is the return value which is going to be sent back to the peer.
// Sending large amounts of data back to the peer in OnOpen is usually not recommended.
func (svr *GNetServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	svr.handleNewConn(c)
	return
}

// OnClose fires when a connection has been closed.
// The parameter err is the last known connection error.
func (svr *GNetServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	svr.opts.FrameLogger.New("gnetserver.OnClosed").Info("closed conn", zap.Error(err), zap.Any("errssss", err))
	src := c.Context()
	if src == nil {
		return
	}
	sess, ok := src.(*GNetSession)
	if !ok {
		return
	}
	// cleanup map
	svr.mux.Lock()
	delete(svr.clients, sess)
	svr.mux.Unlock()
	// cleanup
	svr.acceptLoad.Dec()
	sess.onClose()

	return
}

// OnTraffic fires when a socket receives data from the peer.
//
// Note that the []byte returned from Conn.Peek(int)/Conn.Next(int) is not allowed to be passed to a new goroutine,
// as this []byte will be reused within event-loop after OnTraffic() returns.
// If you have to use this []byte in a new goroutine, you should either make a copy of it or call Conn.Read([]byte)
// to read data into your own []byte, then pass the new []byte to the new goroutine.
func (svr *GNetServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	src := c.Context()
	if src == nil {
		svr.opts.FrameLogger.New("gnetserver.React").Info("react conn no context")
		action = gnet.Close
		return
	}
	sess, ok := src.(*GNetSession)
	if !ok {
		svr.opts.FrameLogger.New("gnetserver.React").Info("react conn context not session")
		action = gnet.Close
		return
	}
	for {
		if c.InboundBuffered() < 4 {
			break
		}
		buf, err := c.Peek(4)
		if err != nil {
			svr.opts.FrameLogger.New("gnetserver.React").Debug("read packet head", zap.Error(err))
			break
		}
		size := binary.BigEndian.Uint32(buf)
		buf, err = c.Peek(int(size) + 4)
		if err != nil {
			svr.opts.FrameLogger.New("gnetserver.React").Debug("read packet full", zap.Error(err))
			break
		}
		if svr.opts.ReuseReadBuffer {
			sess.OnRead(buf)
			c.Discard(int(size) + 4)
		} else {
			data := make([]byte, len(buf))
			copy(data, buf)
			sess.OnRead(data)
			c.Discard(int(size) + 4)
		}
	}
	return
}

// OnTick fires immediately after the engine starts and will fire again
// following the duration specified by the delay return value.
func (svr *GNetServer) OnTick() (delay time.Duration, action gnet.Action) {
	return
}

// serveWs handles websocket requests from the peer.
func (svr *GNetServer) handleNewConn(conn gnet.Conn) {
	log := svr.opts.FrameLogger.New("gnetserver.handleNewConn")
	// cleanup when exit
	cleanup := func() {
		svr.acceptLoad.Dec()
		err := conn.Close()
		if err != nil {
			log.Error("close session failed", zap.Error(err))
		}
	}
	// new session
	sess := &GNetSession{
		conn: conn,
		svr:  svr,
		RPCProcess: rpc.NewRPCProcess(
			process.NewInnerOptions(
				process.WithInnerOptionLoad(&svr.pkgLoad),
				process.WithInnerOptionSequence(&svr.sequence),
			),
			process.NewProcessOptions(
				svr.opts.ProcessOptions...,
			),
		),
		ctx:    context.Background(),
		cancel: func() {},
		udp:    svr.udp,
	}
	// sess.opts = svr.opts
	sess.Inner.ApplyOption(
		process.WithInnerOptionContextPool(GNETServerContextPool),
		process.WithInnerOptionOutput(sess),
	)
	// session count limit
	if svr.opts.AcceptLoadLimit(sess, svr.acceptLoad.Inc()) {
		log.Warn("session count limit")
		cleanup()
		return
	}
	// maybe cusotm session
	newSess, err := svr.opts.NewSession(sess)
	if err != nil {
		log.Error("new session failed", zap.Error(err))
		cleanup()
		return
	}
	// save context
	conn.SetContext(sess)

	// save map
	svr.mux.Lock()
	svr.clients[sess] = newSess
	svr.mux.Unlock()
	// config session context
	if svr.opts.StopImmediately {
		sess.ctx, sess.cancel = context.WithCancel(context.Background())
	}
	// apply config
	sess.Process.Inner.ApplyOption(
		process.WithInnerOptionOutput(newSess),
		process.WithInnerOptionBindData(newSess),
		process.WithInnerOptionRouter(svr.opts.SessionRouter(newSess, svr.opts.Router)),
		process.WithInnerOptionParentCtx(sess.ctx),
	)
	sess.Process.Opts.ApplyOption(
		process.WithLogger(svr.opts.SessionLogger(newSess, sess.Process.Opts.Logger)),
	)
	return
}

func convertServerOptions(opts *ServerOptions) (cfgs []gnet.Option) {
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
		//gnet.WithCodec(opts.Codec),
		gnet.WithLogLevel(zap.DebugLevel),
	)
	// TODO: gnet log options
	return
}
