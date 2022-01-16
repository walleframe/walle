package packet

import (
	"sync"

	"github.com/aggronmagi/walle/process/metadata"
)

// noopPacketPool not cache packet
type noopPacketPool struct{}

func (p noopPacketPool) Get() interface{} {
	return NewPacket()
}

func (p noopPacketPool) Put(interface{}) {
}

// NoPacketPool not use packet pools
var NoPacketPool Pool = noopPacketPool{}

type syncPacketPool struct {
	sync.Pool
}

func (p *syncPacketPool) Get() interface{} {
	return p.Pool.Get()
}

func (p *syncPacketPool) Put(x interface{}) {
	pb, ok := x.(*Packet)
	if !ok {
		return
	}

	// if cap(pb.cache) > 1024 {
	// 	return
	// }

	//mp := mempool.Pool()
	//mp.Free(pb.payload) // free packet payload
	//mp.Free(pb.cache)   // free packet data
	pb.payload = nil
	pb.cache = nil
	pb.flag = 0
	pb.reservd = 0
	pb.msgID = 0
	pb.msgURI = ""
	pb.sessionID = 0
	pb.metadata = make(metadata.MD)

	p.Pool.Put(x)
}

// SyncPacketPool use sync.Pool
var SyncPacketPool Pool = &syncPacketPool{
	Pool: sync.Pool{
		New: func() interface{} {
			return NewPacket()
		},
	},
}

var defaultPacketPool Pool = SyncPacketPool

// GetPacketPool get default packet pool
func GetPool() Pool {
	return defaultPacketPool
}

// SetPacketPool set default packet pool
func SetPacketPool(p Pool) {
	defaultPacketPool = p
}
