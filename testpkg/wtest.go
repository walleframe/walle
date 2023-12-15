package testpkg

import (
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/walleframe/walle/app"
)

// FuncCall 用于测试函数调用。
//
//go:generate mockgen -source wtest.go -package testpkg -destination gentest.go
type FuncCall interface {
	Call(v ...interface{})
}

func Run(svc app.Service) func() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	app.StopSignal = func() <-chan os.Signal {
		c := make(chan os.Signal, 1)
		// signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		go func() {
			wg.Wait()
			c <- syscall.SIGINT
		}()
		return c
	}
	go app.CreateApp(svc).Run()
	runtime.Gosched()
	time.Sleep(time.Millisecond * 50)
	return func() {
		wg.Done()
	}
}
