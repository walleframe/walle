package testpkg

import (
	"runtime"
	"sync"
	"time"

	"github.com/aggronmagi/walle/app"
)

// FuncCall 用于测试函数调用。
//go:generate mockgen -source wtest.go -package testpkg -destination gentest.go
type FuncCall interface {
	Call(v ...interface{})
}

func Run(svc app.Service) func() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	app.WaitStopSignal = func() {
		wg.Wait()
	}
	go app.CreateApp(svc).Run()
	runtime.Gosched()
	time.Sleep(time.Millisecond * 50)
	return func() {
		wg.Done()
	}
}
