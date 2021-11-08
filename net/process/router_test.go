package process

import (
	"context"
	"testing"

	"github.com/aggronmagi/walle/internal/util/test"
	"github.com/aggronmagi/walle/net/packet"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMixRouter_Methods(t *testing.T) {
	datas := []struct {
		name   string
		custom func(r *MixRouter, f *test.MockFuncCall) (pkg *packet.Packet)
	}{
		{
			"method",
			func(r *MixRouter, f *test.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.Method("kk",
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = &packet.Packet{
					Cmd: int32(packet.Command_Request),
					Uri: "kk",
				}
				return
			},
		},
		{
			"rqid",
			func(r *MixRouter, f *test.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.RequestID(1,
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = &packet.Packet{
					Cmd:        int32(packet.Command_Request),
					ReservedRq: 1,
				}
				return
			},
		},
		{
			"norouter-1",
			func(r *MixRouter, f *test.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.NoRouter(true,
					func(ctx Context) {
						f.Call("exec")
					},
					func(ctx Context) {
						f.Call("mid-before")
						ctx.Next(ctx)
						f.Call("mid-after")
					},
				)
				pkg = &packet.Packet{
					Cmd:        int32(packet.Command_Request),
					ReservedRq: 1,
				}
				return
			},
		},
		{
			"norouter-2",
			func(r *MixRouter, f *test.MockFuncCall) (pkg *packet.Packet) {
				r.Use(func(ctx Context) {
					f.Call("global-before")
					ctx.Next(ctx)
					f.Call("global-after")
				})
				r.NoRouter(false,
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
				pkg = &packet.Packet{
					Cmd:        int32(packet.Command_Request),
					ReservedRq: 1,
				}
				return
			},
		},
	}
	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {

			mc := gomock.NewController(t)
			f := test.NewMockFuncCall(mc)

			ctx := &wrapContext{
				src: context.Background(),
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
			ctx.handlers, err = r.GetHandlers(pkg)
			assert.Nil(t, err, "get router handlers")
			assert.Equal(t, int(3), len(ctx.handlers), "handler count")
			ctx.Next(ctx)
		})
	}
}
