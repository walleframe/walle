package process

import (
	"context"
	"time"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/zaplog"
	"go.uber.org/zap"
)

//go:generate gogen imake . -t=wrapContext -r wrapContext=Context -o context.gen.go --merge

// Context 基础
type wrapContext struct {
	// bind to
	p *Process
	// src Context parent is link Context
	src context.Context
	// call chain
	index    int
	handlers []MiddlewareFunc
	// input packet
	in       *packet.Packet
	loadFlag bool
	// log context
	log     *zaplog.Logger
}

// WithValue wrap context.WithValue
func (ctx *wrapContext) WithValue(key, value interface{}) Context {
	ctx.src = context.WithValue(ctx.src, key, value)
	return ctx
}

// WithValue wrap context.WithCancel
func (ctx *wrapContext) WithCancel() (_ Context, cancel func()) {
	ctx.src, cancel = context.WithCancel(ctx.src)
	return ctx, cancel
}

// WithValue wrap context.WithDeadline
func (ctx *wrapContext) WithDeadline(d time.Time) (_ Context, cancel func()) {
	ctx.src, cancel = context.WithDeadline(ctx.src, d)
	return ctx, cancel
}

// WithValue wrap context.WithTimeout
func (ctx *wrapContext) WithTimeout(timeout time.Duration) (_ Context, cancel func()) {
	ctx.src, cancel = context.WithTimeout(ctx.src, timeout)
	return ctx, cancel
}

// Deadline wrap context.Context.Deadline
func (ctx *wrapContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.src.Deadline()
}

// Done wrap context.Context.Done
func (ctx *wrapContext) Done() <-chan struct{} {
	return ctx.src.Done()
}

// Err wrap context.Context.Err
func (ctx *wrapContext) Err() error {
	return ctx.src.Err()
}

// Value wrap context.Context.Value
func (ctx *wrapContext) Value(key interface{}) interface{} {
	return ctx.src.Value(key)
}

// GetRequestPacket get request packet
func (ctx *wrapContext) GetRequestPacket() *packet.Packet {
	return ctx.in
}

// Bind use for unmarshal packet body
func (ctx *wrapContext) Bind(body interface{}) (err error) {
	return ctx.p.unmarshalRspBody(ctx.in, body)
}

// Next call next middleware or router func
func (ctx *wrapContext) Next(nctx Context) {
	index := ctx.index
	if index < len(ctx.handlers) {
		ctx.index++
		ctx.handlers[index](nctx)
	}
	// free packet.Packet
	if index >= len(ctx.handlers) {
		if ctx.in != nil {
			ctx.p.Opts.PacketPool.Push(ctx.in)
			// decr load info
			if ctx.loadFlag {
				ctx.p.Inner.Load.Dec()
			}
			ctx.in = nil
		}
	}
}

// Abort stop call next
func (ctx *wrapContext) Abort() {
	ctx.index = len(ctx.handlers) + 1
}

// Logger get logger
func (ctx *wrapContext) Logger() *zaplog.Logger {
	return ctx.log
}

// WithLogFields
func (ctx *wrapContext) WithLogFields(fields ...zap.Field) {
	if ctx == nil || ctx.log == nil {
		return
	}
	ctx.log = ctx.log.With(fields...)
}

// NewEntry new log entry
func (ctx *wrapContext) NewEntry(funcName string) *zaplog.LogEntities {
	if ctx == nil {
		return nil
	}
	return ctx.NewEntry(funcName)
}
