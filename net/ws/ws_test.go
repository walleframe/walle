package ws

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aggronmagi/walle/app"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	port int
)

func TestMain(m *testing.M) {
	p, err := util.GetFreePort()
	if err != nil {
		panic(err)
	}
	port = p
	m.Run()
}

func runServer(t *testing.T, opt ...ServerOption) (stop func()) {
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
		WithWsPath("/ws"),
	)
	go app.CreateApp(NewService("ws", opt...)).Run()
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
				c.Logger().Error3("new rpc respond failed", zap.Error(err))
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

func TestWs(t *testing.T) {
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
	defer runServer(t,
		WithRouter(r),
		WithProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
		),
	)()

	cli, err := NewClient(fmt.Sprintf("ws://localhost:%d/ws", port), nil, nil,
		process.WithMsgCodec(process.MessageCodecJSON),
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
}
