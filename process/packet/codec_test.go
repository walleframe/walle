package packet

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/aggronmagi/walle/process/metadata"
	"github.com/stretchr/testify/assert"
)

func TestBytesURICodec(t *testing.T) {
	p := NewPacket()
	p.cmd = CmdNotify
	p.flag = FlagError
	p.reservd = 0x12
	p.msgURI = "uxxx"
	p.payload = []byte("xxxxxxxxxxxxxxxxxx")
	p.metadata["x"] = []string{"b"}

	datas := []func(p *Packet){
		func(p *Packet) {
		},
		func(p *Packet) {
			p.metadata = make(metadata.MD)
		},
	}
	for k, f := range datas {
		t.Run(fmt.Sprint(k), func(t *testing.T) {
			f(p)
			data, err := BytesURICodec.Marshal(p)
			assert.Nil(t, err, "marshal result")
			np := NewPacket()
			err = BytesURICodec.Unmarshal(data, np)
			assert.Nil(t, err, "unmarshal result")
			np.CleanForTest()
			p.CleanForTest()
			assert.EqualValues(t, p, np, "compare source packet")
		})
	}

}

func TestBytesMIDCodec(t *testing.T) {
	p := NewPacket()
	p.cmd = CmdNotify
	p.flag = FlagError
	p.reservd = 0x12
	p.msgID = 5654
	p.payload = []byte("xxxxxxxxxxxxxxxxxx")
	p.metadata["x"] = []string{"b"}

	datas := []func(p *Packet){
		func(p *Packet) {
		},
		func(p *Packet) {
			p.metadata = make(metadata.MD)
		},
	}
	for k, f := range datas {
		t.Run(fmt.Sprint(k), func(t *testing.T) {
			f(p)
			data, err := BytesMIDCodec.Marshal(p)
			assert.Nil(t, err, "marshal result")
			np := NewPacket()
			err = BytesMIDCodec.Unmarshal(data, np)
			assert.Nil(t, err, "unmarshal result")
			p.CleanForTest()
			np.CleanForTest()
			t.Log(data[16:20], np.msgID, binary.BigEndian.Uint32(data[16:]))
			assert.EqualValues(t, p, np, "compare source packet")
		})
	}

}
