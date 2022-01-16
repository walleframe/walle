package util

import (
	"encoding/binary"
	"reflect"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// A Builder is used to efficiently build a string using Write methods.
// It minimizes memory copying. The zero value is ready to use.
// Do not copy a non-zero Builder.
// NOTE: this is copy from strings.Builder,And Add some useful functions.
type Builder struct {
	addr *Builder // of receiver, to detect copies by value
	buf  []byte
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input. noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime; see issues 23382 and 7921.
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
	// NOTE: this is copy from strings.Builder
	return unsafe.Pointer(uintptr(p))
}

func (b *Builder) copyCheck() {
	if b.addr == nil {
		// This hack works around a failing of Go's escape analysis
		// that was causing b to escape and be heap allocated.
		// See issue 23382.
		// TODO: once issue 7921 is fixed, this should be reverted to
		// just "b.addr = b".
		b.addr = (*Builder)(noescape(unsafe.Pointer(b)))
	} else if b.addr != b {
		panic("strings: illegal use of non-zero Builder copied by value")
	}
}

// String returns the accumulated string.
func (b *Builder) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *Builder) Cap() int { return cap(b.buf) }

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.addr = nil
	b.buf = nil
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (b *Builder) grow(n int) {
	buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
	copy(buf, b.buf)
	b.buf = buf
}

func (b *Builder) SetBuf(buf []byte) {
	b.buf = buf
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
func (b *Builder) Grow(n int) {
	b.copyCheck()
	if n < 0 {
		panic("strings.Builder.Grow: negative count")
	}
	if cap(b.buf)-len(b.buf) < n {
		b.grow(n)
	}
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (b *Builder) Write(p []byte) (int, error) {
	b.copyCheck()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to b's buffer.
// The returned error is always nil.
func (b *Builder) WriteByte(c byte) error {
	b.copyCheck()
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
// It returns the length of r and a nil error.
func (b *Builder) WriteRune(r rune) (int, error) {
	b.copyCheck()
	// Compare as uint32 to correctly handle negative runes.
	if uint32(r) < utf8.RuneSelf {
		b.buf = append(b.buf, byte(r))
		return 1, nil
	}
	l := len(b.buf)
	if cap(b.buf)-l < utf8.UTFMax {
		b.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(b.buf[l:l+utf8.UTFMax], r)
	b.buf = b.buf[:l+n]
	return n, nil
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
	b.copyCheck()
	b.buf = append(b.buf, s...)
	return len(s), nil
}

func (b *Builder) WriteInt(n int) {
	b.copyCheck()
	b.buf = strconv.AppendInt(b.buf, int64(n), 10)
}

func (b *Builder) WriteInt8(n int8) {
	b.copyCheck()
	b.buf = strconv.AppendInt(b.buf, int64(n), 10)
}

func (b *Builder) WriteInt16(n int16) {
	b.copyCheck()
	b.buf = strconv.AppendInt(b.buf, int64(n), 10)
}

func (b *Builder) WriteInt32(n int32) {
	b.copyCheck()
	b.buf = strconv.AppendInt(b.buf, int64(n), 10)
}

func (b *Builder) WriteInt64(n int64) {
	b.copyCheck()
	b.buf = strconv.AppendInt(b.buf, int64(n), 10)
}

func (b *Builder) WriteUint(n uint) {
	b.copyCheck()
	b.buf = strconv.AppendUint(b.buf, uint64(n), 10)
}

func (b *Builder) WriteUint8(n uint8) {
	b.copyCheck()
	b.buf = strconv.AppendUint(b.buf, uint64(n), 10)
}

func (b *Builder) WriteUint16(n uint16) {
	b.copyCheck()
	b.buf = strconv.AppendUint(b.buf, uint64(n), 10)
}

func (b *Builder) WriteUint32(n uint32) {
	b.copyCheck()
	b.buf = strconv.AppendUint(b.buf, uint64(n), 10)
}

func (b *Builder) WriteUint64(n uint64) {
	b.copyCheck()
	b.buf = strconv.AppendUint(b.buf, uint64(n), 10)
}

func (b *Builder) WriteBool(f bool) {
	b.copyCheck()
	b.buf = strconv.AppendBool(b.buf, f)
}

func (b *Builder) WriteFloat32(f float32) {
	b.copyCheck()
	b.buf = strconv.AppendFloat(b.buf, float64(f), 'f', -1, 32)
}

func (b *Builder) WriteFloat64(f float64) {
	b.copyCheck()
	b.buf = strconv.AppendFloat(b.buf, f, 'f', -1, 64)
}

// WriteFloat64Ex appends the string form of the floating-point number f,
// as generated by FormatFloat, to dst and returns the extended buffer.
func (b *Builder) WriteFloat64Ex(f float64, fmt byte, prec, bitSize int) {
	b.copyCheck()
	b.buf = strconv.AppendFloat(b.buf, f, fmt, prec, bitSize)
}

func (b *Builder) WriteBinaryUint8(n uint8) {
	b.copyCheck()
	b.buf = append(b.buf, byte(n))
}

func (b *Builder) WriteBinaryUint16(order binary.ByteOrder, n uint16) {
	b.copyCheck()
	l := len(b.buf)
	b.buf = append(b.buf, 0, 0)
	order.PutUint16(b.buf[l:], n)
	return
}

func (b *Builder) WriteBinaryUint32(order binary.ByteOrder, n uint32) {
	b.copyCheck()
	l := len(b.buf)
	b.buf = append(b.buf, 0, 0, 0, 0)
	order.PutUint32(b.buf[l:], n)
	return
}

func (b *Builder) WriteBinaryUint64(order binary.ByteOrder, n uint64) {
	b.copyCheck()
	l := len(b.buf)
	b.buf = append(b.buf, 0, 0, 0, 0, 0, 0, 0, 0)
	order.PutUint64(b.buf[l:], n)
	return
}

func (b *Builder) Data() []byte {
	return b.buf
}
