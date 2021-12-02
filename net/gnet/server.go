package gnet

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/iface"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/zaplog"
	"github.com/panjf2000/gnet"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// import type
type (
	Router         = process.Router
	Server         = iface.Server
	Session        = iface.Session
	SessionContext = iface.SessionContext
	Client         = iface.Client
	ClientContext  = iface.ClientContext
	WriteMethod    = iface.WriteMethod
)

// import const value
const (
	WriteAsync       = iface.WriteAsync
	WriteImmediately = iface.WriteImmediately
)

// ServerOption
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
		"Router": Router(nil),
		// SessionRouter custom session router
		"SessionRouter": func(sess Session, global Router) (r Router) { return global },
		// frame log
		"FrameLogger":(*zaplog.Logger)(zaplog.Frame),
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
		// WithCodec sets up a codec to handle TCP stream.
		"Codec": gnet.ICodec(gnet.ICodec(DefaultGNetCodec)),
	}
}

var DefaultGNetCodec = gnet.NewLengthFieldBasedFrameCodec(
	gnet.EncoderConfig{
		ByteOrder:                       binary.LittleEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	},
	gnet.DecoderConfig{
		ByteOrder:           binary.LittleEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	},
)

// GNetServer impletion gnet.EventHandler
type GNetServer struct {
	acceptLoad atomic.Int64
	pkgLoad    atomic.Int64
	sequence   atomic.Int64
	opts       *ServerOptions
	mux        sync.RWMutex
	clients    map[*GNetSession]Session
	gsvr       gnet.Server
	udp        bool
	initNotify chan error
}

func NewServer(opts ...ServerOption) *GNetServer {
	s := &GNetServer{
		opts:    NewServerOptions(opts...),
		clients: make(map[*GNetSession]iface.Session),
	}
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
	return gnet.Serve(svr, addr, opts...)
}

func (s *GNetServer) Broadcast(uri interface{}, msg interface{}, meta ...process.MetadataOption) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return nil
	}

	var buf []byte
	for cli := range s.clients {
		if buf == nil {
			ntf, err := cli.NewPacket(packet.Command_Oneway, uri, msg, meta)
			if err != nil {
				return err
			}
			buf, err = cli.MarshalPacket(ntf)
			if err != nil {
				return err
			}
		}
		cli.Write(buf)
	}
	return nil
}

func (s *GNetServer) BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, meta ...process.MetadataOption) error {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return nil
	}

	var buf []byte
	for cli := range s.clients {
		if filter(cli) {
			continue
		}
		if buf == nil {
			ntf, err := cli.NewPacket(packet.Command_Oneway, uri, msg, meta)
			if err != nil {
				return err
			}
			buf, err = cli.MarshalPacket(ntf)
			if err != nil {
				return err
			}
		}
		cli.Write(buf)
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

// OnInitComplete fires when the server is ready for accepting connections.
// The parameter:server has information and various utilities.
func (svr *GNetServer) OnInitComplete(s gnet.Server) (action gnet.Action) {
	svr.gsvr = s
	if svr.initNotify != nil {
		svr.initNotify <- nil
	}
	return
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (svr *GNetServer) OnShutdown(s gnet.Server) {
	svr.gsvr = s
	if svr.initNotify != nil {
		svr.initNotify <- errors.New("init failed")
	}
}

// OnOpened fires when a new connection has been opened.
// The parameter:c has information about the connection such as it's local and remote address.
// Parameter:out is the return value which is going to be sent back to the client.
func (svr *GNetServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	svr.handleNewConn(c)
	return
}

// OnClosed fires when a connection has been closed.
// The parameter:err is the last known connection error.
func (svr *GNetServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
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

// PreWrite fires just before any data is written to any client socket, this event function is usually used to
// put some code of logging/counting/reporting or any prepositive operations before writing data to client.
// func (svr *GNetServer) PreWrite(c gnet.Conn) { // 1.5.4+
func (svr *GNetServer) PreWrite() { // 1.5.3
}

// AfterWrite fires right after a packet is written to the peer socket, this event function is usually where
// you put the []byte's back to your memory pool.
func (svr *GNetServer) AfterWrite(c gnet.Conn, b []byte) {
}

// React fires when a connection sends the server data.
// Call c.Read() or c.ReadN(n) within the parameter:c to read incoming data from client.
// Parameter:out is the return value which is going to be sent back to the client.
func (svr *GNetServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
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
	sess.OnRead(frame)
	return
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (svr *GNetServer) Tick() (delay time.Duration, action gnet.Action) {
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
		Process: process.NewProcess(
			process.NewInnerOptions(
				process.WithInnerOptionsLoad(&svr.pkgLoad),
				process.WithInnerOptionsSequence(&svr.sequence),
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
	sess.Process.Inner.ApplyOption(
		process.WithInnerOptionsNewContext(sess.newContext),
		process.WithInnerOptionsOutput(sess),
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
		process.WithInnerOptionsOutput(newSess),
		process.WithInnerOptionsBindData(newSess),
		process.WithInnerOptionsRouter(svr.opts.SessionRouter(newSess, svr.opts.Router)),
		process.WithInnerOptionsParentCtx(sess.ctx),
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
		gnet.WithCodec(opts.Codec),
		gnet.WithLogLevel(zap.DebugLevel),
	)
	// TODO: gnet log options
	return
}
