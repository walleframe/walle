// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n ClientOption -f Client -o option.client.go"
// Version: 0.0.2

package gnet

import (
	time "time"

	process "github.com/walleframe/walle/process"
	zaplog "github.com/walleframe/walle/zaplog"
	gnet "github.com/panjf2000/gnet/v2"
)

var _ = walleClient()

// ClientOption
type ClientOptions struct {
	Network string
	// Addr Server Addr
	Addr string
	// Process Options
	ProcessOptions []process.ProcessOption
	// process router
	Router Router
	// frame log
	FrameLogger (*zaplog.Logger)
	// AutoReconnect auto reconnect server. zero means not reconnect!
	AutoReconnectTime int
	// StopImmediately when session finish,business finish immediately.
	StopImmediately bool
	// WithMulticore sets up multi-cores in gnet server.
	Multicore bool
	// WithLockOSThread sets up LockOSThread mode for I/O event-loops.
	LockOSThread bool
	// WithLoadBalancing sets up the load-balancing algorithm in gnet server.
	LoadBalancing gnet.LoadBalancing
	// WithNumEventLoop sets up NumEventLoop in gnet server.
	NumEventLoop int
	// WithReusePort sets up SO_REUSEPORT socket option.
	ReusePort bool
	// WithTCPKeepAlive sets up the SO_KEEPALIVE socket option with duration.
	TCPKeepAlive time.Duration
	// WithTCPNoDelay enable/disable the TCP_NODELAY socket option.
	TCPNoDelay gnet.TCPSocketOpt
	// WithReadBufferCap sets up ReadBufferCap for reading bytes.
	ReadBufferCap int
	// WithSocketRecvBuffer sets the maximum socket receive buffer in bytes.
	SocketRecvBuffer int
	// WithSocketSendBuffer sets the maximum socket send buffer in bytes.
	SocketSendBuffer int
	// WithTicker indicates that a ticker is set.
	Ticker time.Duration
	// BlockConnect 创建客户端时候，是否阻塞等待链接服务器
	BlockConnect bool
	// Write network data method.
	WriteMethods WriteMethod
	// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
	// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
	// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
	ReuseReadBuffer bool
}

func WithClientOptionsNetwork(v string) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Network
		cc.Network = v
		return WithClientOptionsNetwork(previous)
	}
}

// Addr Server Addr
func WithClientOptionsAddr(v string) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Addr
		cc.Addr = v
		return WithClientOptionsAddr(previous)
	}
}

// Process Options
func WithClientOptionsProcessOptions(v ...process.ProcessOption) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ProcessOptions
		cc.ProcessOptions = v
		return WithClientOptionsProcessOptions(previous...)
	}
}

// process router
func WithClientOptionsRouter(v Router) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Router
		cc.Router = v
		return WithClientOptionsRouter(previous)
	}
}

// frame log
func WithClientOptionsFrameLogger(v *zaplog.Logger) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.FrameLogger
		cc.FrameLogger = v
		return WithClientOptionsFrameLogger(previous)
	}
}

// AutoReconnect auto reconnect server. zero means not reconnect!
func WithClientOptionsAutoReconnectTime(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.AutoReconnectTime
		cc.AutoReconnectTime = v
		return WithClientOptionsAutoReconnectTime(previous)
	}
}

// StopImmediately when session finish,business finish immediately.
func WithClientOptionsStopImmediately(v bool) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.StopImmediately
		cc.StopImmediately = v
		return WithClientOptionsStopImmediately(previous)
	}
}

// WithMulticore sets up multi-cores in gnet server.
func WithClientOptionsMulticore(v bool) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Multicore
		cc.Multicore = v
		return WithClientOptionsMulticore(previous)
	}
}

// WithLockOSThread sets up LockOSThread mode for I/O event-loops.
func WithClientOptionsLockOSThread(v bool) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.LockOSThread
		cc.LockOSThread = v
		return WithClientOptionsLockOSThread(previous)
	}
}

// WithLoadBalancing sets up the load-balancing algorithm in gnet server.
func WithClientOptionsLoadBalancing(v gnet.LoadBalancing) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.LoadBalancing
		cc.LoadBalancing = v
		return WithClientOptionsLoadBalancing(previous)
	}
}

// WithNumEventLoop sets up NumEventLoop in gnet server.
func WithClientOptionsNumEventLoop(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.NumEventLoop
		cc.NumEventLoop = v
		return WithClientOptionsNumEventLoop(previous)
	}
}

