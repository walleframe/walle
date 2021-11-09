package process

import (
	"bytes"
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aggronmagi/walle/internal/util/test"
	"github.com/aggronmagi/walle/net/packet"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	zap "go.uber.org/zap"
)

// OnRead 读取请求或者通知消息
func TestProcess_OnRead(t *testing.T) {
	mc := gomock.NewController(t)
	f := test.NewMockFuncCall(mc)

	type testJsonST struct {
		V int `json:"v"`
	}
	testMsg := &testJsonST{
		V: 100,
	}
	jsonMsg, err := MessageCodecJSON.Marshal(testMsg)
	assert.Nil(t, err, "marshal json codec")

	rq := &packet.Packet{
		Cmd:      int32(packet.Command_Request),
		Sequence: 1,
		Metadata: map[string]string{"k": "v", "n": "10"},
		Uri:      "kk",
		Body:     []byte(jsonMsg),
	}

	// 函数调用顺序
	f.EXPECT().Call("dispatch-before")
	f.EXPECT().Call("global-before")
	f.EXPECT().Call("mid-before")
	f.EXPECT().Call("exec")
	f.EXPECT().Call("mid-after")
	f.EXPECT().Call("global-after")
	f.EXPECT().Call("dispatch-after")

	// 路由
	r := &MixRouter{}
	r.Use(func(ctx Context) {
		f.Call("global-before")
		ctx.Next(ctx)
		f.Call("global-after")
	})
	r.Method("kk",
		func(ctx Context) {
			in := ctx.GetRequestPacket()
			assert.NotNil(t, in, "get in packet")
			v, _ := in.GetMetadataString("k")
			assert.Equal(t, "v", v, "check metadata k")
			n, _ := in.GetMetadataInt64("n")
			assert.Equal(t, int64(10), n, "check metadata n")

			assert.EqualValues(t, rq.String(), in.String(), "check pkg")

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
				zaplog.NewLogger(zaplog.DEV, zap.NewNop()),
			),
			WithDispatchDataFilter(func(data []byte, next PacketDispatcherFunc) (err error) {
				f.Call("dispatch-before")
				err = next(data)
				f.Call("dispatch-after")
				return
			}),
			WithMsgCodec(MessageCodecJSON),
		),
	)
	_ = p

	data, _ := PacketCodecProtobuf.Marshal(rq)
	p.OnRead(data)
}

func TestProcess_Call(t *testing.T) {
	mc := gomock.NewController(t)
	f := test.NewMockFuncCall(mc)

	type testJsonST struct {
		V int `json:"v"`
	}
	testRQ := &testJsonST{
		V: 100,
	}
	jsonMsg, err := MessageCodecJSON.Marshal(testRQ)
	assert.Nil(t, err, "marshal json codec")

	rq := &packet.Packet{
		Cmd:      int32(packet.Command_Request),
		Sequence: 1,
		Metadata: map[string]string{"k": "v", "n": "10"},
		Uri:      "kk",
		Body:     []byte(jsonMsg),
	}
	rs := rq.NewResponse()
	testRS := &testJsonST{1000}
	rs.Body, _ = MessageCodecJSON.Marshal(testRS)

	// 函数调用顺序
	f.EXPECT().Call("dispatch-before")
	f.EXPECT().Call("dispatch-after")

	buf := &bytes.Buffer{}
	p := NewProcess(
		NewInnerOptions(
			WithInnerOptionsOutput(buf),
		),
		NewProcessOptions(
			WithLogger(
				zaplog.NewLogger(zaplog.DEV, zap.NewNop()),
			),
			WithDispatchDataFilter(func(data []byte, next PacketDispatcherFunc) (err error) {
				f.Call("dispatch-before")
				err = next(data)
				f.Call("dispatch-after")
				return
			}),
			WithMsgCodec(MessageCodecJSON),
		),
	)
	_ = p
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		rs := &testJsonST{}
		err = p.Call(
			&wrapContext{
				src: context.Background(),
			},
			"kk", testRQ, rs, NewCallOptions(
				WithCallOptionsMetadata(
					MetadataString("k", "v"),
					MetadataInt64("n", 10),
				),
			),
		)
		assert.Nil(t, err, "notify error")
		real := &packet.Packet{}
		err = PacketCodecProtobuf.Unmarshal(buf.Bytes(), real)
		assert.Nil(t, err, "unmarshal data error")
		assert.EqualValues(t, rq.String(), real.String(), "final data")
		assert.EqualValues(t, testRS, rs, "respond value")
	}()
	runtime.Gosched()
	time.Sleep(time.Millisecond)
	rsData, _ := PacketCodecProtobuf.Marshal(rs)
	p.OnRead(rsData)
	wg.Wait()
}

