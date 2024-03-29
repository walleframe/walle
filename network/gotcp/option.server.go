// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n ServerOption -o option.server.go"
// Version: 0.0.4

package gotcp

import (
	"encoding/binary"
	"math"
	"net"
	"time"

	"github.com/walleframe/walle/network/discovery"
	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/process/errcode"
	"github.com/walleframe/walle/zaplog"
)

var _ = walleServer()

// ServerOption
type ServerOptions struct {
	// Addr Server Addr
	Addr string
	// Listen option. can replace kcp wrap
	Listen func(addr string) (ln net.Listener, err error)
	// NetOption modify raw options
	NetConnOption func(net.Conn)
	// accepted load limit
	AcceptLoadLimit func(sess Session, cnt int64) bool
	// Process Options
	ProcessOptions []process.ProcessOption
	// process router
	Router Router
	// SessionRouter custom session router
	SessionRouter func(sess Session, global Router) (r Router)
	// frame log
	FrameLogger *zaplog.Logger
	// SessionLogger custom session logger
	SessionLogger func(sess Session, global *zaplog.Logger) (r *zaplog.Logger)
	// NewSession custom session
	NewSession func(in Session) (Session, error)
	// StopImmediately when session finish,business finish immediately.
	StopImmediately bool
	// ReadTimeout read timetou
	ReadTimeout time.Duration
	// WriteTimeout write timeout
	WriteTimeout time.Duration
	// Write network data method.
	WriteMethods WriteMethod
	// SendQueueSize async send queue size
	SendQueueSize int
	// Heartbeat use websocket ping/pong.
	Heartbeat time.Duration
	// tcp packet head
	PacketHeadBuf func() []byte
	// read tcp packet head size
	ReadSize func(head []byte) (size int)
	// write tcp packet head size
	WriteSize func(head []byte, size int) (err error)
	// ReadBufferSize 一定要大于最大消息的大小.每个链接一个缓冲区。
	ReadBufferSize int
	// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
	// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
	// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
	// 默认为false,是为了防止错误的配置导致bug。
	ReuseReadBuffer bool
	// MaxMessageSizeLimit limit message size
	MaxMessageSizeLimit int
	// Registry
	Registry discovery.Registry
}

// Addr Server Addr
func WithAddr(v string) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Addr
		cc.Addr = v
		return WithAddr(previous)
	}
}

// Listen option. can replace kcp wrap
func WithListen(v func(addr string) (ln net.Listener, err error)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Listen
		cc.Listen = v
		return WithListen(previous)
	}
}

// NetOption modify raw options
func WithNetConnOption(v func(net.Conn)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.NetConnOption
		cc.NetConnOption = v
		return WithNetConnOption(previous)
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
func WithNewSession(v func(in Session) (Session, error)) ServerOption {
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

// tcp packet head
func WithPacketHeadBuf(v func() []byte) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.PacketHeadBuf
		cc.PacketHeadBuf = v
		return WithPacketHeadBuf(previous)
	}
}

// read tcp packet head size
func WithReadSize(v func(head []byte) (size int)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.ReadSize
		cc.ReadSize = v
		return WithReadSize(previous)
	}
}

// write tcp packet head size
func WithWriteSize(v func(head []byte, size int) (err error)) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.WriteSize
		cc.WriteSize = v
		return WithWriteSize(previous)
	}
}

// ReadBufferSize 一定要大于最大消息的大小.每个链接一个缓冲区。
func WithReadBufferSize(v int) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.ReadBufferSize
		cc.ReadBufferSize = v
		return WithReadBufferSize(previous)
	}
}

// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
// 默认为false,是为了防止错误的配置导致bug。
func WithReuseReadBuffer(v bool) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.ReuseReadBuffer
		cc.ReuseReadBuffer = v
		return WithReuseReadBuffer(previous)
	}
}

// MaxMessageSizeLimit limit message size
func WithMaxMessageSizeLimit(v int) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.MaxMessageSizeLimit
		cc.MaxMessageSizeLimit = v
		return WithMaxMessageSizeLimit(previous)
	}
}

// Registry
func WithRegistry(v discovery.Registry) ServerOption {
	return func(cc *ServerOptions) ServerOption {
		previous := cc.Registry
		cc.Registry = v
		return WithRegistry(previous)
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
		Addr: ":8080",
		Listen: func(addr string) (ln net.Listener, err error) {
			return net.Listen("tcp", addr)
		},
		NetConnOption: func(net.Conn) {
		},
		AcceptLoadLimit: func(sess Session, cnt int64) bool {
			return false
		},
		ProcessOptions: nil,
		Router:         process.GetRouter(),
		SessionRouter: func(sess Session, global Router) (r Router) {
			return global
		},
		FrameLogger: zaplog.GetFrameLogger(),
		SessionLogger: func(sess Session, global *zaplog.Logger) (r *zaplog.Logger) {
			return global
		},
		NewSession: func(in Session) (Session, error) {
			return in, nil
		},
		StopImmediately: false,
		ReadTimeout:     0,
		WriteTimeout:    0,
		WriteMethods:    WriteAsync,
		SendQueueSize:   1024,
		Heartbeat:       0,
		PacketHeadBuf: func() []byte {
			return make([]byte, 4)
		},
		ReadSize: func(head []byte) (size int) {
			size = int(binary.LittleEndian.Uint32(head))
			return
		},
		WriteSize: func(head []byte, size int) (err error) {
			if size >= math.MaxUint32 {
				return errcode.ErrPacketsizeInvalid
			}
			binary.LittleEndian.PutUint32(head, uint32(size))
			return
		},
		ReadBufferSize:      65535,
		ReuseReadBuffer:     false,
		MaxMessageSizeLimit: 0,
		Registry:            discovery.NoOpRegistry{},
	}
	return cc
}