// WithReusePort sets up SO_REUSEPORT socket option.
func WithClientOptionsReusePort(v bool) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ReusePort
		cc.ReusePort = v
		return WithClientOptionsReusePort(previous)
	}
}

// WithTCPKeepAlive sets up the SO_KEEPALIVE socket option with duration.
func WithClientOptionsTCPKeepAlive(v time.Duration) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.TCPKeepAlive
		cc.TCPKeepAlive = v
		return WithClientOptionsTCPKeepAlive(previous)
	}
}

// WithTCPNoDelay enable/disable the TCP_NODELAY socket option.
func WithClientOptionsTCPNoDelay(v gnet.TCPSocketOpt) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.TCPNoDelay
		cc.TCPNoDelay = v
		return WithClientOptionsTCPNoDelay(previous)
	}
}

// WithReadBufferCap sets up ReadBufferCap for reading bytes.
func WithClientOptionsReadBufferCap(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ReadBufferCap
		cc.ReadBufferCap = v
		return WithClientOptionsReadBufferCap(previous)
	}
}

// WithSocketRecvBuffer sets the maximum socket receive buffer in bytes.
func WithClientOptionsSocketRecvBuffer(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.SocketRecvBuffer
		cc.SocketRecvBuffer = v
		return WithClientOptionsSocketRecvBuffer(previous)
	}
}

// WithSocketSendBuffer sets the maximum socket send buffer in bytes.
func WithClientOptionsSocketSendBuffer(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.SocketSendBuffer
		cc.SocketSendBuffer = v
		return WithClientOptionsSocketSendBuffer(previous)
	}
}

// WithTicker indicates that a ticker is set.
func WithClientOptionsTicker(v time.Duration) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Ticker
		cc.Ticker = v
		return WithClientOptionsTicker(previous)
	}
}

// BlockConnect 创建客户端时候，是否阻塞等待链接服务器
func WithClientOptionsBlockConnect(v bool) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.BlockConnect
		cc.BlockConnect = v
		return WithClientOptionsBlockConnect(previous)
	}
}

// Write network data method.
func WithClientOptionsWriteMethods(v WriteMethod) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.WriteMethods
		cc.WriteMethods = v
		return WithClientOptionsWriteMethods(previous)
	}
}

// ReuseReadBuffer 复用read缓存区。影响Process.DispatchFilter.
// 如果此选项设置为true，在DispatchFilter内如果开启协程，需要手动复制内存。
// 如果在DispatchFilter内不开启协程，设置为true可以减少内存分配。
func WithClientOptionsReuseReadBuffer(v bool) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ReuseReadBuffer
		cc.ReuseReadBuffer = v
		return WithClientOptionsReuseReadBuffer(previous)
	}
}

// SetOption modify options
func (cc *ClientOptions) SetOption(opt ClientOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *ClientOptions) ApplyOption(opts ...ClientOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *ClientOptions) GetSetOption(opt ClientOption) ClientOption {
	return opt(cc)
}

// ClientOption option define
type ClientOption func(cc *ClientOptions) ClientOption

// NewClientOptions create options instance.
func NewClientOptions(opts ...ClientOption) *ClientOptions {
	cc := newDefaultClientOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogClientOptions != nil {
		watchDogClientOptions(cc)
	}
	return cc
}

// InstallClientOptionsWatchDog install watch dog
func InstallClientOptionsWatchDog(dog func(cc *ClientOptions)) {
	watchDogClientOptions = dog
}

var watchDogClientOptions func(cc *ClientOptions)

// newDefaultClientOptions new option with default value
func newDefaultClientOptions() *ClientOptions {
	cc := &ClientOptions{
		Network:           "tcp",
		Addr:              "localhost:8080",
		ProcessOptions:    nil,
		Router:            nil,
		FrameLogger:       zaplog.GetFrameLogger(),
		AutoReconnectTime: 5,
		StopImmediately:   false,
		Multicore:         false,
		LockOSThread:      false,
		LoadBalancing:     gnet.SourceAddrHash,
		NumEventLoop:      0,
		ReusePort:         false,
		TCPKeepAlive:      0,
		TCPNoDelay:        gnet.TCPNoDelay,
		ReadBufferCap:     0,
		SocketRecvBuffer:  0,
		SocketSendBuffer:  0,
		Ticker:            0,
		BlockConnect:      true,
		WriteMethods:      WriteAsync,
		ReuseReadBuffer:   true,
	}
	return cc
}
