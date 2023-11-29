package hash

import (
	"context"

	"github.com/walleframe/walle/network/balancer"
	"github.com/walleframe/walle/network/discovery"
)

// NewBalancer new hash balancer
func NewBalancer(opts ...balancer.BalanceOption) balancer.PickerBuilder {
	cc := balancer.NewBalanceOptions(opts...)
	return &hashPickerBuilder{
		opts: cc,
	}
}

type hashPickerBuilder struct {
	opts *balancer.BalanceOptions
}

func (b *hashPickerBuilder) Build(es discovery.Entries) balancer.Picker {
	return &hashPicker{
		entries: es,
		opts:    b.opts,
	}
}

type hashPicker struct {
	entries discovery.Entries
	opts    *balancer.BalanceOptions
}

func (b *hashPicker) Pick(ctx context.Context) (discovery.Entry, error) {
	entries := b.entries
	size := int64(len(entries))
	index, err := balancer.GetBalanceIndex(ctx)
	if err != nil {
		return nil, err
	}
	entry := entries[index%size]
	if !b.opts.EntryCheck(entry) {
		return nil, balancer.ErrNotValideEntry
	}
	return entry, nil
}
