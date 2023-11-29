package rpc

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/process"
	"github.com/walleframe/walle/process/message"
	metadata "github.com/walleframe/walle/process/metadata"
	"github.com/walleframe/walle/process/packet"
	"github.com/walleframe/walle/testpkg"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

func TestProcess_Call(t *testing.T) {
	packet.SetPacketWraper(packet.NewPacketWraper())
	mc := gomock.NewController(t)
	f := testpkg.NewMockFuncCall(mc)

	type testJsonST struct {
		V int `json:"v"`
	}
	testRQ := &testJsonST{
		V: 100,
	}
	jsonMsg, err := message.JSONCodec.Marshal(testRQ)
	assert.Nil(t, err, "marshal json codec")

	rq := packet.NewTestPacket(packet.CmdRequest, jsonMsg, metadata.Pairs("k", "v", "n", "10"))
	rq.SetCmd(packet.CmdRequest)
	rq.SetSeesonID(1)
	rq.SetURI("kk")

	testRS := &testJsonST{1000}
	jsonRS, err := message.JSONCodec.Marshal(testRS)
	assert.Nil(t, err, "new response")

	rs := packet.NewTestPacket(packet.CmdResponse, jsonRS, nil)
	rs.SetURI("kk")
	rs.SetSeesonID(1)

	// 函数调用顺序
	f.EXPECT().Call("dispatch-before")
	f.EXPECT().Call("dispatch-after")

	buf := &bytes.Buffer{}
	p := NewRPCProcess(
		process.NewInnerOptions(
			process.WithInnerOptionOutput(buf),
		),
		process.NewProcessOptions(
			process.WithLogger(
				zaplog.NewLogger(zap.NewNop()),
			),
			process.WithDispatchDataFilter(func(data []byte, next process.DataDispatcherFunc) (err error) {
				f.Call("dispatch-before")
				err = next(data)
				f.Call("dispatch-after")
				return
			}),
			process.WithMsgCodec(message.JSONCodec),
		),
	)
	_ = p

	go func() {
		time.Sleep(time.Millisecond * 30)
		rsData, err := packet.GetCodec().Marshal(rs)
		if err != nil {
			panic(err)
		}
		err = p.OnRead(rsData)
		if err != nil {
			panic(err)
		}
	}()
	if true {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		rs := &testJsonST{}
		err = p.Call(
			&process.WrapContext{
				SrcContext: ctx,
			},
			"kk", testRQ, rs, NewCallOptions(
				WithCallOptionsMetadata(
					metadata.Pairs("k", "v", "n", "10"),
				),
			),
		)
		assert.Nil(t, err, "notify error")
		real := packet.NewPacket()
		err = packet.GetCodec().Unmarshal(buf.Bytes(), real)
		assert.Nil(t, err, "unmarshal data error")
		rq.CleanForTest()
		real.CleanForTest()
		assert.EqualValues(t, rq, real, "final data")
		assert.EqualValues(t, testRS, rs, "respond value")
	}
}

func TestProcess_AsyncCall(t *testing.T) {
	packet.SetPacketWraper(packet.NewPacketWraper())
	mc := gomock.NewController(t)
	f := testpkg.NewMockFuncCall(mc)

	type testJsonST struct {
		V int `json:"v"`
	}
	testRQ := &testJsonST{
		V: 100,
	}
	jsonMsg, err := message.JSONCodec.Marshal(testRQ)
	assert.Nil(t, err, "marshal json codec")

	rq := packet.NewTestPacket(packet.CmdRequest, jsonMsg, metadata.Pairs("k", "v", "n", "10"))
	rq.SetCmd(packet.CmdRequest)
	rq.SetSeesonID(1)
	rq.SetURI("kk")

	testRS := &testJsonST{1000}
	jsonRS, err := message.JSONCodec.Marshal(testRS)
	assert.Nil(t, err, "new response")

	rs := packet.NewTestPacket(packet.CmdResponse, jsonRS, nil)
	rs.SetURI("kk")
	rs.SetSeesonID(1)

	// 函数调用顺序
	f.EXPECT().Call("dispatch-before")
	f.EXPECT().Call("filter-before")
	f.EXPECT().Call("async call")
	f.EXPECT().Call("filter-after")
	f.EXPECT().Call("dispatch-after")

	buf := &bytes.Buffer{}
	p := NewRPCProcess(
		process.NewInnerOptions(
			process.WithInnerOptionOutput(buf),
		),
		process.NewProcessOptions(
			process.WithLogger(
				zaplog.NewLogger(zap.NewNop()),
			),
			process.WithDispatchDataFilter(func(data []byte, next process.DataDispatcherFunc) (err error) {
				f.Call("dispatch-before")
				err = next(data)
				f.Call("dispatch-after")
				return
			}),
			process.WithMsgCodec(message.JSONCodec),
		),
	)
	_ = p

	err = p.AsyncCall(
		&process.WrapContext{
			SrcContext: context.Background(),
		},
		"kk", testRQ,
		func(ctx process.Context) {
			f.Call("async call")
			rs := &testJsonST{}
			err := ctx.Bind(rs)
			assert.Nil(t, err, "async response bind")
			assert.EqualValues(t, testRS, rs, "respond value")
		},
		NewAsyncCallOptions(
			WithAsyncCallOptionsMetadata(
				metadata.Pairs("k", "v", "n", "10"),
			),
			WithAsyncCallOptionsResponseFilter(func(ctx process.Context, req, rsp interface{}) {
				f.Call("filter-before")
				ctx.Next(ctx)
				f.Call("filter-after")
			}),
			WithAsyncCallOptionsWaitFilter(func(await func()) {
				go await()
			}),
		),
	)
	assert.Nil(t, err, "notify error")
	real := packet.NewPacket()
	err = packet.GetCodec().Unmarshal(buf.Bytes(), real)
	assert.Nil(t, err, "unmarshal data error")
	rq.CleanForTest()
	real.CleanForTest()
	assert.EqualValues(t, rq, real, "final data")

	rsData, _ := packet.GetCodec().Marshal(rs)
	p.OnRead(rsData)

	//wg.Wait()
}

