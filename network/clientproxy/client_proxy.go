package clientproxy

import (
	"context"
	"sync"
	"time"

	"github.com/walleframe/walle/network"
	"github.com/walleframe/walle/network/balancer"
	"github.com/walleframe/walle/network/discovery"
	"github.com/walleframe/walle/network/rpc"
	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/process/errcode"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Client = network.Client

type LinkMode int8

const (
	LinkModeImmediately LinkMode = iota
	LinkModeDelay
)

// NewClientFunc new client func.
// inner use for custom client
type NewClientFunc func(netwrok, addr string, inner *process.InnerOptions) (cli Client, err error)

type NewDiscoveryFunc func(path string, opts ...discovery.DiscoveryOption) (_ discovery.Discovery, err error)

//go:generate gogen option -n ProxyOption -o option.go
func walleClientProxy() interface{} {
	return map[string]interface{}{
		// NewEntry create custom entry for discovery new entry
		"NewClient": NewClientFunc(nil),
		// DiscoveryOptions custom discovery options
		"DiscoveryOptions": []discovery.DiscoveryOption{},
		// NewDiscovery create discovery
		"NewDiscovery": NewDiscoveryFunc(discovery.NewDiscovery),
		// BalancerName specify default balancer. empty means use server set value.
		"PickerBuilder": balancer.PickerBuilder(nil),
		// AsyncLink client async link server
		"AsyncLink": false,
		// UpdateBalanceAfterLink Whether the link must have been established before notifying the balancer
		"UpdateBalanceAfterLink": true,
		// LinkMode immediately or delay link
		"LinkMode": LinkMode(LinkModeImmediately),
		// UseAfterAllLink the first initialization must be fully linked before it can be used
		"UseAftreAllLink": true,
		// LinkInterval
		"LinkInterval": time.Duration(time.Second),
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
	}
}

// ClientProxy client wrap
type ClientProxy struct {
	pkgLoad  atomic.Int64
	sequence atomic.Int64

	path      string
	opts      *ProxyOptions
	discovery discovery.Discovery
	picker    balancer.Picker
	lock      sync.RWMutex
	entries   discovery.Entries
	closed    atomic.Bool
	init      atomic.Bool
}

func NewClientProxy(path string, opt ...ProxyOption) (proxy *ClientProxy, err error) {
	opts := NewProxyOptions(opt...)
	log := opts.FrameLogger.New("clientproxy.New")
	log.Must().String("path", path)

	proxy = &ClientProxy{
		opts: opts,
		path: path,
	}

	return
}

func (c *ClientProxy) InitProxy(ctx context.Context) (err error) {
	if !c.init.Load() {
		return
	}
	err = c.initProxy(ctx)
	if err != nil {
		return err
	}
	c.init.Store(true)
	return
}

func (c *ClientProxy) initProxy(ctx context.Context) (err error) {
	log := c.opts.FrameLogger.New("clientproxy.initProxy")
	d, err := c.opts.NewDiscovery(c.path, c.opts.DiscoveryOptions...)
	if err != nil {
		log.Error("new discovery failed", zap.Error(err))
		return err
	}
	c.discovery = d
	entries, err := d.GetAll(ctx)
	if err != nil {
		log.Error("initialize get all entries failed", zap.Error(err))
		return err
	}
	c.lock.Lock()
	c.entries = entries
	c.lock.Unlock()
	var wg *sync.WaitGroup
	if c.opts.UseAftreAllLink {
		wg = &sync.WaitGroup{}
	}
	for _, entry := range entries {
		if wg != nil {
			wg.Add(1)
		}
		c.linkEntry(entry, wg)
	}
	// wait all client link finish
	if wg != nil {
		wg.Wait()
	}
	c.picker = c.opts.PickerBuilder.Build(entries)
	err = c.discovery.WatchEventNotify(context.Background(), c.onEntriesUpdate)
	if err != nil {
		log.Error("watch entities failed", zap.Error(err))
		return err
	}

	return
}

func (c *ClientProxy) linkEntry(e discovery.Entry, wg *sync.WaitGroup) (err error) {
	if wg != nil {
		defer wg.Done()
	}
	net, addr := e.Address()
	cli, err := c.opts.NewClient(net, addr, process.NewInnerOptions(
		process.WithInnerOptionBindData(c),
		process.WithInnerOptionLoad(&c.pkgLoad),
		process.WithInnerOptionSequence(&c.sequence),
	))
	if err != nil {
		c.opts.FrameLogger.New("clientproxy.LinkEntity").Error("link failed", zap.Error(err),
			zap.String("net", net), zap.String("addr", addr),
		)
		return err
	}
	e.SetClient(cli)
	return
}

func (c *ClientProxy) onEntriesUpdate(entries discovery.Entries) {
	// sync client interface
	for _, v := range entries {
		for _, l := range c.entries {
			if l.Equals(v) && l.Client() != nil {
				v.SetClient(l.Client())
			}
		}
	}
	var wg *sync.WaitGroup
	if c.opts.UpdateBalanceAfterLink {
		wg = &sync.WaitGroup{}
	}
	// compare difference
	add, remove := c.entries.Diff(entries)
	// remove last client
	for _, v := range remove {
		cli := v.Client()
		if cli != nil {
			cli.Close()
		}
	}
	// add client
	for _, v := range add {
		if wg != nil {
			wg.Add(1)
		}
		c.linkEntry(v, wg)
	}
	if wg != nil {
		wg.Wait()
	}
	c.lock.Lock()
	c.entries = entries
	c.lock.Unlock()
	c.picker = c.opts.PickerBuilder.Build(entries)
}

func (c *ClientProxy) linkEntities() {
	for {
		if c.closed.Load() {
			return
		}
		c.lock.RLock()
		entities := c.entries
		c.lock.RUnlock()
		for _, v := range entities {
			if v.Client() == nil {
				c.linkEntry(v, nil)
			}
		}
		time.Sleep(c.opts.LinkInterval)
	}
}

func (c *ClientProxy) Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *rpc.CallOptions) (err error) {
	entry, err := c.picker.Pick(ctx)
	if err != nil {
		c.opts.FrameLogger.Logger().Error("pick failed", zap.Error(err))
		return err
	}
	if entry.Client() != nil {
		return entry.Client().Call(ctx, uri, rq, rs, opts)
	}
	return errcode.ErrSessionClosed
}

func (c *ClientProxy) AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af process.RouterFunc, opts *rpc.AsyncCallOptions) (err error) {
	entry, err := c.picker.Pick(ctx)
	if err != nil {
		return err
	}
	if entry.Client() != nil {
		return entry.Client().AsyncCall(ctx, uri, rq, af, opts)
	}
	return errcode.ErrSessionClosed
}

func (c *ClientProxy) Notify(ctx context.Context, uri interface{}, rq interface{}, opts *rpc.NoticeOptions) (err error) {
	entry, err := c.picker.Pick(ctx)
	if err != nil {
		return err
	}
	if entry.Client() != nil {
		return entry.Client().Notify(ctx, uri, rq, opts)
	}
	return errcode.ErrSessionClosed
}

func (c *ClientProxy) Close(ctx context.Context) {
	c.discovery.Close(ctx)
	entries := c.entries
	for _, v := range entries {
		if v.Client() != nil {
			v.Client().Close()
		}
	}
	c.closed.Store(true)
	return
}
