package roundrobin

import (
	"context"

	"github.com/walleframe/walle/network/balancer"
	"github.com/walleframe/walle/network/discovery"
	"go.uber.org/atomic"
)

// NewBalancer new roundrobin balancer
func NewBalancer(opts ...balancer.BalanceOption) balancer.PickerBuilder {
	cc := balancer.NewBalanceOptions(opts...)
	return &rrPickerBuilder{
		opts: cc,
	}
}

type rrPickerBuilder struct {
	opts *balancer.BalanceOptions
}

func (b *rrPickerBuilder) Build(es discovery.Entries) balancer.Picker {
	return &rrPicker{
		entries: es,
		opts:    b.opts,
	}
}

type rrPicker struct {
	entries discovery.Entries
	index   atomic.Int32
	opts    *balancer.BalanceOptions
}

func (b *rrPicker) Pick(ctx context.Context) (discovery.Entry, error) {
	entries := b.entries
	size := len(entries)
	for k := 0; k < len(entries); k++ {
		entry := entries[b.index.Add(1)%int32(size)]
		if !b.opts.EntryCheck(entry) {
			continue
		}
		return entry, nil
	}
	return nil, balancer.ErrNotValideEntry
}
