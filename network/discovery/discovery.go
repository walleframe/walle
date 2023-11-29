package discovery

import (
	"context"
	"strings"

	"github.com/walleframe/walle/kvstore"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

//go:generate mockgen -source discovery.go -destination ../../testpkg/mock_discovery/discovery.go

// Discovery use for discover netwrok service entry
type Discovery interface {
	Watch(ctx context.Context) (changes <-chan Entries, err error)
	GetAll(ctx context.Context) (all Entries, err error)
	Close(ctx context.Context)
	WatchEventNotify(ctx context.Context, eventNotify func(Entries)) (err error)
}

//go:generate gogen option -n DiscoveryOption -f Discovery -o option.discovery.go
func walleDiscovery() interface{} {
	return map[string]interface{}{
		// NewEntry create custom entry for discovery new entry
		"NewEntry": func() Entry {
			return &Node{}
		},
		// Codec use for decode entry.
		"Codec": EntryCodec(NodeJsonEntryCodec),
		// Store use for discover entries
		"Store": kvstore.Store(nil),
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
	}
}

type discovery struct {
	opts    *DiscoveryOptions
	entries Entries
	store   kvstore.Store
	path    string
	ch      chan struct{}
}

func NewDiscovery(path string, opts ...DiscoveryOption) (_ Discovery, err error) {
	path = kvstore.Normalize(path) + "/"
	// path = kvstore.Join(kvstore.SplitKey(path)...) + "/"
	return &discovery{
		opts:    NewDiscoveryOptions(opts...),
		entries: make(Entries, 0, 16),
		path:    path,
		ch:      make(chan struct{}),
	}, nil
}

func (d *discovery) Watch(ctx context.Context) (_ <-chan Entries, err error) {
	notify, err := d.opts.Store.WatchTree(ctx, d.path, d.ch)
	if err != nil {
		return nil, err
	}
	changes := make(chan Entries)
	go func() {
		for {
			select {
			case <-d.ch:
				return
			case kvs, ok := <-notify:
				if !ok {
					return
				}
				entries, _ := d.convertEntries(kvs)
				d.entries = entries
				changes <- entries
			}
		}
	}()
	return changes, nil
}

func (d *discovery) WatchEventNotify(ctx context.Context, eventNotify func(Entries)) (err error) {
	notify, err := d.opts.Store.WatchTree(ctx, d.path, d.ch)
	if err != nil {
		d.opts.FrameLogger.New("discovery.Watch").Error("watch failed", zap.Error(err), zap.String("path", d.path))
		return err
	}
	go func() {
		for {
			select {
			case <-d.ch:
				return
			case kvs, ok := <-notify:
				if !ok {
					return
				}
				entries, _ := d.convertEntries(kvs)
				d.entries = entries
				eventNotify(entries)
			}
		}
	}()
	return
}
func (d *discovery) GetAll(ctx context.Context) (all Entries, err error) {
	if d.entries != nil {
		all = d.entries
		return
	}
	kvs, err := d.opts.Store.List(ctx, d.path)
	if err != nil {
		return nil, err
	}

	all, err = d.convertEntries(kvs)

	return
}
func (d *discovery) Close(ctx context.Context) {
	close(d.ch)
	return
}

func (d *discovery) convertEntries(kvs []*kvstore.KVPair) (all Entries, err error) {
	all = make(Entries, 0, len(kvs))
	for _, v := range kvs {
		// remove kv
		if len(v.Value) < 1 {
			continue
		}
		// unmarshal
		e := d.opts.NewEntry()
		v.Key = strings.TrimLeft(v.Key, d.path)
		err2 := d.opts.Codec.Unmarshal(e, v.Key, v.Value)
		if err2 != nil {
			// FIXME: 是否抛出错误？ 或者通过配置让上层选择
			continue
		}
		all = append(all, e)
	}
	return
}
