package process

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/zaplog"
	"go.uber.org/atomic"
)

// MetadataOption set request metadata
type MetadataOption func(p *packet.Packet)

func MetadataString(key, val string) MetadataOption {
	return func(p *packet.Packet) {
		p.Metadata[key] = val
	}
}

func MetadataUint64(key string, val uint64) MetadataOption {
	return func(p *packet.Packet) {
		p.Metadata[key] = strconv.FormatUint(val, 10)
	}
}

func MetadataInt64(key string, val int64) MetadataOption {
	return func(p *packet.Packet) {
		p.Metadata[key] = strconv.FormatInt(val, 10)
	}
}

// AsyncResponseFilter 异步请回回复调用
type AsyncResponseFilter func(ctx Context, req, rsp *packet.Packet)

// CallOption rpc call options
//go:generate gogen option -n CallOption -f Call -o option.call.go
func walleCallOption() interface{} {
	return map[string]interface{}{
		// rpc call timeout
		"Timeout": time.Duration(0),
		// metadata
		"Metadata": []MetadataOption{},
	}
}

// CallOption rpc call options
//go:generate gogen option -n AsyncCallOption -f Async -o option.async.go
func walleAsyncCallOption() interface{} {
	return map[string]interface{}{
		// rpc call timeout
		"Timeout": time.Duration(0),
		// metadata
		"Metadata": []MetadataOption{},
		// response filter. NOTE: req only valid in Filter func.
		"ResponseFilter": AsyncResponseFilter(func(ctx Context, req, rsp *packet.Packet) {
			ctx.Next(ctx)
		}),
		"WaitFilter": func(await func()) {
			go await()
		},
	}
}

// NoticeOption oneway rpc
//go:generate gogen option -n NoticeOption -f Notice -o option.notify.go
func walleNoticeCallOption() interface{} {
	return map[string]interface{}{
		// send message timeout
		"Timeout": time.Duration(0),
		// metadata
		"Metadata": []MetadataOption{},
	}
}

// InnerOption use for process
//go:generate gogen option -n InnerOption -f Inner -o option.inner.go
func walleProcessInner() interface{} {
	return map[string]interface{}{
		// Output: write interface(net.Conn)
		"Output": io.Writer(nil),
		// Specify Real Context
		"NewContext": func(ctx Context, ud interface{}) Context {
			return ctx
		},
		// process context parent
		"ParentCtx": context.Context(context.Background()),
		// Sequence number
		"Sequence": AtomicNumber(&atomic.Int64{}),
		// load number interface
		"Load": AtomicNumber(&atomic.Int64{}),
		// bind data
		"BindData": interface{}(nil),
		// process router.
		"Router": Router(nil),
	}
}

// ProcessOption process option
//go:generate gogen option -n ProcessOption -o option.process.go
func walleProcessOption() interface{} {
	return map[string]interface{}{
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.Logic),
		// frame log
		"FrameLogger":(*zaplog.Logger)(zaplog.Frame),
		// packet pool
		"PacketPool": packet.PacketPool(packet.DefaultPacketPool),
		// packet encoder
		"PacketEncode": PacketEncoder(&EmtpyPacketCoder{}),
		// packet codec
		"PacketCodec": PacketCodec(PacketCodecProtobuf),
		// message codec
		"MsgCodec": MessageCodec(MessageCodecProtobuf),
		// dispatch packet data filter
		"DispatchDataFilter": PacketDispatcherFilter(DefaultPacketFilter),
		// load limit. return true to ignore packet.
		"LoadLimitFilter": func(ctx Context, count int64, req *packet.Packet) bool {
			return false
		},
	}
}
