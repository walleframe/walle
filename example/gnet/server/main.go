package main

import (
	"fmt"

	"github.com/walleframe/walle/network/gnet"
	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/testpkg/wpb"
	"github.com/walleframe/walle/zaplog"
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
