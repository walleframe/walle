package process

import (
	"github.com/aggronmagi/walle/net/packet"
	"github.com/golang/protobuf/proto"
)

// PacketCodec Use for marshal/unmarshal packet.Packet from bytes
type PacketCodec interface {
	Marshal(p *packet.Packet) ([]byte, error)
	Unmarshal(data []byte, p *packet.Packet) error
}

// PacketCodecProtobuf protobuf formtat packet data
var PacketCodecProtobuf PacketCodec = &pbPacketCodec{}

type pbPacketCodec struct{}

func (p *pbPacketCodec) Marshal(pb *packet.Packet) (data []byte, err error) {
	data, err = proto.Marshal(pb)
	err = packet.ErrMarshalFailed.WrapError(err)
	return
}
func (p *pbPacketCodec) Unmarshal(data []byte, pb *packet.Packet) (err error) {
	err = proto.Unmarshal(data, pb)
	err = packet.ErrUnmarshalFailed.WrapError(err)
	return
}
