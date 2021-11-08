package process

import (
	"encoding/json"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/golang/protobuf/proto"
)

// Codec use for marshal and unmarshal message
type MessageCodec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// MessageJSONCodec json format codec
var MessageCodecJSON MessageCodec = new(jsonCodec)

// jsonCodec for json codec
type jsonCodec struct{}

func (c *jsonCodec) Marshal(v interface{}) (data []byte, err error) {
	data, err = json.Marshal(v)
	err = packet.ErrMarshalFailed.WrapError(err)
	return
}

func (c *jsonCodec) Unmarshal(data []byte, v interface{}) (err error) {
	err = json.Unmarshal(data, v)
	err = packet.ErrUnmarshalFailed.WrapError(err)
	return
}

// ProtobufCodec google protobuf codec
var MessageCodecProtobuf MessageCodec = new(pbCodec)

type pbCodec struct{}

func (c *pbCodec) Marshal(v interface{}) (data []byte, err error) {
	if pb, ok := v.(proto.Message); ok {
		data, err = proto.Marshal(pb)
		err = packet.ErrMarshalFailed.WrapError(err)
		return
	}
	err = packet.ErrMarshalFailed
	return
}

func (c *pbCodec) Unmarshal(data []byte, v interface{}) (err error) {
	if pb, ok := v.(proto.Message); ok {
		err = proto.Unmarshal(data, pb)
		err = packet.ErrUnmarshalFailed.WrapError(err)
		return
	}
	err = packet.ErrUnmarshalFailed
	return
}
