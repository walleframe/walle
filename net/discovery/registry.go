package discovery

import (
	"context"
	"net"

	"github.com/aggronmagi/walle/kvstore"
	"github.com/aggronmagi/walle/util"
	"github.com/aggronmagi/walle/zaplog"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Registry use for netwrok service register entry
type Registry interface {
	// NewEntry create entry,wait for register
	NewEntry(ctx context.Context, addr net.Addr) (err error)
	// Online set entry online status and update to store
	Online(ctx context.Context) (err error)
	// Offline set entry offline status and update to store
	Offline(ctx context.Context) (err error)
	// Close clean entry info
	Clean(ctx context.Context) (err error)
}

//go:generate gogen option -n RegistryOption -f Registry -o option.registry.go
func walleRegistry() interface{} {
	return map[string]interface{}{
		// NewEntry create custom entry for registry
		"NewEntry": func(addr net.Addr) (_ Entry, err error) {
			// NOTE: custom set register ip,port
			ip, port, err := net.SplitHostPort(addr.String())
			if err != nil {
				return
			}
			// TODO: read from env (use for docker/k8s)
			if ip == "::" { // listen 0.0.0.0
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
		// Codec use for encode entry.
		"Codec": EntryCodec(NodeJsonEntryCodec),
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.Frame),
	}
}

type registry struct {
	opts  *RegistryOptions
	entry Entry
	store kvstore.Store
	path  string
}

func NewRegistry(path string, store kvstore.Store, opts ...RegistryOption) Registry {
	return &registry{
		opts:  NewRegistryOptions(opts...),
		entry: nil,
		store: store,
		path:  kvstore.Normalize(path),
	}
}

func (r *registry) NewEntry(ctx context.Context, addr net.Addr) (err error) {
	r.entry, err = r.opts.NewEntry(addr)
	if err != nil {
		r.opts.FrameLogger.New("Registry.NewEntry").Error("new failed", zap.Error(err))
		return err
	}
	return
}
func (r *registry) Online(ctx context.Context) (err error) {
	r.entry.ModifyState(EntryStateOnline)
	key, value, err := r.opts.Codec.Mashal(r.entry)
	if err != nil {
		r.opts.FrameLogger.New("Registry.Online").Error("marshal failed", zap.Error(err))
		return err
	}
	err = r.store.Put(ctx, kvstore.Join(r.path, string(key)), value)
	if err != nil {
		r.opts.FrameLogger.New("Registry.Online").Error("put failed", zap.Error(err))
	}
	return
}
func (r *registry) Offline(ctx context.Context) (err error) {
	r.entry.ModifyState(EntryStateOnline)
	key, value, err := r.opts.Codec.Mashal(r.entry)
	if err != nil {
		r.opts.FrameLogger.New("Registry.Offline").Error("marshal failed", zap.Error(err))
		return err
	}
	err = r.store.Put(ctx, kvstore.Join(r.path, string(key)), value)
	if err != nil {
		r.opts.FrameLogger.New("Registry.Offline").Error("put failed", zap.Error(err))
	}
	return
}

func (r *registry) Clean(ctx context.Context) (err error) {
	key, _, err := r.opts.Codec.Mashal(r.entry)
	if err != nil {
		r.opts.FrameLogger.New("Registry.Clean").Error("marshal failed", zap.Error(err))
		return err
	}
	err = r.store.Delete(ctx, kvstore.Join(r.path, string(key)))
	if err != nil {
		r.opts.FrameLogger.New("Registry.Clean").Error("delete failed", zap.Error(err))
	}
	return
}
