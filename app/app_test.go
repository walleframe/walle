package app_test

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aggronmagi/walle/app"
	"github.com/golang/mock/gomock"
)

// This example passes a context with a signal to tell a blocking function that
// it should abandon its work after a signal is received.
func TestApplicationRun(t *testing.T) {
	datas := []struct {
		name string
		f    func(s *MockService) (err error)
	}{
		{"normal", testNornmal},
		{"initFailed", testInitFailed},
		{"startFailed", testStartFailed},
	}

	for _, v := range datas {
		t.Run(v.name, func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			s := NewMockService(mc)
			s.EXPECT().Name().AnyTimes().Return("test-svc")
			except := v.f(s)

			result := runApp(
				t, app.CreateApp(s),
				except == nil,
			)

			if except != result {
				t.Fatal("return value not equal",
					except, result,
				)
			}
		})
	}
}

func runApp(t *testing.T, a *app.Application, signFlag bool) (err error) {
	mu := sync.Mutex{}
	mu.Lock()
	signalCond := sync.NewCond(&mu)

	app.WaitStopSignal = func() {
		t.Log("wait for stop signal")
		signalCond.Wait()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		t.Log("run app...")
		err = a.Run()
		t.Log("finish")
		wg.Done()
	}()
	runtime.Gosched()
	time.Sleep(time.Millisecond)
	if signFlag {
		t.Log("signal stop")
		signalCond.Signal()
	}
	t.Log("wait app stop")
	wg.Wait()
	return
}

func testNornmal(s *MockService) (err error) {
	s.EXPECT().Init().Return(nil)
	s.EXPECT().Start().Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().Finish()
	return
}

func testInitFailed(s *MockService) (err error) {
	err = errors.New("init failed")
	s.EXPECT().Init().Return(err)
	return
}

func testStartFailed(s *MockService) (err error) {
	err = errors.New("start failed")
	s.EXPECT().Init().Return(nil)
	s.EXPECT().Start().Return(err)
	s.EXPECT().Finish()
	return
}
