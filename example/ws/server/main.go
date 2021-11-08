package main

import (
	"fmt"
	"net/http"

	"github.com/aggronmagi/walle/app"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
	"github.com/aggronmagi/walle/net/ws"
	"github.com/aggronmagi/walle/zaplog"
	"go.uber.org/zap"

	. "github.com/aggronmagi/walle/net/ws"
)

var (
	port int = 8080
)

func main() {
	log, err := zaplog.NewLoggerWithCfg(zaplog.DEBUG, zap.NewDevelopmentConfig(), zap.AddStacktrace(zap.WarnLevel))
	if err != nil {
		panic(err)
	}
	zaplog.Default = log

	r := &process.MixRouter{}
	r.Method("f1", rpcServerWrap(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M + rq.N
		rs.V2 = rq.M - rq.N
		ctx.Logger().Debug7("f1", zap.Any("rs", rs))
		return
	}))
	r.Method("f2", rpcServerWrap(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		rs.V1 = rq.M
		rs.V2 = rq.N
		ctx.Logger().Debug7("f2", zap.Any("rs", rs))
		return
	}))
	r.Method("f3", rpcServerWrap(func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error) {
		err = packet.NewError(1000, "custom error")
		ctx.Logger().Debug7("f3", zap.Any("rs", rs), zap.Error(err))
		return
	}))
	runServer(
		WithRouter(r),
		WithProcessOptions(
			process.WithMsgCodec(process.MessageCodecJSON),
		),
		WithNewSession(func(in ws.Session, r *http.Request) (ws.Session, error) {
			fmt.Println(r.Header)
			return in, nil
		}),
	)
}

func runServer(opt ...ServerOption) {
	opt = append(opt,
		WithAddr(fmt.Sprintf("localhost:%d", port)),
		WithWsPath("/ws"),
	)
	app.CreateApp(NewService("ws", opt...)).Run()
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

func rpcServerWrap(f func(ctx SessionContext, rq *rpcRQ, rs *rpcRS) (err error)) func(ctx process.Context) {
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
