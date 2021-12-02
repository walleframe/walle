package main

import (
	"fmt"

	"github.com/aggronmagi/walle/app"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	server "github.com/aggronmagi/walle/net/gotcp"
)

var (
	port int = 8080
)

func main() {

	r := &process.MixRouter{}
	r.Method("f1", rpcServerWrap(func(ctx server.SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M + rq.N
		rs.V2 = rq.M - rq.N
		// ctx.Logger().Debug7("f1", zap.Any("rs", rs))
		return
	}))
	count := atomic.Int32{}
	r.Method("f2", rpcServerWrap(func(ctx server.SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M
		rs.V2 = rq.N
		ctx.Logger().New("rpc").Debug("f2", zap.Any("rs", rs), zap.Int32("count", count.Inc()))
		return
	}))
	r.Method("f3", rpcServerWrap(func(ctx server.SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		err = packet.NewError(1000, "custom error")
		ctx.Logger().New("rpc").Debug("f3", zap.Any("rs", rs), zap.Error(err))
		return
	}))
	runServer(
		server.WithRouter(r),
		server.WithProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
			// process.WithDispatchDataFilter(func(data []byte, next process.PacketDispatcherFunc) (err error) {
			// 	go next(data)
			// 	return
			// }),
		),
		server.WithNewSession(func(in server.Session) (server.Session, error) {
			return in, nil
		}),
	)
}

func runServer(opt ...server.ServerOption) {
	opt = append(opt,
		server.WithAddr(fmt.Sprintf("localhost:%d", port)),
	)
	app.CreateApp(server.NewService("gnet", opt...)).Run()
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

func rpcServerWrap(f func(ctx server.SessionContext, rq *rpcRQ, rs *rpcRS) (err error)) func(ctx process.Context) {
	return func(c process.Context) {
		ctx := c.(server.SessionContext)
		in := ctx.GetRequestPacket()
		writeRespond := func(body interface{}) {
			out, err := ctx.NewResponse(in, body, nil)
			if err != nil {
				c.Logger().New("wrap").Error("new rpc respond failed", zap.Error(err))
				return
			}
			err = ctx.WritePacket(ctx, out)
			if err != nil {
				c.Logger().New("wrap").Error("write respond failed", zap.Error(err))
			}
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
