package gnet

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aggronmagi/walle/network/rpc"
	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/testpkg/wpb"
	"github.com/aggronmagi/walle/util"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"github.com/stretchr/testify/assert"
)

var (
	bp int
)

func TestMain(m *testing.M) {
	zaplog.SetFrameLogger(zaplog.NoopLogger)
	zaplog.SetLogicLogger(zaplog.NoopLogger)
	// zaplog.SetFrameLogger(zaplog.GetLogicLogger())
	p, err := util.GetFreePort()
	if err != nil {
		panic(err)
	}
	bp = p

	fmt.Println("port:", p)

	wpb.RegisterWSvcService(process.GetRouter(), &wpb.WPBSvc{})
	svc := NewServer(
		WithAddr(fmt.Sprintf(":%d", p)),
		WithReuseReadBuffer(true),
	)
	go svc.Run("")
	runtime.Gosched()
	time.Sleep(time.Millisecond * 50)
	//svc.R

	m.Run()
}

func TestGoTCPClient(t *testing.T) {
	cli, err := NewClient(
		WithClientOptionsAddr(fmt.Sprintf("localhost:%d", bp)),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	wcli := wpb.NewWSvcClient(cli)

	addRs, err := wcli.Add(context.Background(), &wpb.AddRq{
		Params: []int64{1, 5},
	})
	assert.Nil(t, err, "call rpc add error")
	if err != nil {
		return
	}
	assert.EqualValues(t, addRs.Value, 6, "rpc add return value")

	// call f2
	mulRs, err := wcli.Mul(context.Background(), &wpb.MulRq{
		A: 100,
		B: 5,
	})
	assert.Nil(t, err, "call rpc mul error")
	assert.EqualValues(t, 500, mulRs.R, "rpc mul return value")

	// call f3
	reRs, err := wcli.Re(ctx, &wpb.AddRq{})
	assert.NotNil(t, err, "re return error")
	assert.Nil(t, reRs, "re return value")

	wg := sync.WaitGroup{}
	wg.Add(1)
	err = wcli.AddAsync(ctx, &wpb.AddRq{Params: []int64{100, 90}}, func(ctx process.Context, rs *wpb.AddRs, err error) {
		assert.Nil(t, err, "async add error")
		assert.EqualValues(t, 190, rs.Value, "async add result")
		wg.Done()
	}, rpc.WithAsyncCallOptionsTimeout(time.Second))
	wg.Wait()

	err = cli.Close()
	assert.Nil(t, err, "close client error")
	time.Sleep(time.Millisecond * 100)
}

func BenchmarkGoTCPClient(b *testing.B) {
	cli, err := NewClient(
		WithClientOptionsAddr(fmt.Sprintf("localhost:%d", bp)),
	)
	if err != nil {
		panic(err)
	}
	wcli := wpb.NewWSvcClient(cli)

	b.Run("Call", func(b *testing.B) {
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			_, err := wcli.Add(context.Background(), &wpb.AddRq{}) //rpc.WithCallOptionsTimeout(time.Second),

			if err != nil {
				b.Error("stop fatal", err)
				b.Fatal(k, err)
			}
		}
	})

	b.Run("CallNoRet", func(b *testing.B) {
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			err := wcli.CallOneWay(context.Background(), &wpb.AddRq{}) //rpc.WithCallOptionsTimeout(time.Second),
			if err != nil {
				b.Error("stop fatal", err)
				b.Fatal(k, err)
			}
		}
	})

	b.Run("AsyncCall", func(b *testing.B) {
		wg := sync.WaitGroup{}
		// fmt.Println("--->", b.N)
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			wg.Add(1)
			err := wcli.AddAsync(context.Background(), &wpb.AddRq{},
				func(ctx process.Context, rs *wpb.AddRs, err error) {
					wg.Done()
				},
				//rpc.WithAsyncCallOptionsTimeout(time.Second),
			)
			if err != nil {
				b.Error("stop fatal", err)
				b.Fatal(k, err)
			}
		}
		//b.Log("wait", num.Load())
		wg.Wait()
	})
	// notify 必须在最后测试，否则会影响测试结果
	// 因为notify只发送消息，未等待服务端处理完成，后续的测试会被notify影响
	b.Run("Notify", func(b *testing.B) {
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			wcli.NotifyFunc(context.Background(), &wpb.AddRq{}) //rpc.WithNoticeOptionsTimeout(time.Second),

		}
	})

}
