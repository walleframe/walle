package balancer

import (
	"context"
	"errors"

	"github.com/aggronmagi/walle/net/discovery"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
)

type Balancer interface {
	Update(discovery.Entries)
	Pick(ctx context.Context, cmd packet.Command, uri, rq interface{}, md []process.MetadataOption) (discovery.Entry, error)
}

type BalanceFactory func(discovery.Discovery) (Balancer, error)

var balanceFactoryMap = make(map[string]BalanceFactory)

func RegisterBalancer(name string, newBalance BalanceFactory) {
	balanceFactoryMap[name] = newBalance
}

func GetBalancer(name string, d discovery.Discovery) (Balancer, error) {
	if fac, ok := balanceFactoryMap[name]; ok {
		return fac(d)
	}
	return nil, ErrInvalidBalanceName
}

var (
	ErrInvalidBalanceName = errors.New("invalid balance name")
	ErrNotValideEntry     = errors.New("not valid entry")
)
