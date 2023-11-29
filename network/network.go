package network

import (
	"context"
	"io"

	"github.com/walleframe/walle/network/rpc"
	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/process/metadata"
)

//go:generate mockgen -source network.go -destination ../testpkg/mock_network/network.go

type Caller interface {
	Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *rpc.CallOptions) (err error)
	AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af process.RouterFunc, opts *rpc.AsyncCallOptions) (err error)
	Notify(ctx context.Context, uri interface{}, rq interface{}, opts *rpc.NoticeOptions) (err error)
}

type CallerResponser interface {
	Write(ctx context.Context, payload interface{}, md metadata.MD)
}

type Link interface {
	// network write or close
	io.WriteCloser
	//
	Caller
}

type Server interface {
	Broadcast(uri interface{}, msg interface{}, md metadata.MD) error
	BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, md metadata.MD) error
	ForEach(f func(Session))
}

type Session interface {
	Link
	// GetConn get raw conn(net.Conn,websocket.Conn...)
	GetConn() interface{}
	// GetServer get raw server(*WsServer,*TcpServer...)
	GetServer() Server

	// WithSessionValue wrap context.WithValue
	WithSessionValue(key, value interface{})

	// SessionValue wrap context.Context.Value
	SessionValue(key interface{}) interface{}
	// AddCloseSessionFunc session close notify
	AddCloseSessionFunc(f func(sess Session))
}

type SessionContext interface {
	Session
	process.Context
}

type WriteMethod int8

const (
	WriteAsync WriteMethod = iota
	WriteImmediately
)

type Client interface {
	Link

	AddCloseClientFunc(f func(sess Client))
}

type ClientContext interface {
	process.Context
	Link
}
