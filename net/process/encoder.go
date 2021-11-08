package process

// PacketEncoder use for encode and decode source packet
type PacketEncoder interface {
	Encode(buf []byte) []byte
	Decode(buf []byte) []byte
}

func NewTeeCoder(coders ...PacketEncoder) PacketEncoder {
	switch len(coders) {
	case 0:
		return &EmtpyPacketCoder{}
	case 1:
		return coders[0]
	default:
		return &teePacketCoder{
			coders: coders,
		}
	}
}

var _ PacketEncoder = &teePacketCoder{}

type teePacketCoder struct {
	coders []PacketEncoder
}

func (tee *teePacketCoder) Encode(buf []byte) []byte {
	for _, v := range tee.coders {
		buf = v.Encode(buf)
	}
	return buf
}

func (tee *teePacketCoder) Decode(buf []byte) []byte {
	for _, v := range tee.coders {
		buf = v.Decode(buf)
	}
	return buf
}

// EmtpyPacketCoder not encode and decode source packet
type EmtpyPacketCoder struct{}

func (tee *EmtpyPacketCoder) Encode(buf []byte) []byte {
	return buf
}

func (tee *EmtpyPacketCoder) Decode(buf []byte) []byte {
	return buf
}
