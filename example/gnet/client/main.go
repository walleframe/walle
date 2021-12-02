package main

import (
	"context"
	"fmt"
	"sync"

	"time"

	"github.com/aggronmagi/walle/net/process"
	"go.uber.org/atomic"

	. "github.com/aggronmagi/walle/net/gnet"
)

type rpcRQ struct {
	M int `json:"m"`
	N int `json:"n"`
}

type rpcRS struct {
	V1 int `json:"v1"`
	V2 int `json:"v2"`
}

func main() {
	recvCnt := atomic.Int64{}
	cli, err := NewClient(
		WithClientOptionsAddr(fmt.Sprintf("localhost:%d", 8080)),
		WithClientOptionsProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
			process.WithDispatchDataFilter(func(data []byte, next process.PacketDispatcherFunc) (err error) {
				recvCnt.Inc()
				return next(data)
			}),
		),
	)

	if err != nil {
		panic(err)
	}
	rq := &rpcRQ{108, 72}

	call := func(uri string) {
		ctx := context.Background()
		rs := &rpcRS{}
		err = cli.Call(ctx, uri, rq, rs, process.NewCallOptions(
			process.WithCallOptionsTimeout(time.Second),
		))
		fmt.Println("call rpc ", uri, err, rs)
	}
	call("f1")
	call("f3")
	call("f2")
	call("f1")

	n := 10000
	// zaplog.Frame = zaplog.NewLogger(zap.NewNop())
	// zaplog.Logic = zaplog.NewLogger(zap.NewNop())

	func() {
		start := time.Now()
		m := 0
		for k := 0; k < n; k++ {
			ctx := context.Background()
			rq.N = k
			rs := &rpcRS{}
			err = cli.Call(ctx, "f2", rq, rs, process.NewCallOptions(
				process.WithCallOptionsTimeout(time.Second),
			))
			if err != nil {
				fmt.Println("call rpc failed: ", k, err)
				break
			}
			m++
		}
		if m < 1 {
			m = 1
		}
		use := time.Now().Sub(start)
		fmt.Println("use:", use, "--", use/time.Duration(m), "allrecv:", recvCnt.Load())
	}()

	func() {
		recvCnt.Store(0)
		n = 10000
		start := time.Now()
		m := 0
		wg := sync.WaitGroup{}
		num := atomic.Int64{}
		wg.Add(n)
		num.Add(int64(n))
		errCnt := atomic.Int32{}
		rq.N = 1
		rs := &rpcRS{}
		for k := 0; k < n; k++ {
			rq.M = k
			ctx := context.Background()
			err = cli.AsyncCall(ctx, "f2", rq,
				func(c process.Context) {
					wg.Done()
					num.Dec()
					err = c.Bind(rs)
					if err != nil {
						errCnt.Add(1)
					}
				},
				// process.WithAsyncCallOptionsTimeout(time.Second),
				process.NewAsyncCallOptions(
					process.WithAsyncCallOptionsWaitFilter(func(await func()) {
						await()
					}),
				))
			if err != nil {
				fmt.Println("call rpc failed: ", k, err)
				break
			}
			m++
		}
		if m < 1 {
			m = 1
		}
		use := time.Now().Sub(start)
		fmt.Println("use:", use, "--", use/time.Duration(m), "send", n, "recv", n-int(num.Load()), errCnt.Load(),
			"recv", recvCnt.Load())
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			fmt.Println(recvCnt.Load())
			if recvCnt.Load() == int64(n) {
				break
			}
		}

		use = time.Now().Sub(start)
		fmt.Println("all use:", use, "err", errCnt.Load(), "left:", num.Load())
	}()

}
