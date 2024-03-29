// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n CallOption -f Call -o option.call.go"
// Version: 0.0.4

package rpc

import (
	"time"

	"github.com/walleframe/walle/process/metadata"
)

var _ = walleCallOption()

// CallOption rpc call options
type CallOptions struct {
	// rpc call timeout
	Timeout time.Duration
	// metadata
	Metadata metadata.MD
}

// rpc call timeout
func WithCallOptionTimeout(v time.Duration) CallOption {
	return func(cc *CallOptions) CallOption {
		previous := cc.Timeout
		cc.Timeout = v
		return WithCallOptionTimeout(previous)
	}
}

// metadata
func WithCallOptionMetadata(v metadata.MD) CallOption {
	return func(cc *CallOptions) CallOption {
		previous := cc.Metadata
		cc.Metadata = v
		return WithCallOptionMetadata(previous)
	}
}

// SetOption modify options
func (cc *CallOptions) SetOption(opt CallOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *CallOptions) ApplyOption(opts ...CallOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *CallOptions) GetSetOption(opt CallOption) CallOption {
	return opt(cc)
}

// CallOption option define
type CallOption func(cc *CallOptions) CallOption

// NewCallOptions create options instance.
func NewCallOptions(opts ...CallOption) *CallOptions {
	cc := newDefaultCallOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogCallOptions != nil {
		watchDogCallOptions(cc)
	}
	return cc
}

// InstallCallOptionsWatchDog install watch dog
func InstallCallOptionsWatchDog(dog func(cc *CallOptions)) {
	watchDogCallOptions = dog
}

var watchDogCallOptions func(cc *CallOptions)

// newDefaultCallOptions new option with default value
func newDefaultCallOptions() *CallOptions {
	cc := &CallOptions{
		Timeout:  0,
		Metadata: nil,
	}
	return cc
}
