package gotcp

import (
	"context"
	"encoding/binary"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/iface"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/zaplog"
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
		// Listen option. can replace kcp wrap
		"Listen": func(addr string) (ln net.Listener, err error) {
			return net.Listen("tcp", addr)
		},
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
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.Default),
		// SessionLogger custom session logger
		"SessionLogger": func(sess Session, global *zaplog.Logger) (r *zaplog.Logger) { return global },
		// NewSession custom session
		"NewSession": func(in Session) (Session, error) { return in, nil },
		// StopImmediately when session finish,business finish immediately.
		"StopImmediately": false,
		// ReadTimeout read timetou
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

// GoServer websocket server
type GoServer struct {
	acceptLoad atomic.Int64
	pkgLoad    atomic.Int64
	sequence   atomic.Int64
	opts       *ServerOptions
	mux        sync.RWMutex
	ln         net.Listener
	clients    map[Session]bool
	stop       chan struct{}
}

func NewServer(opts ...ServerOption) *GoServer {
	s := &GoServer{
		opts:    NewServerOptions(opts...),
		clients: make(map[Session]bool),
	}
	// check option limit
	if s.opts.MaxMessageSizeLimit > s.opts.ReadBufferSize {
		s.opts.ReadBufferSize = s.opts.MaxMessageSizeLimit
	}
	if s.opts.MaxMessageSizeLimit == 0 {
		s.opts.MaxMessageSizeLimit = s.opts.ReadBufferSize
	}
	// modify limit for write check
	s.opts.MaxMessageSizeLimit -= len(s.opts.PacketHeadBuf())
	return s
}

func (s *GoServer) Listen(addr string) (err error) {
	if addr == "" {
		addr = s.opts.Addr
	} else {
		s.opts.Addr = addr
	}
	s.ln, err = s.opts.Listen(addr)
	return
}

func (s *GoServer) Serve(ln net.Listener) (err error) {
	if ln != nil {
		s.ln = ln
	}
	return s.runAcceptLoop()
}

func (s *GoServer) Run(addr string) (err error) {
	if addr == "" {
		addr = s.opts.Addr
	} else {
		s.opts.Addr = addr
	}
	s.ln, err = s.opts.Listen(addr)
	if err != nil {
		return
	}
	return s.Serve(s.ln)
}

func (s *GoServer) runAcceptLoop() (err error) {
	var tempDelay time.Duration
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return io.EOF
			}
			return err
		}
		tempDelay = 0

		go s.accpetConn(conn)
	}
}

// serveWs handles websocket requests from the peer.
func (s *GoServer) accpetConn(conn net.Conn) {
	// cleanup when exit // cleanup :=
	defer func() {
		s.acceptLoad.Dec()
		err := conn.Close()
		if err != nil {
			s.opts.Logger.Error3("close session failed", zap.Error(err))
		}
	}()
	// new session
	sess := &GoSession{
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
		s.opts.Logger.Develop8("session count failed")
		// cleanup()
		return
	}
	// modify options
	s.opts.NetConnOption(conn)
	// maybe cusotm session
	newSess, err := s.opts.NewSession(sess)
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

func (s *GoServer) Broadcast(uri interface{}, msg interface{}, meta ...process.MetadataOption) error {
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

func (s *GoServer) BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, meta ...process.MetadataOption) error {
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

func (s *GoServer) ForEach(f func(Session)) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.clients) < 1 {
		return
	}
	for cli := range s.clients {
		f(cli)
	}
}

func (s *GoServer) Shutdown(ctx context.Context) (err error) {
	err = s.ln.Close()
	s.mux.Lock()
	defer s.mux.Unlock()
	for cli := range s.clients {
		cli.Close()
	}
	s.clients = nil
	return
}
