package consistenthash

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/network"
	"github.com/walleframe/walle/network/balancer"
	"github.com/walleframe/walle/network/discovery"
	"github.com/walleframe/walle/testpkg/mock_discovery"
	"github.com/walleframe/walle/testpkg/mock_network"
)

const (
	balanceTestTimes = 8
)

func TestBalance(t *testing.T) {
	// prepare test
	mctl := gomock.NewController(t)
	entries := make([]discovery.Entry, 0, balanceTestTimes)
	state := make([]discovery.EntryState, balanceTestTimes)
	for i := 0; i < balanceTestTimes; i++ {
		k := i
		_ = k
		state[i] = discovery.EntryStateOnline
		client := mock_network.NewMockClient(mctl)
		entry := mock_discovery.NewMockEntry(mctl)
		entry.EXPECT().String().AnyTimes().Return(fmt.Sprintf("10.0.1.%d:8080", i))
		entry.EXPECT().Client().AnyTimes().Return(network.Client(client))
		entry.EXPECT().State().AnyTimes().DoAndReturn(func() discovery.EntryState {
			return state[k]
		})
		entry.EXPECT().Metadata("id").AnyTimes().Return(fmt.Sprint(i))
		entries = append(entries, entry)
	}
	// start
	builder := NewBalancer()
	picker := builder.Build(entries)
	assert.NotNil(t, picker, "new balancer picker")

	// pick
	for i := 0; i < balanceTestTimes; i++ {
		ctx := balancer.WithBalanceIndex(context.Background(), int64(i))
		entry, err := picker.Pick(ctx)
		assert.Nil(t, err, "[%d]pick one not error", i)
		assert.NotNil(t, entry, "[%d]pick one entry", i)
		assert.NotNil(t, entry.Client(), "[%d]get entry client", i)
		//assert.Equal(t, fmt.Sprint((i+1)%balanceTestTimes), entry.Metadata("id"), "[%d]compare id", i)
	}
	for k := range state {
		if k%2 == 0 {
			state[k] = discovery.EntryStateOffline
		}
	}
	for i := 0; i < balanceTestTimes; i++ {
		ctx := balancer.WithBalanceIndex(context.Background(), int64(i))
		entry, err := picker.Pick(ctx)
		assert.Nil(t, err, "[%d]pick one not error", i)
		id, _ := strconv.ParseInt(entry.Metadata("id"), 10, 8)
		assert.True(t, id%2 != 0, "[%d]get disabled entry", i)
	}
}
