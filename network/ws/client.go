package ws

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/walleframe/walle/network/rpc"
	"github.com/walleframe/walle/process"
)

// NewClientEx 创建客户端。NOTE: websocket socket 客户端不支持自动重连.仅用于测试
// inner *process.InnerOptions 选项应该由上层ClientProxy去决定如何设置。
// svr 内部应该设置链接相关的参数。比如读写超时，如何发送数据
// opts 业务方决定
func NewClientEx(addr string, head http.Header,
	inner *process.InnerOptions,
	svr *ServerOptions, // TODO 客户端独立选项配置
) (cli *WsSession, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, head)
	if err != nil {
		return
	}
	cli = &WsSession{
		RPCProcess: rpc.NewRPCProcess(
			inner,
			process.NewProcessOptions(svr.ProcessOptions...),
		),
		conn:   conn,
		logger: svr.FrameLogger,
	}
	cli.Process.Inner.ApplyOption(
		process.WithInnerOptionOutput(cli),
		process.WithInnerOptionBindData(cli),
		process.WithInnerOptionContextPool(GoServerContextPool),
	)
	cli.opts = svr // TODO 客户端独立配置转换
	cli.ctx = context.Background()
	cli.cancel = func() {}
	go cli.Run()
	return cli, nil
}

// NewClient 新建客户端。NOTE: websocket socket 客户端不支持自动重连.仅用于测试
func NewClient(addr string, head http.Header, opts ...process.ProcessOption) (*WsSession, error) {
	return NewClientEx(addr, head, process.NewInnerOptions(), NewServerOptions(
		WithProcessOptions(opts...),
	))
}
