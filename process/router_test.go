package process

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/process/packet"
	"github.com/walleframe/walle/testpkg"
)

func TestMixRouter_Methods(t *testing.T) {
	datas := []struct {
		name   string
		custom func(r *MixRouter, f *testpkg.MockFuncCall) (pkg *packet.Packet)
	}{
		{
			"method",
			func(r *MixRouter, f *testpkg.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.Register("kk",
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = packet.NewPacket()
				pkg.SetCmd(packet.CmdRequest)
				pkg.SetURI("kk")
				return
			},
		},
		{
			"rqid",
			func(r *MixRouter, f *testpkg.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.Register(1,
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = packet.NewPacket()
				pkg.SetCmd(packet.CmdRequest)
				pkg.SetMsgID(1)
				return
			},
		},
		{
			"norouter-1",
			func(r *MixRouter, f *testpkg.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.NoRouter(
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = packet.NewPacket()
				pkg.SetCmd(packet.CmdRequest)
				pkg.SetMsgID(1)
				return
			},
		},
		{
			"norouter-2",
			func(r *MixRouter, f *testpkg.MockFuncCall) (pkg *packet.Packet) {

				r.NoRouter(
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("global-before")
						ctx.Next(ctx)
						f.Call("global-after")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = packet.NewPacket()
				pkg.SetCmd(packet.CmdRequest)
				pkg.SetMsgID(1)

				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				return
			},
		},
	}
	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {

			mc := gomock.NewController(t)
			f := testpkg.NewMockFuncCall(mc)

			ctx := &WrapContext{
				SrcContext: context.Background(),
			}

			r := &MixRouter{}
			// func call check
			f.EXPECT().Call("global-before")
			f.EXPECT().Call("mid-before")
			f.EXPECT().Call("exec")
			f.EXPECT().Call("mid-after")
			f.EXPECT().Call("global-after")
			pkg := data.custom(r, f)
			var err error
			ctx.Handlers, err = r.GetHandlers(pkg)
			assert.Nil(t, err, "get router handlers")
			assert.Equal(t, int(3), len(ctx.Handlers), "handler count")
			ctx.Next(ctx)
		})
	}
}
