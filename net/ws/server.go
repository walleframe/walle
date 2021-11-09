package ws

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/iface"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/zaplog"
	"github.com/gorilla/websocket"
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
		"Router": Router(nil),
		// SessionRouter custom session router
		"SessionRouter": func(sess Session, global Router) (r Router) { return global },
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.Default),
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
	server     *http.Server
	mux        sync.RWMutex
	clients    map[Session]bool
}

func NewServer(opts ...ServerOption) *WsServer {
	s := &WsServer{
		opts:    NewServerOptions(opts...),
		clients: make(map[Session]bool),
		server:  &http.Server{},
	}
	return s
}

func (s *WsServer) Serve(ln net.Listener) (err error) {
	http.HandleFunc(s.opts.WsPath, s.HttpServeWs)
	return s.server.Serve(ln)
}

func (s *WsServer) Run(addr string) (err error) {
	http.HandleFunc(s.opts.WsPath, s.HttpServeWs)
	if addr == "" {
		s.server.Addr = s.opts.Addr
	} else {
		s.server.Addr = addr
	}
	return s.server.ListenAndServe()
}

// serveWs handles websocket requests from the peer.
func (s *WsServer) HttpServeWs(w http.ResponseWriter, r *http.Request) {
	// upgrade websocket
	conn, err := DefaultUpgrade.Upgrade(w, r, nil)
	if err != nil {
		s.opts.Logger.Error3("upgrade websocket failed", zap.Error(err))
		s.opts.UpgradeFail(w, r, err)
		return
	}
	// cleanup when exit // cleanup :=
	defer func() {
		s.acceptLoad.Dec()
		err := conn.Close()
		if err != nil {
			s.opts.Logger.Error3("close session failed", zap.Error(err))
		}
	}()
	// new session
	sess := &WsSession{
		conn: conn,
		svr:  s,
		Process: process.NewProcess(
			process.NewInnerOptions(
				process.WithInnerOptionsLoad(&s.pkgLoad),
				process.WithInnerOptionsSequence(&s.sequence),
			),
			process.NewProcessOptions(
				s.opts.ProcessOptions...,
			),
		),
		ctx:    context.Background(),
		cancel: func() {},
	}
	sess.opts = s.opts
	sess.Process.Inner.ApplyOption(
		process.WithInnerOptionsNewContext(sess.newContext),
		process.WithInnerOptionsOutput(sess),
	)
	// session count limit
	if s.opts.AcceptLoadLimit(sess, s.acceptLoad.Inc()) {
		s.opts.Logger.Develop8("websocket session count failed", zap.Error(err))
		// cleanup()
		return
	}
	// maybe cusotm session
	newSess, err := s.opts.NewSession(sess, r)
	if err != nil {
		s.opts.Logger.Error3("new session failed", zap.Error(err))
		// cleanup()
		return
	}
	// save map
	s.mux.Lock()
	s.clients[newSess] = true
	s.mux.Unlock()
	// config session context
	if s.opts.StopImmediately {
		sess.ctx, sess.cancel = context.WithCancel(context.Background())
	}
	// apply config
	sess.Process.Inner.ApplyOption(
		process.WithInnerOptionsOutput(newSess),
		process.WithInnerOptionsBindData(newSess),
		process.WithInnerOptionsRouter(s.opts.SessionRouter(newSess, s.opts.Router)),
		process.WithInnerOptionsParentCtx(sess.ctx),
	)
	sess.Process.Opts.ApplyOption(
		process.WithLogger(s.opts.SessionLogger(newSess, sess.Process.Opts.Logger)),
	)
	// cleanup map
	defer func() {
		s.mux.Lock()
		delete(s.clients, newSess)
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

func (s *WsServer) Broadcast(uri interface{}, msg interface{}, meta ...process.MetadataOption) error {
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

func (s *WsServer) BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, meta ...process.MetadataOption) error {
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

func (s *WsServer) Shutdown() (err error) {
	err = s.server.Close()
	s.mux.Lock()
	defer s.mux.Unlock()
	for cli := range s.clients {
		cli.Close()
	}
	s.clients = nil
	return
}
