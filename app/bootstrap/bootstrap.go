package bootstrap

import (
	"log"
	"sort"
	"sync/atomic"

	"github.com/walleframe/walle/app"
)

type priorityService struct {
	app.Service
	priority int
}

var registerService []priorityService
var startFlag atomic.Bool

// RegisterServiceByPriority regist service with specified priority.
func RegisterServiceByPriority(priority int, svc app.Service) {
	if startFlag.Load() {
		log.Panic("application is already started, CAN NOT register service now")
	}
	if svc.Name() == "noop" {
		log.Panicf("invalid service name, should REWRITE `Name() string` method. %#v", svc)
	}
	registerService = append(registerService, priorityService{
		Service:  svc,
		priority: priority,
	})
}

// RegisterService register normal priority service
func RegisterService(svc app.Service) {
	RegisterServiceByPriority(500, svc)
}

// RemoveService remove register service
func RemoveService(svc app.Service) {
	if startFlag.Load() {
		return
	}
	for k, v := range registerService {
		if v.Service == svc {
			registerService = append(registerService[:k], registerService[k+1:]...)
			break
		}
	}
}

// RemoveServiceByPriority remove register service by priority
func RemoveServiceByPriority(priority int) {
	if startFlag.Load() {
		return
	}
	for k, v := range registerService {
		if v.priority == priority {
			registerService = append(registerService[:k], registerService[k+1:]...)
			break
		}
	}
}

// Run new app and run
func Run() {
	if startFlag.Load() {
		return
	}
	startFlag.Store(true)
	sort.Slice(registerService, func(i, j int) bool {
		return registerService[i].priority < registerService[j].priority
	})
	service := make([]app.Service, 0, len(registerService))
	for _, v := range registerService {
		service = append(service, v.Service)
	}
	err := app.CreateApp(app.TeeService(service...)).Run()
	if err != nil {
		log.Fatal(err)
	}
}
