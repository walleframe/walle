package consistenthash

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"sort"

	"github.com/aggronmagi/walle/network/balancer"
	"github.com/aggronmagi/walle/network/discovery"
	"github.com/aggronmagi/walle/util"
)

//go:generate gogen option -n ConsistentOption -o option.go
func walleConsistentOption() interface{} {
	return map[string]interface{}{
		// EntryCheck check entry state when pick.
		"EntryCheck": balancer.PickerCheckFunc(balancer.CheckEntryState),
		// GenHashValueByEntry generate visual node value
		"GenHashValueByEntry": func(entry discovery.Entry) (ids []uint32) {
			ids = make([]uint32, 0, 32)
			for i := 0; i < 8; i++ {
				key := genKey(entry, i)
				data := md5.Sum(util.StringToBytes(key))
				ids = append(ids, binary.LittleEndian.Uint32(data[0*4:]))
				ids = append(ids, binary.LittleEndian.Uint32(data[1*4:]))
				ids = append(ids, binary.LittleEndian.Uint32(data[2*4:]))
				ids = append(ids, binary.LittleEndian.Uint32(data[3*4:]))
			}
			return
		},
	}
}

func genKey(entry discovery.Entry, id int) string {
	key := entry.String()
	buf := util.Builder{}
	buf.Grow(len(key) + 4)
	buf.WriteString(key)
	buf.WriteByte('#')
	buf.WriteInt(id)
	return buf.String()
}

type consistentEntry struct {
	entry discovery.Entry
	value uint32
}

// NewBalancer new consistent balancer
func NewBalancer(opts ...ConsistentOption) balancer.PickerBuilder {
	cc := NewConsistentOptions(opts...)
	return &consistentPickerBuilder{
		opts: cc,
	}
}

type consistentPickerBuilder struct {
	opts *ConsistentOptions
}

func (b *consistentPickerBuilder) Build(es discovery.Entries) balancer.Picker {
	enties := make([]consistentEntry, 0, 32)
	for _, entry := range es {
		values := b.opts.GenHashValueByEntry(entry)
		for _, value := range values {
			enties = append(enties, consistentEntry{
				entry: entry,
				value: value,
			})
		}
	}
	sort.Slice(enties, func(i, j int) bool {
		return enties[i].value <= enties[j].value
	})
	return &consistentPicker{
		entries: enties,
		opts:    b.opts,
	}
}

type consistentPicker struct {
	entries []consistentEntry
	opts    *ConsistentOptions
}

func (b *consistentPicker) Pick(ctx context.Context) (discovery.Entry, error) {
	entries := b.entries
	index, err := balancer.GetBalanceIndex(ctx)
	if err != nil {
		return nil, err
	}
	id := sort.Search(len(b.entries), func(i int) bool {
		return entries[i].value >= uint32(index)
	})
	if id >= len(entries) {
		id = 0
	}
	entry := entries[id]
	if !b.opts.EntryCheck(entry.entry) {
		return nil, balancer.ErrNotValideEntry
	}
	return entry.entry, nil
}
