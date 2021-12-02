// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n ServerOption -o option.server.go"
// Version: 0.0.2

package ws

import (
	http "net/http"
	time "time"

	process "github.com/aggronmagi/walle/net/process"
	zaplog "github.com/aggronmagi/walle/zaplog"
	websocket "github.com/gorilla/websocket"
)

var _ = walleServer()

// ServerOption
type ServerOptions struct {
	// Addr Server Addr
	Addr string
	// WsPath websocket server path
	WsPath string
	// Upgrade websocket upgrade
	Upgrade (*websocket.Upgrader)
	// SessoinFilter
	UpgradeFail func(w http.ResponseWriter, r *http.Request, reason error)
	// accepted load limit
	AcceptLoadLimit func(sess Session, cnt int64) bool
	// Process Options
	ProcessOptions []process.ProcessOption
	// process router
	Router Router
	// SessionRouter custom session router
	SessionRouter func(sess Session, global Router) (r Router)
	// frame log
	FrameLogger (*zaplog.Logger)
	// SessionLogger custom session logger
	SessionLogger func(sess Session, global *zaplog.Logger) (r *zaplog.Logger)
	// NewSession custom session
	NewSession func(in Session, r *http.Request) (Session, error)
	// StopImmediately when session finish,business finish immediately.
	StopImmediately bool
	// ReadTimeout read timetou
	ReadTimeout time.Duration
	// WriteTimeout write timeout
	WriteTimeout time.Duration
	// MaxMessageLimit limit message size
	MaxMessageLimit int
	// Write network data method.
	WriteMethods WriteMethod
	// SendQueueSize async send queue size
	SendQueueSize int
	// Heartbeat use websocket ping/pong.
	Heartbeat time.Duration
	// HttpServeMux custom set mux
	HttpServeMux (*http.ServeMux)
}

// Addr Server Addr
func WithAddr(v string) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Addr
		cc.Addr = v
		return WithAddr(previous)
	}
}

// WsPath websocket server path
func WithWsPath(v string) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.WsPath
		cc.WsPath = v
		return WithWsPath(previous)
	}
}

// Upgrade websocket upgrade
func WithUpgrade(v *websocket.Upgrader) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Upgrade
		cc.Upgrade = v
		return WithUpgrade(previous)
	}
}

// SessoinFilter
func WithUpgradeFail(v func(w http.ResponseWriter, r *http.Request, reason error)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.UpgradeFail
		cc.UpgradeFail = v
		return WithUpgradeFail(previous)
	}
}

// accepted load limit
func WithAcceptLoadLimit(v func(sess Session, cnt int64) bool) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.AcceptLoadLimit
		cc.AcceptLoadLimit = v
		return WithAcceptLoadLimit(previous)
	}
}

// Process Options
func WithProcessOptions(v ...process.ProcessOption) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.ProcessOptions
		cc.ProcessOptions = v
		return WithProcessOptions(previous...)
	}
}

// process router
func WithRouter(v Router) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Router
		cc.Router = v
		return WithRouter(previous)
	}
}

// SessionRouter custom session router
func WithSessionRouter(v func(sess Session, global Router) (r Router)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.SessionRouter
		cc.SessionRouter = v
		return WithSessionRouter(previous)
	}
}

// frame log
func WithFrameLogger(v *zaplog.Logger) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.FrameLogger
		cc.FrameLogger = v
		return WithFrameLogger(previous)
	}
}

// SessionLogger custom session logger
func WithSessionLogger(v func(sess Session, global *zaplog.Logger) (r *zaplog.Logger)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.SessionLogger
		cc.SessionLogger = v
		return WithSessionLogger(previous)
	}
}

// NewSession custom session
func WithNewSession(v func(in Session, r *http.Request) (Session, error)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.NewSession
		cc.NewSession = v
		return WithNewSession(previous)
	}
}

// StopImmediately when session finish,business finish immediately.
func WithStopImmediately(v bool) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.StopImmediately
		cc.StopImmediately = v
		return WithStopImmediately(previous)
	}
}

// ReadTimeout read timetou
func WithReadTimeout(v time.Duration) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.ReadTimeout
		cc.ReadTimeout = v
		return WithReadTimeout(previous)
	}
}

// WriteTimeout write timeout
func WithWriteTimeout(v time.Duration) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.WriteTimeout
		cc.WriteTimeout = v
		return WithWriteTimeout(previous)
	}
}

// MaxMessageLimit limit message size
func WithMaxMessageLimit(v int) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.MaxMessageLimit
		cc.MaxMessageLimit = v
		return WithMaxMessageLimit(previous)
	}
}

// Write network data method.
func WithWriteMethods(v WriteMethod) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.WriteMethods
		cc.WriteMethods = v
		return WithWriteMethods(previous)
	}
}

// SendQueueSize async send queue size
func WithSendQueueSize(v int) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.SendQueueSize
		cc.SendQueueSize = v
		return WithSendQueueSize(previous)
	}
}

// Heartbeat use websocket ping/pong.
func WithHeartbeat(v time.Duration) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Heartbeat
		cc.Heartbeat = v
		return WithHeartbeat(previous)
	}
}

// HttpServeMux custom set mux
func WithHttpServeMux(v *http.ServeMux) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.HttpServeMux
		cc.HttpServeMux = v
		return WithHttpServeMux(previous)
	}
}

// SetOption modify options
func (cc *ServerOptions) SetOption(opt ServerOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *ServerOptions) ApplyOption(opts ...ServerOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *ServerOptions) GetSetOption(opt ServerOption) ServerOption {
	return opt(cc)
}

// ServerOption option define
type ServerOption func(cc *ServerOptions) ServerOption

// NewServerOptions create options instance.
func NewServerOptions(opts ...ServerOption) *ServerOptions {
	cc := newDefaultServerOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogServerOptions != nil {
		watchDogServerOptions(cc)
	}
	return cc
}

// InstallServerOptionsWatchDog install watch dog
func InstallServerOptionsWatchDog(dog func(cc *ServerOptions)) {
	watchDogServerOptions = dog
}

var watchDogServerOptions func(cc *ServerOptions)

// newDefaultServerOptions new option with default value
func newDefaultServerOptions() *ServerOptions {
	cc := &ServerOptions{
		Addr:    ":8080",
		WsPath:  "/ws",
		Upgrade: DefaultUpgrade,
		UpgradeFail: func(w http.ResponseWriter, r *http.Request, reason error) {
		},
		AcceptLoadLimit: func(sess Session, cnt int64) bool {
			return false
		},
		ProcessOptions: nil,
		Router:         nil,
		SessionRouter: func(sess Session, global Router) (r Router) {
			return global
		},
		FrameLogger: zaplog.Frame,
		SessionLogger: func(sess Session, global *zaplog.Logger) (r *zaplog.Logger) {
			return global
		},
		NewSession: func(in Session, r *http.Request) (Session, error) {
			return in, nil
		},
		StopImmediately: false,
		ReadTimeout:     0,
		WriteTimeout:    0,
		MaxMessageLimit: 0,
		WriteMethods:    WriteAsync,
		SendQueueSize:   1024,
		Heartbeat:       0,
		HttpServeMux:    http.DefaultServeMux,
	}
	return cc
}
