package main

import (
	"fmt"

	"github.com/walleframe/walle/network/gotcp"
	"github.com/walleframe/walle/network/kcp"
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
	svc := gotcp.NewServer(
		gotcp.WithReuseReadBuffer(true),
		gotcp.WithListen(kcp.GoTCPServerOptionListen),
	)
	svc.Run(fmt.Sprintf(":%d", p))
}
