// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n ConsistentOption -o option.go"
// Version: 0.0.2

package consistenthash

import (
	md5 "crypto/md5"
	binary "encoding/binary"

	balancer "github.com/walleframe/walle/network/balancer"
	discovery "github.com/walleframe/walle/network/discovery"
	util "github.com/walleframe/walle/util"
)

var _ = walleConsistentOption()

type ConsistentOptions struct {
	// EntryCheck check entry state when pick.
	EntryCheck balancer.PickerCheckFunc
	// GenHashValueByEntry generate visual node value
	GenHashValueByEntry func(entry discovery.Entry) (ids []uint32)
}

// EntryCheck check entry state when pick.
func WithEntryCheck(v balancer.PickerCheckFunc) ConsistentOption {
	return func(cc *ConsistentOptions) ConsistentOption {
		previous := cc.EntryCheck
		cc.EntryCheck = v
		return WithEntryCheck(previous)
	}
}

// GenHashValueByEntry generate visual node value
func WithGenHashValueByEntry(v func(entry discovery.Entry) (ids []uint32)) ConsistentOption {
	return func(cc *ConsistentOptions) ConsistentOption {
		previous := cc.GenHashValueByEntry
		cc.GenHashValueByEntry = v
		return WithGenHashValueByEntry(previous)
	}
}

// SetOption modify options
func (cc *ConsistentOptions) SetOption(opt ConsistentOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *ConsistentOptions) ApplyOption(opts ...ConsistentOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *ConsistentOptions) GetSetOption(opt ConsistentOption) ConsistentOption {
	return opt(cc)
}

// ConsistentOption option define
type ConsistentOption func(cc *ConsistentOptions) ConsistentOption

// NewConsistentOptions create options instance.
func NewConsistentOptions(opts ...ConsistentOption) *ConsistentOptions {
	cc := newDefaultConsistentOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogConsistentOptions != nil {
		watchDogConsistentOptions(cc)
	}
	return cc
}

// InstallConsistentOptionsWatchDog install watch dog
func InstallConsistentOptionsWatchDog(dog func(cc *ConsistentOptions)) {
	watchDogConsistentOptions = dog
}

var watchDogConsistentOptions func(cc *ConsistentOptions)

// newDefaultConsistentOptions new option with default value
func newDefaultConsistentOptions() *ConsistentOptions {
	cc := &ConsistentOptions{
		EntryCheck: balancer.CheckEntryState,
		GenHashValueByEntry: func(entry discovery.Entry) (ids []uint32) {
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
	return cc
}
