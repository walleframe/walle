package iface

import (
	"context"
	"io"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/net/process"
)

type Caller interface {
	Call(ctx context.Context, uri interface{}, rq, rs interface{}, opts *process.CallOptions) (err error)
	AsyncCall(ctx context.Context, uri interface{}, rq interface{}, af process.RouterFunc, opts *process.AsyncCallOptions) (err error)
	Notify(ctx context.Context, uri interface{}, rq interface{}, opts *process.NoticeOptions) (err error)
}

type Link interface {
	// network write or close
	io.WriteCloser
	//
	Caller

	// process wrap
	NewPacket(cmd packet.Command, uri, rq interface{}, md []process.MetadataOption, errflag ...bool) (req *packet.Packet, err error)
	NewResponse(in *packet.Packet, body interface{}, md []process.MetadataOption) (rsp *packet.Packet, err error)
	MarshalPacket(req *packet.Packet) (data []byte, err error)
	WritePacket(ctx context.Context, req *packet.Packet) (err error)
}

type Server interface {
	Broadcast(uri interface{}, msg interface{}, meta ...process.MetadataOption) error
	BroadcastFilter(filter func(Session) bool, uri interface{}, msg interface{}, meta ...process.MetadataOption) error
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
	// AddCloseFunc session close notify
	AddCloseFunc(f func(sess Session))
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
