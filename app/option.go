// Code generated by "gogen option"; DO NOT EDIT.
// Exec: "gogen option -n FuncSvcOption -o option.go"
// Version: 0.0.2

package app

var _ = walleFuncService()

// FuncService for functions
type FuncSvcOptions struct {
	// Service Name
	Name string
	// Init Function
	Init func() (err error)
	// Start Function
	Start func() (err error)
	// Stop Function
	Stop func()
	// Finish Function
	Finish func()
}

// Service Name
func WithName(v string) FuncSvcOption {
	return func(cc *FuncSvcOptions) FuncSvcOption {
		previous := cc.Name
		cc.Name = v
		return WithName(previous)
	}
}

// Init Function
func WithInit(v func() (err error)) FuncSvcOption {
	return func(cc *FuncSvcOptions) FuncSvcOption {
		previous := cc.Init
		cc.Init = v
		return WithInit(previous)
	}
}

// Start Function
func WithStart(v func() (err error)) FuncSvcOption {
	return func(cc *FuncSvcOptions) FuncSvcOption {
		previous := cc.Start
		cc.Start = v
		return WithStart(previous)
	}
}

// Stop Function
func WithStop(v func()) FuncSvcOption {
	return func(cc *FuncSvcOptions) FuncSvcOption {
		previous := cc.Stop
		cc.Stop = v
		return WithStop(previous)
	}
}

// Finish Function
func WithFinish(v func()) FuncSvcOption {
	return func(cc *FuncSvcOptions) FuncSvcOption {
		previous := cc.Finish
		cc.Finish = v
		return WithFinish(previous)
	}
}

// SetOption modify options
func (cc *FuncSvcOptions) SetOption(opt FuncSvcOption) {
	_ = opt(cc)
}

// ApplyOption modify options
func (cc *FuncSvcOptions) ApplyOption(opts ...FuncSvcOption) {
	for _, opt := range opts {
		_ = opt(cc)
	}
}

// GetSetOption modify and get last option
func (cc *FuncSvcOptions) GetSetOption(opt FuncSvcOption) FuncSvcOption {
	return opt(cc)
}

// FuncSvcOption option define
type FuncSvcOption func(cc *FuncSvcOptions) FuncSvcOption

// NewFuncSvcOptions create options instance.
func NewFuncSvcOptions(opts ...FuncSvcOption) *FuncSvcOptions {
	cc := newDefaultFuncSvcOptions()
	for _, opt := range opts {
		_ = opt(cc)
	}
	if watchDogFuncSvcOptions != nil {
		watchDogFuncSvcOptions(cc)
	}
	return cc
}

// InstallFuncSvcOptionsWatchDog install watch dog
func InstallFuncSvcOptionsWatchDog(dog func(cc *FuncSvcOptions)) {
	watchDogFuncSvcOptions = dog
}

var watchDogFuncSvcOptions func(cc *FuncSvcOptions)

// newDefaultFuncSvcOptions new option with default value
func newDefaultFuncSvcOptions() *FuncSvcOptions {
	cc := &FuncSvcOptions{
		Name: "funcSvc",
		Init: func() (err error) {
			return nil
		},
		Start: func() (err error) {
			return nil
		},
		Stop: func() {
		},
		Finish: func() {
		},
	}
	return cc
}
