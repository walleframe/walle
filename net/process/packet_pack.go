package process

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/aggronmagi/walle/mempool"
)

// 网络分包,沾包处理.同时用作网络包缓存
// Network subcontracting and packet processing.
// At the same time, it is used as a network packet cache
type NetPackager interface {
	Reset()
	Read(r io.Reader) (pkgs net.Buffers, err error)
	Write(w io.Writer, data []byte) (n int, err error)
}

// NewNetPackager create
type NewNetPackager func(head, packetLimit int, pool mempool.MemoryPool, byteOrder binary.ByteOrder) NetPackager

// type assert
var _ NewNetPackager = NewPackNetPackager

type packNetPackager struct {
	head        []byte
	packetLimit int
	byteOrder   binary.ByteOrder
	pool        mempool.MemoryPool
	sendBuf     []byte
}

func NewPackNetPackager(head, packetLimit int, pool mempool.MemoryPool, byteOrder binary.ByteOrder) NetPackager {
	if head != 2 && head != 4 {
		return nil
	}
	return &packNetPackager{
		head:        make([]byte, head),
		packetLimit: packetLimit,
		pool:        pool,
		sendBuf:     pool.Alloc(uint32(packetLimit + 4)),
	}
}

func (p *packNetPackager) Reset() {
	p.head = make([]byte, len(p.head))
}

func (p *packNetPackager) Read(r io.Reader) (pkgs net.Buffers, err error) {
	_, err = io.ReadFull(r, p.head)
	if err != nil {
		return
	}
	size := uint32(0)
	switch len(p.head) {
	case 2:
		size = uint32(p.byteOrder.Uint16(p.head))
	case 4:
		size = p.byteOrder.Uint32(p.head)
	}

	if size > uint32(p.packetLimit) {
		err = fmt.Errorf("invalid packet size:%d limit:%d", size, p.packetLimit)
		return
	}
	buf := p.pool.Alloc(size)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		p.pool.Free(buf)
		return
	}
	pkgs = append(pkgs, buf)

	return
}

func (p *packNetPackager) Write(w io.Writer, data []byte) (n int, err error) {
	buf := p.sendBuf
	need := len(data) + len(p.head)
	// FIXME: limit send data size ?
	if len(buf) < need {
		p.pool.Free(p.sendBuf)
		p.sendBuf = p.pool.Alloc(uint32(need))
	}
	switch len(p.head) {
	case 2:
		p.byteOrder.PutUint16(buf, uint16(len(data)))
		copy(buf[2:], data)
	case 4:
		p.byteOrder.PutUint32(buf, uint32(len(data)))
		copy(buf[4:], data)
	}
	n, err = w.Write(buf)
	return
}
