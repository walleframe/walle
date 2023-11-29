package metadata

import (
	"encoding/binary"
	"errors"
	"net/url"
	"sort"

	"github.com/walleframe/walle/util"
)

// Codec use for marshal and unmarshal message
type Codec interface {
	Marshal(v MD) ([]byte, error)
	Unmarshal(data []byte, v MD) error
}

type urlMetaCodec struct{}

func (urlMetaCodec) Marshal(v MD) (data []byte, err error) {
	if v == nil || len(v) == 0 {
		return
	}
	data = util.StringToBytes(url.Values(v).Encode())
	return
}

func (urlMetaCodec) Unmarshal(data []byte, md MD) (err error) {
	if len(data) <= 0 {
		return nil
	}
	val, err := url.ParseQuery(util.BytesToString(data))
	if err != nil {
		return err
	}
	for k, v := range val {
		md[k] = v
	}
	return
}

type binaryMDCodec struct{}

var BinaryCodec Codec = binaryMDCodec{}

func (binaryMDCodec) Marshal(v MD) (data []byte, err error) {
	if v == nil || len(v) == 0 {
		return
	}
	var buf util.Builder
	keys := make([]string, 0, len(v))
	size := 0
	for k, v := range v {
		keys = append(keys, k)
		size += 4 + len(k)
		for _, m := range v {
			size += 2 + len(m)
		}
	}
	buf.SetBuf(make([]byte, 0, size)) // mempool.Pool().Alloc(size)[:0])
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		buf.WriteBinaryUint16(binary.BigEndian, uint16(len(k)))
		buf.WriteBinaryUint16(binary.BigEndian, uint16(len(vs)))
		buf.WriteString(k)
		for _, v := range vs {
			buf.WriteBinaryUint16(binary.BigEndian, uint16(len(v)))
			buf.WriteString(v)
		}
	}
	data = buf.Data()

	return
}

func (binaryMDCodec) Unmarshal(data []byte, v MD) (err error) {
	if len(data) <= 0 {
		return nil
	}
	for len(data) >= 4 {
		l := int(binary.BigEndian.Uint16(data))
		size := int(binary.BigEndian.Uint16(data[2:]))
		if len(data) < 4+l {
			err = ErrInvalidSize
			return
		}
		k := string(data[4 : 4+l])
		data = data[4+l:]
		v[k] = make([]string, 0, size)
		for i := 0; i < int(size); i++ {
			if len(data) < 2 {
				err = ErrInvalidSize
				return
			}
			l = int(binary.BigEndian.Uint16(data))
			if len(data) < 2+l {
				err = ErrInvalidSize
				return
			}
			v[k] = append(v[k], string(data[2:2+l]))
			data = data[2+l:]
		}
	}
	if len(data) > 0 {
		return ErrInvalidSize
	}
	return
}

var ErrInvalidSize = errors.New("invalid size")

var defaultCodec = BinaryCodec

func GetCodec() Codec {
	return defaultCodec
}

func SetCodec(c Codec) {
	defaultCodec = c
}
