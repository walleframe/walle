package mempool

type MemoryPool interface {
	Alloc(size uint32) []byte
	Free([]byte)
}

var None = &noneMemPool{}

type noneMemPool struct{}

func (m *noneMemPool) Alloc(size uint32) (buf []byte) {
	buf = make([]byte, size)
	return
}

func (m *noneMemPool) Free(buf []byte) {
}
