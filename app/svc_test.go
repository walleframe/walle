package app_test

import (
	"errors"
	"fmt"

	"testing"

	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/testpkg/mock_app"

	"github.com/golang/mock/gomock"
	"go.uber.org/atomic"
)

func TestFuncService(t *testing.T) {
	mc := gomock.NewController(t)
	check := mock_app.NewMockService(mc)
	check.EXPECT().Name().AnyTimes().Return("test-svc")
	testNornmal(check)
	svc := app.FuncService(
		app.WithName("TestFuncService"),
		app.WithInit(func(s app.Stoper) (err error) {
			return check.Init(s)
		}),
		app.WithFinish(func() {
			check.Finish()
		}),
		app.WithStart(func(s app.Stoper) (err error) {
			return check.Start(s)
		}),
		app.WithStop(func() {
			check.Stop()
		}),
	)
	ret := runApp(t, app.CreateApp(svc), true)
	if ret != nil {
		t.Fatal(ret)
	}
}

func TestTeeService_Normal(t *testing.T) {
	datas := []struct {
		name string
		num  int
		f    func(t *testing.T, mc *gomock.Controller, num int) (svcs []app.Service, ret error, signFlag bool)
	}{
		{"normal", 0, testTeeNomal},
		{"normal", 1, testTeeNomal},
		{"normal", 2, testTeeNomal},
		{"normal", 10, testTeeNomal},
	}
	for _, v := range datas {
		t.Run(fmt.Sprintf("%s - %d", v.name, v.num), func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			svcs, ret, flag := v.f(t, mc, v.num)
			check := runApp(t, app.CreateApp(app.TeeService(svcs...)), flag)
			if ret != check {
				t.Fatal(ret)
			}
		})
	}

}

func testTeeNomal(t *testing.T, mc *gomock.Controller, num int) (svcs []app.Service, ret error, signFlag bool) {
	init := atomic.Int32{}
	start := atomic.Int32{}
	for i := 0; i < num; i++ {
		index := int32(i)
		svc := mock_app.NewMockService(mc)
		svc.EXPECT().Name().AnyTimes().Return(fmt.Sprintf("test-%d", i))
		svc.EXPECT().Init(gomock.Any()).DoAndReturn(func(s app.Stoper) error {
			ret := init.Add(1)
			if ret != index+1 {
				t.Error("start sequece invalid")
			}
			t.Log("action init ", index)
			return nil
		})
		svc.EXPECT().Start(gomock.Any()).DoAndReturn(func(s app.Stoper) error {
			ret := start.Add(1)
			if ret != index+1 {
				t.Error("start sequece invalid")
			}
			t.Log("action start ", index)
			return nil
		})
		svc.EXPECT().Stop().Do(func() {
			ret := start.Sub(1)
			if ret != index {
				t.Error("stop sequece invalid")
			}
			t.Log("action stop ", index)
		})
		svc.EXPECT().Finish().Do(func() {
			ret := init.Sub(1)
			if ret != index {
				t.Error("finish sequece invalid")
			}
			t.Log("action finish ", index)
		})
		svcs = append(svcs, svc)
	}
	signFlag = true
	return
}

func TestTeeService_Failed(t *testing.T) {
	datas := []struct {
		name  string // 测试名称
		num   int    // 测试生成服务数量
		init  int    // 服务初始化成功索引,-1没有初始化失败
		start int    // 服务启动成功索引,-1没有启动失败
	}{
		{"normal", 0, -1, -1},
		{"normal", 1, -1, -1},
		{"normal", 2, -1, -1},
		{"normal", 10, -1, -1},
		{"initFailed", 1, 0, -1},
		{"initFailed", 3, 1, -1},
		{"initFailed", 3, 2, -1},
		{"startFailed", 1, -1, 0},
		{"startFailed", 3, -1, 1},
		{"startFailed", 3, -1, 2},
	}
	for _, v := range datas {
		t.Run(fmt.Sprintf("%v", v), func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			svcs, ret, flag := testTeeCheck(t, mc, v.num, v.init, v.start)
			check := runApp(t, app.CreateApp(app.TeeService(svcs...)), flag)
			if check != ret && !errors.Is(check, ret) {
				t.Fatal("invalid result:", ret, check)
			}
		})
	}

}

func testTeeCheck(t *testing.T, mc *gomock.Controller, num, initFailedIdx, startFailedIdx int) (svcs []app.Service, ret error, signFlag bool) {
	init := atomic.Int32{}
	start := atomic.Int32{}
	initFailed := errors.New("init failed")
	startFailed := errors.New("start failed")
	var arrs []*mock_app.MockService
	for i := 0; i < num; i++ {
		svc := mock_app.NewMockService(mc)
		svc.EXPECT().Name().AnyTimes().Return(fmt.Sprintf("test-%d", i))
		arrs = append(arrs, svc)
	}
	for i, svc := range arrs {
		index := i
		t.Log("need exce run", index, "init")
		svc.EXPECT().Init(gomock.Any()).DoAndReturn(func(s app.Stoper) error {
			t.Log("exce run", index, "init")
			v := int(init.Add(1))
			if v != index+1 {
				t.Error("init sequece invalid", v, index+1)
			}
			if initFailedIdx >= 0 && initFailedIdx == index {
				return initFailed
			}
			return nil
		})
		if initFailedIdx >= 0 && index >= initFailedIdx {
			ret = initFailed
			break
		}
	}
	for i, svc := range arrs {
		index := i
		if initFailedIdx >= 0 {
			break
		}

		t.Log("need exce run", index, "start")
		svc.EXPECT().Start(gomock.Any()).DoAndReturn(func(s app.Stoper) error {
			t.Log("exce run", index, "start")
			v := int(start.Add(1))
			if v != index+1 {
				t.Error("start sequece invalid", v, index+1)
			}
			if startFailedIdx >= 0 && startFailedIdx == int(index) {
				return startFailed
			}
			return nil
		})
		if startFailedIdx >= 0 && index >= startFailedIdx {
			ret = startFailed
			break
		}
	}
	for i := len(arrs) - 1; i >= 0; i-- {
		svc := arrs[i]
		index := i
		if initFailedIdx >= 0 {
			break
		}
		if startFailedIdx >= 0 && index >= startFailedIdx {
			continue
		}
		t.Log("need exce run", index, "stop")
		svc.EXPECT().Stop().Do(func() {
			t.Log("exce run", index, "stop")
		})
	}
	for i := len(arrs) - 1; i >= 0; i-- {
		svc := arrs[i]
		index := i
		// if initFailedIdx >= 0 {
		// 	break
		// }
		if initFailedIdx >= 0 && index >= initFailedIdx {
			continue
		}
		t.Log("need exce run", index, "finish")
		svc.EXPECT().Finish().Do(func() {
			t.Log("exce run", index, "finish")
		})
	}
	if ret == nil {
		signFlag = true
	}

	for _, svc := range arrs {
		svcs = append(svcs, svc)
	}
	if len(svcs) != num {
		t.Error("service count not match", num, len(svcs))
	}
	return
}
