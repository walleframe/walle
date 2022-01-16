package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aggronmagi/walle/network/gotcp"
	"github.com/aggronmagi/walle/network/kcp"
	"github.com/aggronmagi/walle/testpkg/wpb"
	"github.com/aggronmagi/walle/util"
	"github.com/aggronmagi/walle/zaplog"
)

func main() {
	zaplog.SetFrameLogger(zaplog.NoopLogger)
	zaplog.SetLogicLogger(zaplog.NoopLogger)
	// zaplog.SetFrameLogger(zaplog.GetLogicLogger())

	cli, err := gotcp.NewClient(
		gotcp.WithClientOptionsAddr(fmt.Sprintf("localhost:%d", 12345)),
		gotcp.WithClientOptionsDialer(kcp.GoTCPClientOptionDialer),
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
