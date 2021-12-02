package clientproxy

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aggronmagi/walle/net/balancer"
	"github.com/aggronmagi/walle/net/iface"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/zaplog"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/aggronmagi/walle/net/discovery"
)

type Client = iface.Client

// NewClientFunc new client func.
// inner use for custom client
type NewClientFunc func(netwrok, addr string, inner *process.InnerOptions) (cli Client, err error)

type NewDiscoveryFunc func(path string, opts ...discovery.DiscoveryOption) (_ discovery.Discovery, err error)

//go:generate gogen option -n ProxyOption -o option.go
func walleClientProxy() interface{} {
	return map[string]interface{}{
		// NewEntry create custom entry for discovery new entry
		"NewClient": NewClientFunc(nil),
		// NewDiscovery create discovery
		"NewDiscovery": NewDiscoveryFunc(discovery.NewDiscovery),
		// DiscoveryOptions custom discovery options
		"DiscoveryOptions": []discovery.DiscoveryOption{},
		// BalancerName specify default balancer. empty means use server set value.
		"BalanceName": "",
		// LinkInterval
		"LinkInterval": time.Duration(time.Second),
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.Frame),
	}
}

// ClientProxy client wrap
type ClientProxy struct {
	pkgLoad  atomic.Int64
	sequence atomic.Int64

	discovery discovery.Discovery
	balancer  balancer.Balancer
	opts      *ProxyOptions
	lock      sync.RWMutex
	entries   discovery.Entries
	closed    atomic.Bool
}

func NewClient(path string, opt ...ProxyOption) (proxy *ClientProxy, err error) {
	opts := NewProxyOptions(opt...)
	log := opts.FrameLogger.New("clientproxy.New")
	log.Must().String("path", path)
	d, err := opts.NewDiscovery(path, opts.DiscoveryOptions...)
	if err != nil {
		log.Error("new discovery failed", zap.Error(err))
		return nil, err
	}
	var b balancer.Balancer
	if len(opts.BalanceName) == 0 {
		entities, err := d.GetAll(context.Background())
		if err != nil {
			log.Error("get remote entities failed", zap.Error(err))
			return nil, err
		}
		if len(entities) < 1 {
			log.Error("remote entities empty")
			return nil, errors.New("not found entities and not config balance")
		}
		opts.BalanceName = entities[0].BalanceName()
	}

	b, err = balancer.GetBalancer(opts.BalanceName, d)
	if err != nil {
		log.Error("invalid balancer name", zap.String("balancer", opts.BalanceName))
		return nil, err
	}

	proxy = &ClientProxy{
		opts:      opts,
		discovery: d,
		balancer:  b,
	}
	err = proxy.discovery.WatchEventNotify(context.Background(), proxy.onEntriesUpdate)
	if err != nil {
		log.Error("watch entities failed", zap.Error(err))
		return nil, err
	}
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
		c.linkEntry(v)
	}
	c.lock.Lock()
	c.entries = entries
	c.lock.Unlock()
	c.balancer.Update(entries)
}

func (c *ClientProxy) linkEntry(e discovery.Entry) (err error) {
	net, addr := e.Address()
	cli, err := c.opts.NewClient(net, addr, process.NewInnerOptions(
		process.WithInnerOptionsBindData(c),
		process.WithInnerOptionsLoad(&c.pkgLoad),
		process.WithInnerOptionsSequence(&c.sequence),
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
				c.linkEntry(v)
			}
		}
		time.Sleep(c.opts.LinkInterval)
	}
}

//
func (c *ClientProxy) Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *process.CallOptions) (err error) {
	entry, err := c.balancer.Pick(ctx, packet.Command_Request, uri, rq, opts.Metadata)
	if err != nil {
		c.opts.FrameLogger.Logger().Error("pick failed", zap.Error(err))
		return err
	}
	if entry.Client() != nil {
		return entry.Client().Call(ctx, uri, rq, rs, opts)
	}
	return packet.ErrSessionClosed
}
func (c *ClientProxy) AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af process.RouterFunc, opts *process.AsyncCallOptions) (err error) {
	entry, err := c.balancer.Pick(ctx, packet.Command_Request, uri, rq, opts.Metadata)
	if err != nil {
		return err
	}
	if entry.Client() != nil {
		return entry.Client().AsyncCall(ctx, uri, rq, af, opts)
	}
	return packet.ErrSessionClosed
}
func (c *ClientProxy) Notify(ctx context.Context, uri interface{}, rq interface{}, opts *process.NoticeOptions) (err error) {
	entry, err := c.balancer.Pick(ctx, packet.Command_Oneway, uri, rq, opts.Metadata)
	if err != nil {
		return err
	}
	if entry.Client() != nil {
		return entry.Client().Notify(ctx, uri, rq, opts)
	}
	return packet.ErrSessionClosed
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
