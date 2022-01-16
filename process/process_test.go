package process

import (
	"bytes"
	"testing"

	message "github.com/aggronmagi/walle/process/message"
	"github.com/aggronmagi/walle/process/metadata"
	"github.com/aggronmagi/walle/process/packet"
	"github.com/aggronmagi/walle/testpkg"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	zap "go.uber.org/zap"
)

// OnRead 读取请求或者通知消息
func TestProcess_OnRead(t *testing.T) {
	mc := gomock.NewController(t)
	f := testpkg.NewMockFuncCall(mc)

	type testJsonST struct {
		V int `json:"v"`
	}
	testMsg := &testJsonST{
		V: 100,
	}
	jsonMsg, err := message.JSONCodec.Marshal(testMsg)
	assert.Nil(t, err, "marshal json codec")

	rq := packet.NewTestPacket(packet.CmdRequest, jsonMsg, metadata.Pairs("k", "v", "n", "10"))
	// &np.Packet{
	// 	Cmd:      int32(np.Command_Request),
	// 	Sequence: 1,
	// 	Metadata: map[string]string{"k": "v", "n": "10"},
	// 	Uri:      "kk",
	// 	Body:     []byte(jsonMsg),
	// }
	rq.SetSeesonID(1)
	rq.SetURI("kk")

	// 函数调用顺序
	f.EXPECT().Call("dispatch-before")
	f.EXPECT().Call("global-before")
	f.EXPECT().Call("mid-before")
	f.EXPECT().Call("exec")
	f.EXPECT().Call("mid-after")
	f.EXPECT().Call("global-after")
	f.EXPECT().Call("dispatch-after")

	// 路由
	r := GetRouter()
	r.Use(func(ctx Context) {
		f.Call("global-before")
		ctx.Next(ctx)
		f.Call("global-after")
	})
	r.Register("kk",
		func(ctx Context) {
			in := ctx.GetRequestPacket().(*packet.Packet)
			assert.NotNil(t, in, "get in packet")
			v, _ := in.GetMD().GetFirstString("k")
			assert.Equal(t, "v", v, "check metadata k")
			n, _ := in.GetMD().GetFirstInt("n")
			assert.Equal(t, int64(10), n, "check metadata n")

			in.CleanForTest()
			rq.CleanForTest()
			assert.EqualValues(t, rq, in, "check pkg")

			rq := &testJsonST{}
			err := ctx.Bind(rq)
			assert.Nil(t, err, "bind result")
			assert.EqualValues(t, testMsg, rq, "check request")
			f.Call("exec")
		},
		func(ctx Context) {
			f.Call("mid-before")
			ctx.Next(ctx)
			f.Call("mid-after")
		},
	)

	buf := &bytes.Buffer{}
	p := NewProcess(
		NewInnerOptions(
			WithInnerOptionsOutput(buf),
			WithInnerOptionsRouter(r),
		),
		NewProcessOptions(
			WithLogger(
				zaplog.NewLogger(zap.NewNop()),
			),
			WithDispatchDataFilter(func(data []byte, next DataDispatcherFunc) (err error) {
				f.Call("dispatch-before")
				err = next(data)
				f.Call("dispatch-after")
				return
			}),
			WithMsgCodec(message.JSONCodec),
		),
	)
	_ = p

	data, err := packet.GetCodec().Marshal(rq)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(data))
	p.OnRead(data)
}

func BenchmarkProcess(b *testing.B) {

	type testJsonST struct {
		V int `json:"v"`
	}
	testMsg := &testJsonST{
		V: 100,
	}
	jsonMsg, err := message.JSONCodec.Marshal(testMsg)
	if err != nil {
		b.Fatal(err)
	}
	rq := packet.NewTestPacket(packet.CmdRequest, jsonMsg, metadata.Pairs("k", "v", "n", "10"))
	rq.SetCmd(packet.CmdRequest)
	rq.SetSeesonID(1)
	rq.SetURI("kk")

	// 路由
	r := &MixRouter{}
	r.Use(func(ctx Context) {
		ctx.Next(ctx)
	})
	r.Register("kk",
		func(ctx Context) {
			// rq := &testJsonST{}
			// err := ctx.Bind(rq)
			// if err != nil {
			// 	b.Fatal(err)
			// }
		},
		func(ctx Context) {
			ctx.Next(ctx)
		},
	)

	buf := &bytes.Buffer{}
	p := NewProcess(
		NewInnerOptions(
			WithInnerOptionsOutput(buf),
			WithInnerOptionsRouter(r),
		),
		NewProcessOptions(
			WithLogger(
				zaplog.NewLogger(zap.NewNop()),
			),
			WithDispatchDataFilter(func(data []byte, next DataDispatcherFunc) (err error) {
				err = next(data)
				return
			}),
			WithMsgCodec(message.JSONCodec),
		),
	)
	_ = p

	data, _ := packet.GetCodec().Marshal(rq)

	b.ResetTimer()

	for k := 0; k < b.N; k++ {
		p.OnRead(data)
	}

}
