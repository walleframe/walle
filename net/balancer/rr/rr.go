package rr

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
	balancer.RegisterBalancer(BalancerName, NewBalancerRR)
}

type rrBalancer struct {
	entries discovery.Entries
	index   atomic.Int32
}

func NewBalancerRR(discovery.Discovery) (balancer.Balancer, error) {
	return &rrBalancer{}, nil
}

func (b *rrBalancer) Update(es discovery.Entries) {
	b.entries = es
	return
}
func (b *rrBalancer) Pick(ctx context.Context, cmd packet.Command,
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
