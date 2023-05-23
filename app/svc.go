package app

import "fmt"

// Service 接口 是基础服务对象容器. 负责 各个基础服务对象的状态维护
//go:generate mockgen -source svc.go -package app_test -destination mock_test.go
type Service interface {
	Name() string
	Init() (err error)
	Start() (err error)
	Stop()
	Finish()
}

// 聚合多个服务。正序启动，逆序清理
type teeService struct {
	services []Service
	initd    []Service
	started  []Service
}

// TeeService 聚合多个服务。正序启动，逆序清理
func TeeService(svcs ...Service) Service {
	return &teeService{
		services: svcs,
		initd:    make([]Service, 0, len(svcs)),
		started:  make([]Service, 0, len(svcs)),
	}
}

func (t *teeService) Name() string {
	return "TeeService"
}

// Init 初始化
func (t *teeService) Init() (err error) {
	for _, v := range t.services {
		err = v.Init()
		if err != nil {
			err = fmt.Errorf("service %s init failed:%w", v.Name(), err)
			defer t.Finish()
			return
		}
		t.initd = append(t.initd, v)
	}
	return
}

// Start 启动
func (t *teeService) Start() (err error) {
	for _, v := range t.initd {
		err = v.Start()
		if err != nil {
			err = fmt.Errorf("service %s start failed:%w", v.Name(), err)
			defer t.Stop()
			return
		}
		t.started = append(t.started, v)
	}
	return
}

// Stop 关闭
func (t *teeService) Stop() {
	for k := len(t.started) - 1; k >= 0; k-- {
		t.started[k].Stop()
	}
}

// Finish 清理
func (t *teeService) Finish() {
	for k := len(t.initd) - 1; k >= 0; k-- {
		t.initd[k].Finish()
	}
}

// FuncService for functions
//go:generate gogen option -n FuncSvcOption -o option.go
func walleFuncService() interface{} {
	return map[string]interface{}{
		// Service Name
		"Name": "funcSvc",
		// Init Function
		"Init": func() (err error) {
			return nil
		},
		// Start Function
		"Start": func() (err error) {
			return nil
		},
		// Stop Function
		"Stop": func() {
		},
		// Finish Function
		"Finish": func() {
		},
	}
}

// FuncService 使用函数构造服务
func FuncService(opts ...FuncSvcOption) Service {
	return &funcSvc{
		cc: NewFuncSvcOptions(opts...),
	}
}

type funcSvc struct {
	cc *FuncSvcOptions
}

func (f *funcSvc) Name() string {
	return f.cc.Name
}

// Init 初始化
func (f *funcSvc) Init() (err error) {
	return f.cc.Init()
}

// Start 启动
func (f *funcSvc) Start() (err error) {
	return f.cc.Start()
}

// Stop 关闭
func (f *funcSvc) Stop() {
	f.cc.Stop()
}

// Finish 清理
func (f *funcSvc) Finish() {
	f.cc.Finish()
}
