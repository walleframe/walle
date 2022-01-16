package main

import (
	"fmt"

	"github.com/aggronmagi/walle/network/gnet"
	"github.com/aggronmagi/walle/process"
	"github.com/aggronmagi/walle/testpkg/wpb"
	"github.com/aggronmagi/walle/zaplog"
)

func main() {
	p := 12345
	fmt.Println("port:", p)

	zaplog.SetFrameLogger(zaplog.NoopLogger)
	zaplog.SetLogicLogger(zaplog.NoopLogger)
	// zaplog.SetFrameLogger(zaplog.GetLogicLogger())

	wpb.RegisterWSvcService(process.GetRouter(), &wpb.WPBSvc{})
	svc := gnet.NewServer(
		gnet.WithReuseReadBuffer(true),
	)
	svc.Run(fmt.Sprintf(":%d", p))
}
