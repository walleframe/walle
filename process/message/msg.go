package message

import (
	"github.com/aggronmagi/walle/process/errcode"
	//"github.com/golang/protobuf/proto"

	proto "github.com/gogo/protobuf/proto"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//go:generate mockgen -source msg.go -destination ../../testpkg/mock_message/codec_msg.go

// Codec use for marshal and unmarshal message
type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// MessageJSONCodec json format codec
var JSONCodec Codec = new(jsonCodec)

// jsonCodec for json codec
type jsonCodec struct{}

func (c *jsonCodec) Marshal(v interface{}) (_ []byte, err error) {
	//return json.Marshal(v)
	// copy from "github.com/json-iterator/go"， reduce one memory copy
	stream := json.BorrowStream(nil)
	defer json.ReturnStream(stream)
	stream.WriteVal(v)
	if stream.Error != nil {
		return nil, errcode.WrapError(errcode.ErrMarshalFailed, stream.Error)
	}
	result := stream.Buffer()
	copied := make([]byte, len(result)) // mempool.Pool().Alloc(len(result))
	copy(copied, result)
	return copied, nil
}

func (c *jsonCodec) Unmarshal(data []byte, v interface{}) (err error) {
	err = json.Unmarshal(data, v)
	err = errcode.WrapError(errcode.ErrUnmarshalFailed, err)
	return
}

// ProtobufCodec google protobuf codec
var ProtobufCodec Codec = new(pbCodec)

type pbCodec struct{}

func (c *pbCodec) Marshal(v interface{}) (data []byte, err error) {
	if pb, ok := v.(proto.Message); ok {
		//data =  mempool.Pool().Alloc(proto.Size(pb))
		//buf := proto.NewBuffer(data[:0])
		data, err = proto.Marshal(pb)
		err = errcode.WrapError(errcode.ErrMarshalFailed, err) // buf.Marshal(pb))
		return
	}
	err = errcode.ErrMarshalFailed
	return
}

func (c *pbCodec) Unmarshal(data []byte, v interface{}) (err error) {
	if pb, ok := v.(proto.Message); ok {
		err = proto.Unmarshal(data, pb)
		err = errcode.WrapError(errcode.ErrUnmarshalFailed, err)
		return
	}
	err = errcode.ErrUnmarshalFailed
	return
}
