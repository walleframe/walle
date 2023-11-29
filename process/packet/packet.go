package packet

import (
	"github.com/walleframe/walle/process/message"
	"github.com/walleframe/walle/process/metadata"
)

//go:generate mockgen -source packet.go -destination ../../testpkg/mock_packet/packet.go

// PacketCmd first byte,
type PacketCmd byte

const (
	// CmdNotify notify,push,oneway
	CmdNotify PacketCmd = iota
	CmdRequest
	CmdResponse
)

// PacketFlag second byte,internal message flag.
type PacketFlag byte

const (
	// FlagError message is an error response
	FlagError PacketFlag = 0x01
)

// Encoder use for encode and decode source packet
type Encoder interface {
	Encode(buf []byte) []byte
	Decode(buf []byte) []byte
}

// Pool reuse packet pool
type Pool interface {
	Get() interface{}
	Put(interface{})
}

// Codec Use for marshal/unmarshal packet.Packet from bytes
type Codec interface {
	Marshal(p interface{}) ([]byte, error)
	Unmarshal(data []byte, p interface{}) error
}

// ProtocolWraper wrap all packet operate, use for custom packet struct.
type ProtocolWraper interface {
	// unmarshal packet's payload by msg codec
	PayloadUnmarshal(pkg interface{}, codec message.Codec, obj interface{}) error
	// marshal packet's payload by msg codec,then set payload binary data into message buf.
	PayloadMarshal(pkg interface{}, codec message.Codec, payload interface{}) (err error)
	// new response packet
	NewResponse(inPkg, outPkg interface{}, md metadata.MD) (err error)
	// new request packet
	NewPacket(inPkg interface{}, cmd PacketCmd, uri interface{}, md metadata.MD) (err error)
	// get pkg metadata
	GetMetadata(pkg interface{}) (md metadata.MD, err error)
}
