package metadata

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	var arrays = []struct {
		name  string
		codec Codec
	}{
		{"url", urlMetaCodec{}},
		{"binary", binaryMDCodec{}},
	}

	var data = []MD{
		Pairs("k", "v"),
		{"k":{"v1","v2"}},
	}

	for _, v := range arrays {
		t.Run(v.name, func(t *testing.T) {
			for _, dv := range data {
				t.Run(fmt.Sprintf("%#v", dv), func(t *testing.T) {
					data, err := v.codec.Marshal(dv)
					assert.Nil(t, err, "marshal md")
					assert.NotNil(t, data, "marshal data")
					nv := MD{}
					err = v.codec.Unmarshal(data, nv)
					assert.Nil(t, err, "unmarshal failed")
					//So(dv, ShouldEqual, nv)
					ok := reflect.DeepEqual(dv, nv)
					assert.True(t, ok, "final value")
				})
			}
		})
	}
	assert.Nil(t, nil)
}
