package packet

//
type emtpyPacketCoder struct{}

func (emtpyPacketCoder) Encode(buf []byte) []byte {
	return buf
}

func (emtpyPacketCoder) Decode(buf []byte) []byte {
	return buf
}

// EmtpyPacketEncoder not encode and decode source packet
var EmtpyPacketEncoder Encoder = emtpyPacketCoder{}


var defaultPacketEncoder Encoder = EmtpyPacketEncoder

func GetEncoder() Encoder {
	return defaultPacketEncoder
}

func SetPacketEncoder(coder Encoder) {
	defaultPacketEncoder = coder
}


//
func NewTeeCoder(coders ...Encoder) Encoder {
	switch len(coders) {
	case 0:
		return EmtpyPacketEncoder
	case 1:
		return coders[0]
	default:
		return &teePacketCoder{
			coders: coders,
		}
	}
}

var _ Encoder = (*teePacketCoder)(nil)

type teePacketCoder struct {
	coders []Encoder
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
