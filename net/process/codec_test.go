package process

import (
	"reflect"
	"testing"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/stretchr/testify/assert"
)

func TestPacketCodec(t *testing.T) {
	data := []PacketCodec{
		PacketCodecProtobuf,
	}
	for _, codec := range data {
		t.Run(reflect.TypeOf(codec).Name(), func(t *testing.T) {
			rq := &packet.Packet{
				Cmd:      int32(packet.Command_Request),
				Sequence: 1,
				Metadata: map[string]string{"k": "v", "n": "10"},
				Uri:      "kk",
			}
			data, _ := codec.Marshal(rq)
			pkg := &packet.Packet{}
			err := codec.Unmarshal(data, pkg)
			assert.Nil(t, err, "unmarhal")
			assert.EqualValues(t, rq.String(), pkg.String(), "check pkg")
		})
	}

}

func TestMessageCodec(t *testing.T) {
	data := []MessageCodec{
		MessageCodecJSON,
		MessageCodecProtobuf,
	}
	for _, codec := range data {
		t.Run(reflect.TypeOf(codec).Name(), func(t *testing.T) {
			rq := &packet.Packet{
				Cmd:      int32(packet.Command_Request),
				Sequence: 1,
				Metadata: map[string]string{"k": "v", "n": "10"},
				Uri:      "kk",
			}
			data, _ := codec.Marshal(rq)
			pkg := &packet.Packet{}
			err := codec.Unmarshal(data, pkg)
			assert.Nil(t, err, "unmarhal")
			assert.EqualValues(t, rq.String(), pkg.String(), "check pkg")
		})
	}

}
