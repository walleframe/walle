// Code generated by "gogen imake"; DO NOT EDIT.
// Exec: "gogen imake . -t=RPCProcess -r RPCProcess=RPCProcesser -o rpc_processer.go --merge"
// Version: 0.0.7

package rpc

import (
	context "context"

	process "github.com/walleframe/walle/process"
)

// RPCProcess 通用rpc处理流程封装 封装
type RPCProcesser interface {
	// OnReply rpc请求返回处理
	OnReply(in interface{}) (filter bool)
	// Call 同步rpc请求
	Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *CallOptions) (err error)
	// AsyncCall 异步RPC请求
	AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af process.RouterFunc, opts *AsyncCallOptions) (err error)
	// Notify 通知请求(one way)
	Notify(ctx context.Context, uri interface{}, rq interface{}, opts *NoticeOptions) (err error)
	// Clean session 清理（rpc请求等缓存清理）
	Clean()
}
