package app

import (
	"fmt"
	"log"
)

// Service 接口 是基础服务对象容器. 负责 各个基础服务对象的状态维护
//
//go:generate mockgen -source svc.go -destination ../testpkg/mock_app/mock_app.go
type Service interface {
	Name() string
	Init(Stoper) (err error)
	Start(Stoper) (err error)
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
func (t *teeService) Init(s Stoper) (err error) {
	for _, v := range t.services {
		log.Printf("service %s wait init %T\n", v.Name(), v)
		err = v.Init(s)
		if err != nil {
			err = fmt.Errorf("service %s init failed:%w", v.Name(), err)
			log.Println(err)
			t.Finish()
			return
		}
		t.initd = append(t.initd, v)
		if s.IsStop() {
			return
		}
		log.Println("service", v.Name(), "init finish")
	}
	return
}

// Start 启动
func (t *teeService) Start(s Stoper) (err error) {
	for _, v := range t.initd {
		log.Println("service", v.Name(), "wait start", fmt.Sprintf("%T", v))
		err = v.Start(s)
		if err != nil {
			err = fmt.Errorf("service %s start failed:%w", v.Name(), err)
			log.Println(err)
			t.Stop()
			return
		}
		t.started = append(t.started, v)
		if s.IsStop() {
			return
		}
		log.Println("service", v.Name(), "start finish")
	}
	return
}

// Stop 关闭
func (t *teeService) Stop() {
	for k := len(t.started) - 1; k >= 0; k-- {
		log.Println("service", t.started[k].Name(), "stop")
		t.started[k].Stop()
	}
	return
}

// Finish 清理
func (t *teeService) Finish() {
	for k := len(t.initd) - 1; k >= 0; k-- {
		log.Println("service", t.initd[k].Name(), "finish")
		t.initd[k].Finish()
	}
	return
}

// FuncService for functions
//
//go:generate gogen option -n FuncSvcOption -o option.go
func walleFuncService() interface{} {
	return map[string]interface{}{
		// Service Name
		"Name": "funcSvc",
		// Init Function
		"Init": func(Stoper) (err error) {
			return nil
		},
		// Start Function
		"Start": func(Stoper) (err error) {
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
func (f *funcSvc) Init(s Stoper) (err error) {
	return f.cc.Init(s)
}

// Start 启动
func (f *funcSvc) Start(s Stoper) (err error) {
	return f.cc.Start(s)
}

// Stop 关闭
func (f *funcSvc) Stop() {
	f.cc.Stop()
}

// Finish 清理
func (f *funcSvc) Finish() {
	f.cc.Finish()
}
