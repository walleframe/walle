package process

import (
	"context"
	"sync"
	"time"

	"github.com/aggronmagi/walle/process/errcode"
	"github.com/aggronmagi/walle/process/metadata"
	"github.com/aggronmagi/walle/zaplog"
	"go.uber.org/zap"
)

//go:generate gogen imake . -t=WrapContext -r WrapContext=Context -o context.gen.go --merge
//go:generate mockgen -source context.gen.go -destination ../testpkg/mock_process/context.go

// Context 基础
type WrapContext struct {
	// config
	Inner *InnerOptions
	Opts  *ProcessOptions
	// SrcContext Context parent is link Context
	SrcContext context.Context
	// call chain
	Index    int
	Handlers []MiddlewareFunc
	// input packet
	InPkg    interface{}
	LoadFlag bool
	// Log context
	Log *zaplog.Logger
	// use to free context
	FreeContext Context
}

// WithValue wrap context.WithValue
func (ctx *WrapContext) WithValue(key, value interface{}) Context {
	ctx.SrcContext = context.WithValue(ctx.SrcContext, key, value)
	return ctx
}

// WithValue wrap context.WithCancel
func (ctx *WrapContext) WithCancel() (_ Context, cancel func()) {
	ctx.SrcContext, cancel = context.WithCancel(ctx.SrcContext)
	return ctx, cancel
}

// WithValue wrap context.WithDeadline
func (ctx *WrapContext) WithDeadline(d time.Time) (_ Context, cancel func()) {
	ctx.SrcContext, cancel = context.WithDeadline(ctx.SrcContext, d)
	return ctx, cancel
}

// WithValue wrap context.WithTimeout
func (ctx *WrapContext) WithTimeout(timeout time.Duration) (_ Context, cancel func()) {
	ctx.SrcContext, cancel = context.WithTimeout(ctx.SrcContext, timeout)
	return ctx, cancel
}

// Deadline wrap context.Context.Deadline
func (ctx *WrapContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.SrcContext.Deadline()
}

// Done wrap context.Context.Done
func (ctx *WrapContext) Done() <-chan struct{} {
	return ctx.SrcContext.Done()
}

// Err wrap context.Context.Err
func (ctx *WrapContext) Err() error {
	return ctx.SrcContext.Err()
}

// Value wrap context.Context.Value
func (ctx *WrapContext) Value(key interface{}) interface{} {
	return ctx.SrcContext.Value(key)
}

// GetRequestPacket get request packet
func (ctx *WrapContext) GetRequestPacket() interface{} {
	return ctx.InPkg
}

// GetReqeustMD get request metadata
func (ctx *WrapContext) GetReqeustMD() (metadata.MD, error) {
	return ctx.Opts.PacketWraper.GetMetadata(ctx.InPkg)
}

// Bind use for unmarshal packet body
func (ctx *WrapContext) Bind(body interface{}) (err error) {
	return ctx.Opts.PacketWraper.PayloadUnmarshal(ctx.InPkg, ctx.Opts.MsgCodec, body)
}

// Respond write response.
func (ctx *WrapContext) Respond(_ context.Context, body interface{}, md metadata.MD) (err error) {
	if ctx.Inner.Output == nil {
		err = errcode.ErrUnexpectedCode
		return
	}
	wp := ctx.Opts.PacketWraper
	outPkg := ctx.Opts.PacketPool.Get()
	err = wp.NewResponse(ctx.InPkg, outPkg, md)
	if err != nil {
		return
	}
	err = wp.PayloadMarshal(outPkg, ctx.Opts.MsgCodec, body)
	if err != nil {
		return
	}
	data, err := ctx.Opts.PacketCodec.Marshal(outPkg)
	if err != nil {
		return
	}
	_, err = ctx.Inner.Output.Write(data)
	ctx.Opts.PacketPool.Put(outPkg)
	return
}

// Next call next middleware or router func
func (ctx *WrapContext) Next(nctx Context) {
	index := ctx.Index
	if index < len(ctx.Handlers) {
		ctx.Index++
		ctx.Handlers[index](nctx)
	}
	// free packet.Packet
	if index+1 >= len(ctx.Handlers) {
		if ctx.InPkg != nil {
			ctx.Opts.PacketPool.Put(ctx.InPkg)
			// decr load info
			if ctx.LoadFlag {
				ctx.Inner.Load.Dec()
			}
			ctx.InPkg = nil
		}
		if ctx.FreeContext != nil {
			ctx.Inner.ContextPool.FreeContext(ctx.FreeContext)
		}
	}
}

// Abort stop call next
func (ctx *WrapContext) Abort() {
	ctx.Index = len(ctx.Handlers) + 1
}

// Logger get logger
func (ctx *WrapContext) Logger() *zaplog.Logger {
	return ctx.Log
}

// WithLogFields
func (ctx *WrapContext) WithLogFields(fields ...zap.Field) {
	if ctx == nil || ctx.Log == nil {
		return
	}
	ctx.Log = ctx.Log.With(fields...)
}

// NewEntry new log entry
func (ctx *WrapContext) NewEntry(funcName string) *zaplog.LogEntities {
	if ctx == nil {
		return nil
	}
	return ctx.NewEntry(funcName)
}

type ContextPool interface {
	NewContext(inner *InnerOptions, opts *ProcessOptions, inPkg interface{}, handlers []MiddlewareFunc, loadFlag bool) Context
	FreeContext(Context)
}
type wrapContextPool struct {
	sync.Pool
}

func (p *wrapContextPool) NewContext(inner *InnerOptions, opts *ProcessOptions, inPkg interface{}, handlers []MiddlewareFunc, loadFlag bool) Context {
	ctx := p.Get().(*WrapContext)
	ctx.Inner = inner
	ctx.Opts = opts
	ctx.SrcContext = inner.ParentCtx
	ctx.Index = 0
	ctx.Handlers = handlers
	ctx.InPkg = inPkg
	ctx.LoadFlag = loadFlag
	ctx.Log = opts.Logger
	ctx.FreeContext = ctx
	return ctx
}

func (p *wrapContextPool) FreeContext(ctx Context) {
	p.Put(ctx)
}

var WrapContextPool ContextPool = &wrapContextPool{
	Pool: sync.Pool{
		New: func() interface{} {
			return &WrapContext{}
		},
	},
}