func TestProcess_Notify(t *testing.T) {
	packet.SetPacketWraper(packet.NewPacketWraper())
	type testJsonST struct {
		V int `json:"v"`
	}
	testMsg := &testJsonST{
		V: 100,
	}
	jsonMsg, err := message.JSONCodec.Marshal(testMsg)
	assert.Nil(t, err, "marshal json codec")

	rq := packet.NewTestPacket(packet.CmdNotify, jsonMsg, metadata.Pairs("k", "v", "n", "10"))
	rq.SetCmd(packet.CmdRequest)
	rq.SetSeesonID(1)
	rq.SetURI("kk")

	buf := &bytes.Buffer{}
	p := NewRPCProcess(
		process.NewInnerOptions(
			process.WithInnerOptionOutput(buf),
		),
		process.NewProcessOptions(
			process.WithLogger(
				zaplog.NewLogger(zap.NewNop()),
			),
			process.WithMsgCodec(message.JSONCodec),
		),
	)
	_ = p

	err = p.Notify(&process.WrapContext{}, "kk", testMsg, NewNoticeOptions(
		WithNoticeOptionsMetadata(
			metadata.Pairs("k", "v", "n", "10"),
		),
	))
	assert.Nil(t, err, "notify error")
	real := packet.NewPacket()
	err = packet.GetCodec().Unmarshal(buf.Bytes(), real)
	assert.Nil(t, err, "unmarshal data error")
	rq.CleanForTest()
	real.CleanForTest()
	assert.EqualValues(t, rq, real, "final data")
}

func TestProcess_WithRouter(t *testing.T) {
	type testJsonST struct {
		V int `json:"v"`
	}
	testMsg := &testJsonST{
		V: 100,
	}
	jsonMsg, err := message.JSONCodec.Marshal(testMsg)
	assert.Nil(t, err, "marshal json codec")

	rq := packet.NewTestPacket(packet.CmdNotify, jsonMsg, metadata.Pairs("k", "v", "n", "10"))
	rq.SetCmd(packet.CmdRequest)
	rq.SetSeesonID(1)
	rq.SetURI("kk")

	router := process.GetRouter()

	router.Register("kk", func(ctx process.Context) {
		rq := &testJsonST{}
		err := ctx.Bind(rq)
		assert.Nil(t, err, "kk.rq.bind")
		assert.EqualValues(t, 100, rq.V, "kk.rq.v")
		err = ctx.Respond(ctx, rq, nil)
		assert.Nil(t, err, "kk.rq.response")
	})

	buf := &bytes.Buffer{}
	p := NewRPCProcess(
		process.NewInnerOptions(
			process.WithInnerOptionOutput(buf),
		),
		process.NewProcessOptions(
			process.WithLogger(
				zaplog.NewLogger(zap.NewNop()),
			),
			process.WithMsgCodec(message.JSONCodec),
		),
	)
	_ = p

	data, err := packet.GetCodec().Marshal(rq)
	assert.Nil(t, err, "marshal request error")
	err = p.OnRead(data)
	assert.Nil(t, err, "onread error")
	real := packet.NewPacket()
	err = packet.GetCodec().Unmarshal(buf.Bytes(), real)
	assert.Nil(t, err, "unmarshal data error")
	rq.SetCmd(packet.CmdResponse)
	rq.SetMD(make(metadata.MD))
	rq.CleanForTest()
	real.CleanForTest()
	assert.EqualValues(t, rq, real, "final data")
}
