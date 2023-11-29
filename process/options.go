package process

import (
	"context"
	"io"

	"github.com/walleframe/walle/process/message"
	"github.com/walleframe/walle/process/packet"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/atomic"
)

// InnerOption use for process
//
//go:generate gogen option -n InnerOption -f Inner -o option.inner.go
func walleProcessInner() interface{} {
	return map[string]interface{}{
		// Output: write interface(net.Conn)
		"Output": io.Writer(nil),
		// Specify Real Context
		"ContextPool": ContextPool(WrapContextPool),
		// process context parent
		"ParentCtx": context.Context(context.Background()),
		// Sequence number
		"Sequence": AtomicNumber(&atomic.Int64{}),
		// load number interface
		"Load": AtomicNumber(&atomic.Int64{}),
		// bind data
		"BindData": interface{}(nil),
		// process router.
		"Router": Router(GetRouter()),
	}
}

// ProcessOption process option
//
//go:generate gogen option -n ProcessOption -o option.process.go
func walleProcessOption() interface{} {
	return map[string]interface{}{
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.GetLogicLogger()),
		// frame log
		"FrameLogger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
		// packet pool
		"PacketPool": packet.Pool(packet.GetPool()),
		// packet wraper
		"PacketWraper": packet.ProtocolWraper(packet.GetProtocolWraper()),
		// packet encoder
		"PacketEncode": packet.Encoder(packet.GetEncoder()),
		// packet codec
		"PacketCodec": packet.Codec(packet.GetCodec()),
		// message codec
		"MsgCodec": message.Codec(message.WalleCodec),
		// dispatch packet data filter
		"DispatchDataFilter": DataDispatcherFilter(DefaultDataFilter),
		// dispatch packet struct filter
		"DispatchPacketFilter": PacketDispatcherFilter(DefaultPacketFilter),
		// load limit. return true to ignore packet.
		"LoadLimitFilter": func(req interface{}, count AtomicNumber) bool {
			return false
		},
	}
}