func TestProcess_AsyncCall(t *testing.T) {
	mc := gomock.NewController(t)
	f := test.NewMockFuncCall(mc)

	type testJsonST struct {
		V int `json:"v"`
	}
	testRQ := &testJsonST{
		V: 100,
	}
	jsonMsg, err := MessageCodecJSON.Marshal(testRQ)
	assert.Nil(t, err, "marshal json codec")

	rq := &packet.Packet{
		Cmd:      int32(packet.Command_Request),
		Flag:     uint32(packet.Flag_ClientAsync),
		Sequence: 1,
		Metadata: map[string]string{"k": "v", "n": "10"},
		Uri:      "kk",
		Body:     []byte(jsonMsg),
	}
	rs := rq.NewResponse()
	testRS := &testJsonST{1000}
	rs.Body, _ = MessageCodecJSON.Marshal(testRS)

	// 函数调用顺序
	f.EXPECT().Call("dispatch-before")
	f.EXPECT().Call("filter-before")
	f.EXPECT().Call("async call")
	f.EXPECT().Call("filter-after")
	f.EXPECT().Call("dispatch-after")

	buf := &bytes.Buffer{}
	p := NewProcess(
		NewInnerOptions(
			WithInnerOptionsOutput(buf),
		),
		NewProcessOptions(
			WithLogger(
				zaplog.NewLogger(zaplog.DEV, zap.NewNop()),
			),
			WithDispatchDataFilter(func(data []byte, next PacketDispatcherFunc) (err error) {
				f.Call("dispatch-before")
				err = next(data)
				f.Call("dispatch-after")
				return
			}),
			WithMsgCodec(MessageCodecJSON),
		),
	)
	_ = p

	wg := sync.WaitGroup{}
	wg.Add(1)

	err = p.AsyncCall(
		&wrapContext{
			src: context.Background(),
		},
		"kk", testRQ,
		func(ctx Context) {
			f.Call("async call")
			rs := &testJsonST{}
			err := ctx.Bind(rs)
			assert.Nil(t, err, "async response bind")
			assert.EqualValues(t, testRS, rs, "respond value")
		},
		NewAsyncCallOptions(
			WithAsyncCallOptionsMetadata(
				MetadataString("k", "v"),
				MetadataInt64("n", 10),
			),
			WithAsyncCallOptionsResponseFilter(func(ctx Context, req, rsp *packet.Packet) {
				f.Call("filter-before")
				ctx.Next(ctx)
				f.Call("filter-after")
			}),
			WithAsyncCallOptionsWaitFilter(func(await func()) {
				go func() {
					await()
					wg.Done()
				}()
			}),
		),
	)
	assert.Nil(t, err, "notify error")
	real := &packet.Packet{}
	err = PacketCodecProtobuf.Unmarshal(buf.Bytes(), real)
	assert.Nil(t, err, "unmarshal data error")
	assert.EqualValues(t, rq.String(), real.String(), "final data")

	rsData, _ := PacketCodecProtobuf.Marshal(rs)
	p.OnRead(rsData)

	//wg.Wait()
}

func TestProcess_Notify(t *testing.T) {
	type testJsonST struct {
		V int `json:"v"`
	}
	testMsg := &testJsonST{
		V: 100,
	}
	jsonMsg, err := MessageCodecJSON.Marshal(testMsg)
	assert.Nil(t, err, "marshal json codec")

	rq := &packet.Packet{
		Cmd:      int32(packet.Command_Oneway),
		Flag:     uint32(packet.Flag_ClientAsync),
		Sequence: 1,
		Metadata: map[string]string{"k": "v", "n": "10"},
		Uri:      "kk",
		Body:     []byte(jsonMsg),
	}

	buf := &bytes.Buffer{}
	p := NewProcess(
		NewInnerOptions(
			WithInnerOptionsOutput(buf),
		),
		NewProcessOptions(
			WithLogger(
				zaplog.NewLogger(zaplog.DEV, zap.NewNop()),
			),
			WithMsgCodec(MessageCodecJSON),
		),
	)
	_ = p

	err = p.Notify(&wrapContext{}, "kk", testMsg, NewNoticeOptions(
		WithNoticeOptionsMetadata(
			MetadataString("k", "v"),
			MetadataInt64("n", 10),
		),
	))
	assert.Nil(t, err, "notify error")
	real := &packet.Packet{}
	err = PacketCodecProtobuf.Unmarshal(buf.Bytes(), real)
	assert.Nil(t, err, "unmarshal data error")
	assert.EqualValues(t, rq.String(), real.String(), "final data")
}
