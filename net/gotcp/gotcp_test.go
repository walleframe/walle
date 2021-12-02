package gotcp

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aggronmagi/walle/app"
	"github.com/aggronmagi/walle/internal/util/test"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var (
	client Client
	bp     int
)

func TestMain(m *testing.M) {
	p, err := util.GetFreePort()
	if err != nil {
		panic(err)
	}
	bp = p

	fmt.Println("port:", p)

	r := &process.MixRouter{}
	r.Method("f1", rpcServerRF(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M + rq.N
		rs.V2 = rq.M - rq.N
		return
	}))
	r.Method("f2", rpcServerRF(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M
		rs.V2 = rq.N
		return
	}))
	r.Method("f3", rpcServerRF(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		err = packet.NewError(1000, "custom error")
		return
	}))

	r.Method("ntf", func(c process.Context) {
		return
	})

	defer runServer(p,
		WithRouter(r),
		WithProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
		),
		WithHeartbeat(time.Second),
	)()

	// cli, err := NewClient(
	// 	NewClientOptions(
	// 		WithClientOptionsAddr(fmt.Sprintf("localhost:%d", p)),
	// 	),
	// 	process.WithMsgCodec(process.MessageCodecJSON),
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// client = cli

	// err = client.Call(context.Background(), "f1", &rpcRQ{}, &rpcRS{}, process.NewCallOptions())
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("rpc call success")

	m.Run()
}

func runServer(port int, opt ...ServerOption) (stop func()) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	app.WaitStopSignal = func() {
		wg.Wait()
	}
	stop = func() {
		wg.Done()
	}
	opt = append(opt,
		WithAddr(fmt.Sprintf("localhost:%d", port)),
	)
	go app.CreateApp(NewService("gnet", opt...)).Run()
	runtime.Gosched()
	time.Sleep(time.Microsecond)
	return
}

type rpcRQ struct {
	M int `json:"m"`
	N int `json:"n"`
}

type rpcRS struct {
	V1 int `json:"v1"`
	V2 int `json:"v2"`
}

func rpcServerRF(f func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error)) func(ctx process.Context) {
	return func(c process.Context) {
		ctx := c.(SessionContext)
		in := ctx.GetRequestPacket()
		writeRespond := func(body interface{}) {
			out, err := ctx.NewResponse(in, body, nil)
			if err != nil {
				c.Logger().New("wrap").Error("new rpc respond failed", zap.Error(err))
				return
			}
			ctx.WritePacket(ctx, out)
		}

		rq := &rpcRQ{}
		rs := &rpcRS{}
		err := ctx.Bind(rq)
		if err != nil {
			writeRespond(err)
			return
		}
		err = f(ctx, rq, rs)
		if err != nil {
			writeRespond(err)
			return
		}
		writeRespond(rs)
		return
	}
}

func TestGoTCP(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()
	f := test.NewMockFuncCall(mc)

	p, err := util.GetFreePort()
	if err != nil {
		panic(err)
	}
	r := &process.MixRouter{}
	r.Method("f1", rpcServerRF(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M + rq.N
		rs.V2 = rq.M - rq.N
		return
	}))
	r.Method("f2", rpcServerRF(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M
		rs.V2 = rq.N
		return
	}))
	r.Method("f3", rpcServerRF(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		err = packet.NewError(1000, "custom error")
		return
	}))
	stopServer := runServer(p,
		WithRouter(r),
		WithProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
		),
		WithHeartbeat(time.Second),
		WithNewSession(func(in Session) (Session, error) {
			f.EXPECT().Call("close notify", in)
			in.AddCloseFunc(func(sess Session) {
				f.Call("close notify", sess)
			})
			return in, nil
		}),
	)

	cli, err := NewClient(
		WithClientOptionsAddr(fmt.Sprintf("localhost:%d", p)),
		WithClientOptionsProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	rq := &rpcRQ{108, 72}
	rs := &rpcRS{}

	ctx := context.Background()

	// call f1
	err = cli.Call(ctx, "f1", rq, rs, process.NewCallOptions(
		process.WithCallOptionsTimeout(time.Second),
	))
	assert.Nil(t, err, "call rpc f1 error")
	if err != nil {
		return
	}
	assert.Equal(t, rq.M+rq.N, rs.V1, "rpc f1 return v1 value")
	assert.Equal(t, rq.M-rq.N, rs.V2, "rpc f1 return v2 value")

	for k := 0; k < 2000; k++ {
		err = cli.Call(ctx, "f1", &rpcRQ{}, &rpcRS{}, process.NewCallOptions(
			process.WithCallOptionsTimeout(time.Second*10),
		))
		assert.Nil(t, err, "call rpc f1 error %d", k)
		if err != nil {
			return
		}
	}

	// call f2
	err = cli.Call(ctx, "f2", rq, rs, process.NewCallOptions(
		process.WithCallOptionsTimeout(time.Second),
	))
	assert.Nil(t, err, "call rpc f2 error")
	assert.Equal(t, rq.M, rs.V1, "rpc f2 return v1 value")
	assert.Equal(t, rq.N, rs.V2, "rpc f2 return v2 value")

	// call f3
	err = cli.Call(ctx, "f3", rq, rs, process.NewCallOptions(
		process.WithCallOptionsTimeout(time.Second),
	))
	assert.NotNil(t, err, "f3 return error")

	f.EXPECT().Call("async")
	cli.AsyncCall(ctx, "f2", rq, func(c process.Context) {
		f.Call("async")
	}, process.NewAsyncCallOptions(
		process.WithAsyncCallOptionsWaitFilter(func(await func()) {
			await()
		}),
	))
	time.Sleep(time.Millisecond * 100)

	err = cli.Close()
	assert.Nil(t, err, "close client error")
	time.Sleep(time.Millisecond * 100)
	stopServer()
	// time.Sleep(time.Second)
}

func BenchmarkWsClient(b *testing.B) {
	cli, err := NewClient(
		WithClientOptionsAddr(fmt.Sprintf("localhost:%d", bp)),
		WithClientOptionsProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
		),
	)
	if err != nil {
		panic(err)
	}

	b.Run("Call", func(b *testing.B) {
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			err := cli.Call(context.Background(), "f1", &rpcRQ{}, &rpcRS{}, process.NewCallOptions(
				process.WithCallOptionsTimeout(time.Second),
			))
			if err != nil {
				b.Error("stop fatal", err)
				b.Fatal(k, err)
			}
		}
	})

	b.Run("AsyncCall", func(b *testing.B) {
		wg := sync.WaitGroup{}
		num := atomic.Int32{}
		b.ResetTimer()

		for k := 0; k < b.N; k++ {
			wg.Add(1)
			num.Inc()
			cli.AsyncCall(context.Background(), "f2", &rpcRQ{},
				func(c process.Context) {
					num.Dec()
					wg.Done()
				}, process.NewAsyncCallOptions(
					process.WithAsyncCallOptionsTimeout(time.Second),
				))
		}
		b.Log("wait", num.Load())
		wg.Wait()
	})

	b.Run("Notify", func(b *testing.B) {
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			cli.Notify(context.Background(), "ntf", &rpcRQ{}, process.NewNoticeOptions(
				process.WithNoticeOptionsTimeout(time.Second),
			))
		}
	})
}
