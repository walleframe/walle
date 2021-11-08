package packet

import "sync"

// PacketPool pool
type PacketPool interface {
	Pop() *Packet
	Push(*Packet)
}

var DefaultPacketPool PacketPool = SyncPool

type NativePool struct{}

func (p *NativePool) Pop() *Packet {
	return new(Packet)
}

func (p *NativePool) Push(*Packet) {
}

var SyncPool PacketPool = &syncPool{}

type syncPool struct {
	sync.Pool
}

func (p *syncPool) Pop() *Packet {
	if p.Pool.New == nil {
		p.Pool.New = func() interface{} {
			return new(Packet)
		}
	}
	return p.Pool.Get().(*Packet)
}

func (p *syncPool) Push(v *Packet) {
	p.Pool.Put(v)
}
