// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n BalanceOption -o option.go"
// Version: 0.0.4

package balancer

var _ = walleBalanceOption()

type BalanceOptions struct {
	// EntryCheck check entry state when pick.
	EntryCheck PickerCheckFunc
}

// EntryCheck check entry state when pick.
func WithEntryCheck(v PickerCheckFunc) BalanceOption {
	return func(cc *BalanceOptions) BalanceOption {
		previous := cc.EntryCheck
		cc.EntryCheck = v
		return WithEntryCheck(previous)
	}
}

// SetOption modify options
func (cc *BalanceOptions) SetOption(opt BalanceOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *BalanceOptions) ApplyOption(opts ...BalanceOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *BalanceOptions) GetSetOption(opt BalanceOption) BalanceOption {
	return opt(cc)
}

// BalanceOption option define
type BalanceOption func(cc *BalanceOptions) BalanceOption

// NewBalanceOptions create options instance.
func NewBalanceOptions(opts ...BalanceOption) *BalanceOptions {
	cc := newDefaultBalanceOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogBalanceOptions != nil {
		watchDogBalanceOptions(cc)
	}
	return cc
}

// InstallBalanceOptionsWatchDog install watch dog
func InstallBalanceOptionsWatchDog(dog func(cc *BalanceOptions)) {
	watchDogBalanceOptions = dog
}

var watchDogBalanceOptions func(cc *BalanceOptions)

// newDefaultBalanceOptions new option with default value
func newDefaultBalanceOptions() *BalanceOptions {
	cc := &BalanceOptions{
		EntryCheck: CheckEntryState,
	}
	return cc
}
