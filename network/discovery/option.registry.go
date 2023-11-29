// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n RegistryOption -f Registry -o option.registry.go"
// Version: 0.0.2

package discovery

import (
	net "net"

	util "github.com/walleframe/walle/util"
	zaplog "github.com/walleframe/walle/zaplog"
	uuid "github.com/google/uuid"
)

var _ = walleRegistry()

type RegistryOptions struct {
	// NewEntry create custom entry for registry
	NewEntry func(addr net.Addr) (_ Entry, err error)
	// Codec use for encode entry.
	Codec EntryCodec
	// frame log
	FrameLogger (*zaplog.Logger)
}

// NewEntry create custom entry for registry
func WithRegistryOptionsNewEntry(v func(addr net.Addr) (_ Entry, err error)) RegistryOption {
	return func(cc *RegistryOptions) RegistryOption {
		previous := cc.NewEntry
		cc.NewEntry = v
		return WithRegistryOptionsNewEntry(previous)
	}
}

// Codec use for encode entry.
func WithRegistryOptionsCodec(v EntryCodec) RegistryOption {
	return func(cc *RegistryOptions) RegistryOption {
		previous := cc.Codec
		cc.Codec = v
		return WithRegistryOptionsCodec(previous)
	}
}

// frame log
func WithRegistryOptionsFrameLogger(v *zaplog.Logger) RegistryOption {
	return func(cc *RegistryOptions) RegistryOption {
		previous := cc.FrameLogger
		cc.FrameLogger = v
		return WithRegistryOptionsFrameLogger(previous)
	}
}

// SetOption modify options
func (cc *RegistryOptions) SetOption(opt RegistryOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *RegistryOptions) ApplyOption(opts ...RegistryOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *RegistryOptions) GetSetOption(opt RegistryOption) RegistryOption {
	return opt(cc)
}

// RegistryOption option define
type RegistryOption func(cc *RegistryOptions) RegistryOption

// NewRegistryOptions create options instance.
func NewRegistryOptions(opts ...RegistryOption) *RegistryOptions {
	cc := newDefaultRegistryOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogRegistryOptions != nil {
		watchDogRegistryOptions(cc)
	}
	return cc
}

// InstallRegistryOptionsWatchDog install watch dog
func InstallRegistryOptionsWatchDog(dog func(cc *RegistryOptions)) {
	watchDogRegistryOptions = dog
}

var watchDogRegistryOptions func(cc *RegistryOptions)

// newDefaultRegistryOptions new option with default value
func newDefaultRegistryOptions() *RegistryOptions {
	cc := &RegistryOptions{
		NewEntry: func(addr net.Addr) (_ Entry, err error) {

			ip, port, err := net.SplitHostPort(addr.String())
			if err != nil {
				return
			}

			if ip == "::" {
				ip = util.GetLocalIP()
			}

			return &Node{
				Identifier: uuid.New().String(),
				Network:    addr.Network(),
				Addr:       net.JoinHostPort(ip, port),
				Balance:    "rr",
				Status:     int(EntryStateOffline),
				MD:         map[string]string{},
			}, nil
		},
		Codec:       NodeJsonEntryCodec,
		FrameLogger: zaplog.GetFrameLogger(),
	}
	return cc
}
