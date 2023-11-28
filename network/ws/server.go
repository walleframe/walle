package ws

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aggronmagi/walle/network"
	"github.com/aggronmagi/walle/network/rpc"
	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/process/metadata"
	"github.com/aggronmagi/walle/process/packet"
	"github.com/aggronmagi/walle/zaplog"
	"github.com/gorilla/websocket"
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
		"Addr": string(":8080"),
		// WsPath websocket server path
		"WsPath": string("/ws"),
		// Upgrade websocket upgrade
		"Upgrade": (*websocket.Upgrader)(DefaultUpgrade),
		// SessoinFilter
		"UpgradeFail": func(w http.ResponseWriter, r *http.Request, reason error) {},
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
		"NewSession": func(in Session, r *http.Request) (Session, error) { return in, nil },
		// StopImmediately when session finish,business finish immediately.
		"StopImmediately": false,
		// ReadTimeout read timetou
		"ReadTimeout": time.Duration(0),
		// WriteTimeout write timeout
		"WriteTimeout": time.Duration(0),
		// MaxMessageLimit limit message size
		"MaxMessageLimit": int(0),
		// Write network data method.
		"WriteMethods": WriteMethod(WriteAsync),
		// SendQueueSize async send queue size
		"SendQueueSize": int(1024),
		// Heartbeat use websocket ping/pong.
		"Heartbeat": time.Duration(0),
		// HttpServeMux custom set mux
		"HttpServeMux": (*http.ServeMux)(http.DefaultServeMux),
	}
}

var DefaultUpgrade = &websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// WsServer websocket server
type WsServer struct {
	acceptLoad atomic.Int64
	pkgLoad    atomic.Int64
	sequence   atomic.Int64
	opts       *ServerOptions
	procInner  *process.InnerOptions
	procOpts   *process.ProcessOptions
	server     *http.Server
	mux        sync.RWMutex
	clients    map[*WsSession]Session
}

func NewServer(opts ...ServerOption) *WsServer {
	s := &WsServer{
		opts:    NewServerOptions(opts...),
		clients: make(map[*WsSession]network.Session),
		server:  &http.Server{},
	}
	s.server.Handler = s.opts.HttpServeMux
	// process opts
	s.procInner = process.NewInnerOptions(
		process.WithInnerOptionLoad(&s.pkgLoad),
		process.WithInnerOptionSequence(&s.sequence),
	)
	s.procOpts = process.NewProcessOptions(
		s.opts.ProcessOptions...,
	)
	return s
}

func (s *WsServer) Serve(ln net.Listener) (err error) {
	s.opts.HttpServeMux.HandleFunc(s.opts.WsPath, s.HttpServeWs)
	return s.server.Serve(ln)
}

func (s *WsServer) Run(addr string) (err error) {
	s.opts.HttpServeMux.HandleFunc(s.opts.WsPath, s.HttpServeWs)
	if addr == "" {
		s.server.Addr = s.opts.Addr
	} else {
		s.server.Addr = addr
	}
	return s.server.ListenAndServe()
}

// serveWs handles websocket requests from the peer.
func (s *WsServer) HttpServeWs(w http.ResponseWriter, r *http.Request) {
	log := s.opts.FrameLogger.New("ws.HttpServeWs").ResetTime()
	// upgrade websocket
	conn, err := DefaultUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade websocket failed", zap.Error(err))
		s.opts.UpgradeFail(w, r, err)
		return
	}
	// cleanup when exit // cleanup :=
	defer func() {
		s.acceptLoad.Dec()
		err := conn.Close()
		if err != nil {
			log.Error("close session failed", zap.Error(err))
		}
	}()
	// new session
	sess := &WsSession{
		conn: conn,
		svr:  s,
		RPCProcess: rpc.NewRPCProcess(
			process.NewInnerOptions(
				process.WithInnerOptionLoad(&s.pkgLoad),
				process.WithInnerOptionSequence(&s.sequence),
			),
			process.NewProcessOptions(
				s.opts.ProcessOptions...,
			),
		),
		ctx:    context.Background(),
		cancel: func() {},
		logger: s.opts.FrameLogger,
	}
	sess.opts = s.opts
	sess.Inner.ApplyOption(
		process.WithInnerOptionContextPool(GoServerContextPool),
		process.WithInnerOptionOutput(sess),
	)
	// session count limit
	if s.opts.AcceptLoadLimit(sess, s.acceptLoad.Inc()) {
		log.Warn("websocket session count limit", zap.Error(err))
		// cleanup()
		return
	}
	// maybe cusotm session
	newSess, err := s.opts.NewSession(sess, r)
	if err != nil {
		log.Error("new session failed", zap.Error(err))
		// cleanup()
		return
	}
	// save map
	s.mux.Lock()
	s.clients[sess] = newSess
	s.mux.Unlock()
	// config session context
	if s.opts.StopImmediately {
		sess.ctx, sess.cancel = context.WithCancel(context.Background())
	}
	// apply config
	sess.Process.Inner.ApplyOption(
		process.WithInnerOptionOutput(newSess),
		process.WithInnerOptionBindData(newSess),
		process.WithInnerOptionRouter(s.opts.SessionRouter(newSess, s.opts.Router)),
		process.WithInnerOptionParentCtx(sess.ctx),
	)
	sess.Process.Opts.ApplyOption(
		process.WithLogger(s.opts.SessionLogger(newSess, sess.Process.Opts.Logger)),
	)
	log.ClearTime()
	// cleanup map
	defer func() {
		s.mux.Lock()
		delete(s.clients, sess)
		s.mux.Unlock()
	}()
	// run client loop
	if nrun, ok := newSess.(interface {
		Run()
	}); ok {
		// wrap client session
		nrun.Run()
	} else {
		sess.Run()
	}
}

func (s *WsServer) Broadcast(uri interface{}, msg interface{}, md metadata.MD) error {
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

func (s *WsServer) BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, md metadata.MD) error {
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

func (s *WsServer) ForEach(f func(Session)) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return
	}
	for cli := range s.clients {
		f(cli)
	}
}
func (s *WsServer) Shutdown(ctx context.Context) (err error) {
	err = s.server.Shutdown(ctx)
	s.mux.Lock()
	defer s.mux.Unlock()
	for cli := range s.clients {
		cli.Close()
	}
	s.clients = nil
	return
}
