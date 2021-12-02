package hash

import (
	"context"

	"github.com/aggronmagi/walle/net/balancer"
	"github.com/aggronmagi/walle/net/discovery"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"go.uber.org/atomic"
)

const (
	BalancerName = "rr"
)

func RegisterBalancer() {
	balancer.RegisterBalancer(BalancerName, NewBalancerHash)
}

type hashBalancer struct {
	entries discovery.Entries
	index   atomic.Int32
}

func NewBalancerHash(discovery.Discovery) (balancer.Balancer, error) {
	return &hashBalancer{}, nil
}

func (b *hashBalancer) Update(es discovery.Entries) {
	b.entries = es
	return
}
func (b *hashBalancer) Pick(ctx context.Context, cmd packet.Command,
	uri, rq interface{}, md []process.MetadataOption) (discovery.Entry, error) {
	n := int(b.index.Add(1))
	entries := b.entries
	size := len(entries)
	for k := 0; k < len(entries); k++ {
		entry := entries[(n+k)%size]
		if !CheckEntryState(entry) {
			continue
		}
		return entry, nil
	}
	return nil, balancer.ErrNotValideEntry
}

var CheckEntryState = func(e discovery.Entry) bool {
	if e.State() != discovery.EntryStateOnline {
		return false
	}
	if e.Client() == nil {
		return false
	}
	return true
}
