package ws

import (
	"context"
	"net/http"

	"github.com/aggronmagi/walle/net/process"
	"github.com/gorilla/websocket"
)

func NewClient(addr string, head http.Header, inner *process.InnerOptions, opts ...process.ProcessOption) (_ Client, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, head)
	if err != nil {
		return
	}
	cli := &WsSession{
		Process: process.NewProcess(
			process.NewInnerOptions(),
			process.NewProcessOptions(opts...),
		),
		conn: conn,
	}
	cli.Process.Inner.ApplyOption(
		process.WithInnerOptionsOutput(cli),
		process.WithInnerOptionsBindData(cli),
		process.WithInnerOptionsNewContext(cli.newContext),
	)
	cli.writeMethod = WriteAsync
	cli.ctx = context.Background()
	cli.cancel = func() {}
	go cli.Run()
	return cli, nil
}
