// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n ClientOption -f Client -o option.client.go"
// Version: 0.0.2

package gotcp

import (
	binary "encoding/binary"
	math "math"
	net "net"
	time "time"

	process "github.com/walleframe/walle/process"
	errcode "github.com/walleframe/walle/process/errcode"
	zaplog "github.com/walleframe/walle/zaplog"
)

var _ = walleClient()

// ClientOption
type ClientOptions struct {
	// Network tcp/tcp4/tcp6/unix
	Network string
	// Addr Server Addr
	Addr string
	// Dialer config net dialer
	Dialer func(network, addr string) (conn net.Conn, err error)
	// Process Options
	ProcessOptions []process.ProcessOption
	// process router
	Router Router
	// frame log
	FrameLogger (*zaplog.Logger)
	// AutoReconnect auto reconnect server. zero means not reconnect! -1 means always reconnect, >0 : reconnect times
	AutoReconnectTime int
	// AutoReconnectWait reconnect wait time
	AutoReconnectWait time.Duration
	// StopImmediately when session finish,business finish immediately.
	StopImmediately bool
	// ReadTimeout read timeout
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
	ReuseReadBuffer bool
	// MaxMessageSizeLimit limit message size
	MaxMessageSizeLimit int
	// BlockConnect 创建客户端时候，是否阻塞等待链接服务器
	BlockConnect bool
}

// Network tcp/tcp4/tcp6/unix
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

// Dialer config net dialer
func WithClientOptionsDialer(v func(network, addr string) (conn net.Conn, err error)) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Dialer
		cc.Dialer = v
		return WithClientOptionsDialer(previous)
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

// AutoReconnect auto reconnect server. zero means not reconnect! -1 means always reconnect, >0 : reconnect times
func WithClientOptionsAutoReconnectTime(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.AutoReconnectTime
		cc.AutoReconnectTime = v
		return WithClientOptionsAutoReconnectTime(previous)
	}
}

// AutoReconnectWait reconnect wait time
func WithClientOptionsAutoReconnectWait(v time.Duration) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.AutoReconnectWait
		cc.AutoReconnectWait = v
		return WithClientOptionsAutoReconnectWait(previous)
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

// ReadTimeout read timeout
func WithClientOptionsReadTimeout(v time.Duration) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ReadTimeout
		cc.ReadTimeout = v
		return WithClientOptionsReadTimeout(previous)
	}
}

// WriteTimeout write timeout
func WithClientOptionsWriteTimeout(v time.Duration) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.WriteTimeout
		cc.WriteTimeout = v
		return WithClientOptionsWriteTimeout(previous)
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

// SendQueueSize async send queue size
func WithClientOptionsSendQueueSize(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.SendQueueSize
		cc.SendQueueSize = v
		return WithClientOptionsSendQueueSize(previous)
	}
}

// Heartbeat use websocket ping/pong.
func WithClientOptionsHeartbeat(v time.Duration) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.Heartbeat
		cc.Heartbeat = v
		return WithClientOptionsHeartbeat(previous)
	}
}

// tcp packet head
func WithClientOptionsPacketHeadBuf(v func() []byte) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.PacketHeadBuf
		cc.PacketHeadBuf = v
		return WithClientOptionsPacketHeadBuf(previous)
	}
}

// read tcp packet head size
func WithClientOptionsReadSize(v func(head []byte) (size int)) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ReadSize
		cc.ReadSize = v
		return WithClientOptionsReadSize(previous)
	}
}

// write tcp packet head size
func WithClientOptionsWriteSize(v func(head []byte, size int) (err error)) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.WriteSize
		cc.WriteSize = v
		return WithClientOptionsWriteSize(previous)
	}
}

// ReadBufferSize 一定要大于最大消息的大小.每个链接一个缓冲区。
func WithClientOptionsReadBufferSize(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.ReadBufferSize
		cc.ReadBufferSize = v
		return WithClientOptionsReadBufferSize(previous)
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

// MaxMessageSizeLimit limit message size
func WithClientOptionsMaxMessageSizeLimit(v int) ClientOption {
	return func(cc *ClientOptions) ClientOption {
		previous := cc.MaxMessageSizeLimit
		cc.MaxMessageSizeLimit = v
		return WithClientOptionsMaxMessageSizeLimit(previous)
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
		Network: "tcp",
		Addr:    "localhost:8080",
		Dialer: func(network, addr string) (conn net.Conn, err error) {
			return net.Dial(network, addr)
		},
		ProcessOptions:    nil,
		Router:            nil,
		FrameLogger:       zaplog.GetFrameLogger(),
		AutoReconnectTime: -1,
		AutoReconnectWait: time.Millisecond * 500,
		StopImmediately:   false,
		ReadTimeout:       0,
		WriteTimeout:      0,
		WriteMethods:      WriteAsync,
		SendQueueSize:     1024,
		Heartbeat:         0,
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
		ReuseReadBuffer:     true,
		MaxMessageSizeLimit: 0,
		BlockConnect:        true,
	}
	return cc
}
