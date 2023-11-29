package message

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/testpkg/msg"
)

func TestMessageCodec(t *testing.T) {
	data := []Codec{
		JSONCodec,
		ProtobufCodec,
	}
	for _, codec := range data {
		t.Run(reflect.TypeOf(codec).Name(), func(t *testing.T) {
			rq := &msg.TestMsg{
				V1: 100,
				V2: "5s4df",
			}
			data, err := codec.Marshal(rq)
			assert.Nil(t, err, "marhal")
			t.Log(string(data))
			pkg := &msg.TestMsg{}
			err = codec.Unmarshal(data, pkg)
			assert.Nil(t, err, "unmarhal")
			assert.EqualValues(t, rq.String(), pkg.String(), "check pkg")
		})
	}

}
