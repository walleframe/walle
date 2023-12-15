// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n Option -o option.go"
// Version: 0.0.3

package etcdv3

import (
	context "context"
	time "time"

	clientv3 "go.etcd.io/etcd/client/v3"
	zap "go.uber.org/zap"
)

var _ = walleOptions()

// Option config etcdv3 store
type Options struct {
	// etcd endpoints
	Endpoints []string
	UserName  string
	Password  string
	// context
	Context context.Context
	// dial etcd timeout config
	DialTimeout time.Duration
	// CustomSet custom set etcd options
	CustomSet func(cfg *clientv3.Config)
	// namespace
	Namespace string
	// Lease Second and keepalive
	Lease  int64
	Logger *zap.Logger
}

// etcd endpoints
func WithEndpoints(v ...string) Option {
	return func(cc *Options) Option {
		previous := cc.Endpoints
		cc.Endpoints = v
		return WithEndpoints(previous...)
	}
}
func WithUserName(v string) Option {
	return func(cc *Options) Option {
		previous := cc.UserName
		cc.UserName = v
		return WithUserName(previous)
	}
}
func WithPassword(v string) Option {
	return func(cc *Options) Option {
		previous := cc.Password
		cc.Password = v
		return WithPassword(previous)
	}
}

// context
func WithContext(v context.Context) Option {
	return func(cc *Options) Option {
		previous := cc.Context
		cc.Context = v
		return WithContext(previous)
	}
}

// dial etcd timeout config
func WithDialTimeout(v time.Duration) Option {
	return func(cc *Options) Option {
		previous := cc.DialTimeout
		cc.DialTimeout = v
		return WithDialTimeout(previous)
	}
}

// CustomSet custom set etcd options
func WithCustomSet(v func(cfg *clientv3.Config)) Option {
	return func(cc *Options) Option {
		previous := cc.CustomSet
		cc.CustomSet = v
		return WithCustomSet(previous)
	}
}

// namespace
func WithNamespace(v string) Option {
	return func(cc *Options) Option {
		previous := cc.Namespace
		cc.Namespace = v
		return WithNamespace(previous)
	}
}

// Lease Second and keepalive
func WithLease(v int64) Option {
	return func(cc *Options) Option {
		previous := cc.Lease
		cc.Lease = v
		return WithLease(previous)
	}
}
func WithLogger(v *zap.Logger) Option {
	return func(cc *Options) Option {
		previous := cc.Logger
		cc.Logger = v
		return WithLogger(previous)
	}
}

// SetOption modify options
func (cc *Options) SetOption(opt Option) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *Options) ApplyOption(opts ...Option) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *Options) GetSetOption(opt Option) Option {
	return opt(cc)
}

// Option option define
type Option func(cc *Options) Option

// NewOptions create options instance.
func NewOptions(opts ...Option) *Options {
	cc := newDefaultOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogOptions != nil {
		watchDogOptions(cc)
	}
	return cc
}

// InstallOptionsWatchDog install watch dog
func InstallOptionsWatchDog(dog func(cc *Options)) {
	watchDogOptions = dog
}

var watchDogOptions func(cc *Options)

// newDefaultOptions new option with default value
func newDefaultOptions() *Options {
	cc := &Options{
		Endpoints:   []string{"127.0.0.1:2379"},
		UserName:    "",
		Password:    "",
		Context:     context.Background(),
		DialTimeout: 0,
		CustomSet: func(cfg *clientv3.Config) {
		},
		Namespace: "",
		Lease:     5,
		Logger:    zap.NewNop(),
	}
	return cc
}
