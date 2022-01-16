package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aggronmagi/walle/network/gnet"
	"github.com/aggronmagi/walle/testpkg/wpb"
	"github.com/aggronmagi/walle/util"
	"github.com/aggronmagi/walle/zaplog"
)

func main() {
	zaplog.SetFrameLogger(zaplog.NoopLogger)
	zaplog.SetLogicLogger(zaplog.NoopLogger)
	// zaplog.SetFrameLogger(zaplog.GetLogicLogger())

	cli, err := gnet.NewClient(
		gnet.WithClientOptionsAddr(fmt.Sprintf("localhost:%d", 12345)),
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
