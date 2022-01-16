package balancer

import (
	"context"
	"errors"

	"github.com/aggronmagi/walle/network/discovery"
)
//go:generate mockgen -source balance.go -destination ../../testpkg/mock_balancer/balancer.go
// PickerBuilder use for build picker
type PickerBuilder interface {
	Build(discovery.Entries) Picker
}

// Picker pick one useful entry
type Picker interface {
	Pick(ctx context.Context) (discovery.Entry, error)
}

// PickerCheckFunc check entry function,check it when picker pick entry
type PickerCheckFunc func(e discovery.Entry) bool

var CheckEntryState PickerCheckFunc = func(e discovery.Entry) bool {
	if e.State() != discovery.EntryStateOnline {
		return false
	}
	if e.Client() == nil {
		return false
	}
	return true
}

type PickerPrepareFunc func(context.Context)

//go:generate gogen option -n BalanceOption -o option.go
func walleBalanceOption() interface{} {
	return map[string]interface{}{
		// EntryCheck check entry state when pick.
		"EntryCheck": PickerCheckFunc(CheckEntryState),
	}
}

var (
	ErrInvalidBalanceName = errors.New("invalid balance name")
	ErrNotValideEntry     = errors.New("not valid entry")
	ErrNotSetBalanceIndex = errors.New("not set balance index")
)

type balanceIndex struct{}

func WithBalanceIndex(ctx context.Context, value int64) context.Context {
	return context.WithValue(ctx, balanceIndex{}, value)
}

func GetBalanceIndex(ctx context.Context) (int64, error) {
	v := ctx.Value(balanceIndex{})
	if v == nil {
		return 0, ErrNotSetBalanceIndex
	}
	i, ok := v.(int64)
	if !ok {
		return 0, ErrNotSetBalanceIndex
	}
	return i, nil
}

// type BalanceFactory func(discovery.Discovery) (PickerBuilder, error)
// var balanceFactoryMap = make(map[string]BalanceFactory)
// func RegisterBalancer(name string, newBalance BalanceFactory) {
// 	balanceFactoryMap[name] = newBalance
// }
// func GetBalancer(name string, d discovery.Discovery) (PickerBuilder, error) {
// 	if fac, ok := balanceFactoryMap[name]; ok {
// 		return fac(d)
// 	}
// 	return nil, ErrInvalidBalanceName
// }
