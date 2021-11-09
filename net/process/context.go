package process

import (
	"context"
	"net"
	"time"

	"github.com/aggronmagi/walle/net/packet"
	"github.com/aggronmagi/walle/zaplog"
	"go.uber.org/zap"
)

type Link interface {
	Call(uri interface{}, req, rsp interface{}, opts ...interface{})
	AsyncCall()
	Notice(msg interface{}) (err error)
	GetNetConn() net.Conn
}
//go:generate gogen imake . -t=wrapContext -r wrapContext=Context -t=wrapLogEntry -r wrapLogEntry=LogEntry -o context.gen.go --merge

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
	checked []zap.Field
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
	ctx.checked = append(ctx.checked, fields...)
}

// NewEntry new log entry
func (ctx *wrapContext) NewEntry(funcName string) LogEntry {
	if ctx == nil {
		return nil
	}
	entry := &wrapLogEntry{
		log:     ctx.log,
		ctx:     ctx,
		checked: []zap.Field{zap.String("fname", funcName)},
	}
	return entry
}

// log context
type wrapLogEntry struct {
	log     *zaplog.Logger
	ctx     *wrapContext
	whenErr []zap.Field
	checked []zap.Field
	start   *time.Time
}

// appendLogFields use for logger.
func (entry *wrapLogEntry) appendLogFields(lv zaplog.Level, fields []zap.Field) []zap.Field {
	if entry == nil || entry.log == nil {
		return fields
	}
	cnt := len(entry.checked) + len(fields) + 1 + len(entry.ctx.checked)
	if lv <= zaplog.ERR || lv >= zaplog.DEV {
		cnt += len(entry.whenErr)
	}
	ret := make([]zap.Field, 1, cnt)
	switch lv {
	case zaplog.EMERG:
		ret[0] = zaplog.ZapLevelFieldEmerg
	case zaplog.ALERT:
		ret[0] = zaplog.ZapLevelFieldAlert
	case zaplog.CRIT:
		ret[0] = zaplog.ZapLevelFieldCrit
	case zaplog.ERR:
		ret[0] = zaplog.ZapLevelFieldErr
	case zaplog.WARNING:
		ret[0] = zaplog.ZapLevelFieldWarning
	case zaplog.INFO:
		ret[0] = zaplog.ZapLevelFieldInfo
	case zaplog.NOTICE:
		ret[0] = zaplog.ZapLevelFieldNotice
	case zaplog.DEBUG:
		ret[0] = zaplog.ZapLevelFieldDebug
	case zaplog.DEV:
		ret[0] = zaplog.ZapLevelFieldDev
	default:
		if lv > zaplog.DEV {
			ret[0] = zaplog.ZapLevelFieldDev
		} else {
			ret[0] = zaplog.ZapLevelFieldEmerg
		}
	}
	ret = append(ret, entry.ctx.checked...)
	ret = append(ret, entry.checked...)
	if lv <= zaplog.ERR || lv >= zaplog.DEV {
		ret = append(ret, entry.whenErr...)
	}
	ret = append(ret, fields...)
	return ret
}

// Must must write fields to log
func (entry *wrapLogEntry) Must(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	entry.checked = append(entry.checked, fields...)
}

// WhenErr write fields when write error level log
func (entry *wrapLogEntry) WhenErr(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	entry.whenErr = append(entry.whenErr, fields...)
}

// IfWarn4 if enable warn level,write this fields to log
func (entry *wrapLogEntry) IfWarn4(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	if !entry.log.Enabled(zaplog.WARNING) {
		return
	}
	entry.checked = append(entry.checked, fields...)
}

// IfInfo5 if enable info level,write this fields to log
func (entry *wrapLogEntry) IfInfo5(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	if !entry.log.Enabled(zaplog.INFO) {
		return
	}
	entry.checked = append(entry.checked, fields...)
}

// IfNotice6 if enable notice level,write this fields to log
func (entry *wrapLogEntry) IfNotice6(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	if !entry.log.Enabled(zaplog.NOTICE) {
		return
	}
	entry.checked = append(entry.checked, fields...)
}

// IfDebug7 if enable debug level,write this fields to log
func (entry *wrapLogEntry) IfDebug7(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	if !entry.log.Enabled(zaplog.DEBUG) {
		return
	}
	entry.checked = append(entry.checked, fields...)
}

// IfDevelop8 if enable dev,write this fields to log
func (entry *wrapLogEntry) IfDevelop8(fields ...zap.Field) {
	if entry == nil || entry.log == nil {
		return
	}
	if !entry.log.Enabled(zaplog.DEV) {
		return
	}
	entry.checked = append(entry.checked, fields...)
}

// Emerg  (emergency): a situation that will cause the host system to be unavailable
func (entry *wrapLogEntry) Emerg0(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.EMERG) {
		return
	}
	entry.log.Emerg0(msg, entry.appendLogFields(zaplog.EMERG, fields)...)
	return
}

// Alert problems that must be resolved immediately
func (entry *wrapLogEntry) Alert1(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.ALERT) {
		return
	}
	entry.log.Alert1(msg, entry.appendLogFields(zaplog.ALERT, fields)...)
	return
}

// CRIT (serious): a more serious situation
func (entry *wrapLogEntry) Crit2(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.CRIT) {
		return
	}
	entry.log.Crit2(msg, entry.appendLogFields(zaplog.CRIT, fields)...)
	return
}

func (entry *wrapLogEntry) Error3(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.ERR) {
		return
	}
	entry.log.Error3(msg, entry.appendLogFields(zaplog.ERR, fields)...)
	return
}

// WARNING: events that may affect the function of the system
func (entry *wrapLogEntry) Warn4(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.WARNING) {
		return
	}
	entry.log.Warn4(msg, entry.appendLogFields(zaplog.WARNING, fields)...)
	return
}

func (entry *wrapLogEntry) Info5(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.INFO) {
		return
	}
	entry.log.Info5(msg, entry.appendLogFields(zaplog.INFO, fields)...)
	return
}

// NOTICE: will not affect the system but it is worth noting
func (entry *wrapLogEntry) Notice6(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.NOTICE) {
		return
	}
	entry.log.Notice6(msg, entry.appendLogFields(zaplog.NOTICE, fields)...)
	return
}

// Debug7 Flow log, debugging information. Used for program debugging.
func (entry *wrapLogEntry) Debug7(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.DEBUG) {
		return
	}
	entry.log.Debug7(msg, entry.appendLogFields(zaplog.DEBUG, fields)...)
	return
}

// Develop8 Log details. All operation logs, component operation logs
func (entry *wrapLogEntry) Develop8(msg string, fields ...zap.Field) {
	if !entry.log.Enabled(zaplog.DEV) {
		return
	}
	entry.log.Develop8(msg, entry.appendLogFields(zaplog.DEV, fields)...)
	return
}
