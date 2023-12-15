package app_test

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/testpkg/mock_app"
)

// This example passes a context with a signal to tell a blocking function that
// it should abandon its work after a signal is received.
func TestApplicationRun(t *testing.T) {
	datas := []struct {
		name string
		f    func(s *mock_app.MockService) (err error)
	}{
		{"normal", testNornmal},
		{"initFailed", testInitFailed},
		{"startFailed", testStartFailed},
	}

	for _, v := range datas {
		t.Run(v.name, func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			s := mock_app.NewMockService(mc)
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
		a.Stop()
	}
	t.Log("wait app stop")
	wg.Wait()
	return
}

func testNornmal(s *mock_app.MockService) (err error) {
	s.EXPECT().Init(gomock.Any()).Return(nil)
	s.EXPECT().Start(gomock.Any()).Return(nil)
	s.EXPECT().Stop()
	s.EXPECT().Finish()
	return
}

func testInitFailed(s *mock_app.MockService) (err error) {
	err = errors.New("init failed")
	s.EXPECT().Init(gomock.Any()).Return(err)
	return
}

func testStartFailed(s *mock_app.MockService) (err error) {
	err = errors.New("start failed")
	s.EXPECT().Init(gomock.Any()).Return(nil)
	s.EXPECT().Start(gomock.Any()).Return(err)
	s.EXPECT().Finish()
	return
}

func TestMultiClose(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()
	s1 := mock_app.NewMockService(mc)
	s1.EXPECT().Init(gomock.Any()).Do(func(s app.Stoper) error {
		t.Log("s1 init")
		go func() {
			t.Log("s1 stop")
			s.Stop()
		}()
		runtime.Gosched()
		time.Sleep(time.Microsecond)
		return nil
	})
	s1.EXPECT().Finish()
	exit := make(chan struct{})
	go func() {
		app.CreateApp(s1).Run()
		exit <- struct{}{}
	}()
	runtime.Gosched()
	time.Sleep(time.Millisecond)
	select {
	case <-time.After(time.Microsecond * 10):
		t.Fatal("timeout, not stop")
	case <-exit:
		t.Log("stop ok")
	}

}
