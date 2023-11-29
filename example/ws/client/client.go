package main

import (
	"context"
	"fmt"
	"time"

	"github.com/walleframe/walle/network/ws"
	"github.com/walleframe/walle/testpkg/wpb"
	"github.com/walleframe/walle/util"
	"github.com/walleframe/walle/zaplog"
)

func main() {
	zaplog.SetFrameLogger(zaplog.NoopLogger)
	zaplog.SetLogicLogger(zaplog.NoopLogger)
	// zaplog.SetFrameLogger(zaplog.GetLogicLogger())

	cli, err := ws.NewClient(
		fmt.Sprintf("ws://localhost:%d/ws", 12345), nil,
	)
	if err != nil {
		util.PanicIfError(err)
	}
	time.Sleep(time.Second)

	wcli := wpb.NewWSvcClient(cli)
	ctx := context.Background()

	rs, err := wcli.Add(ctx, &wpb.AddRq{Params: []int64{100}})
	fmt.Println(rs, err)
}
