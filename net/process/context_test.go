package process

import (
	"fmt"
	"testing"

	"github.com/aggronmagi/walle/internal/util/test"
	zaplog "github.com/aggronmagi/walle/zaplog"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestContext_Bind(t *testing.T) {

}

func TestContext_CallChain(t *testing.T) {
	datas := []struct {
		name  string
		num   int
		abort int
	}{
		{"normal", 0, -1},
		{"normal", 1, -1},
		{"normal", 2, -1},
		{"normal", 5, -1},
		{"normal", 10, -1},
		{"abort", 2, 0},
		{"abort", 5, 2},
	}
	for k, data := range datas {
		t.Run(fmt.Sprintf("%s-%d", data.name, k), func(t *testing.T) {
			mc := gomock.NewController(t)
			tf := test.NewMockFuncCall(mc)
			ctx := &wrapContext{}
			for k := 0; k < data.num; k++ {
				index := k
				if data.abort >= 0 && k > data.abort {
					ctx.handlers = append(ctx.handlers, func(ctx Context) {})
					continue
				}
				tf.EXPECT().Call(index)
				ctx.handlers = append(ctx.handlers, func(ctx Context) {
					tf.Call(index)
					if data.abort >= 0 && data.abort == index {
						ctx.Abort()
					} else {
						ctx.Next(ctx)
					}
				})
			}
			ctx.Next(ctx)
		})
	}

}

type testLogCore struct {
	notify func(ent zapcore.Entry, fields []zap.Field)
}

func (c *testLogCore) Enabled(zapcore.Level) bool {
	return true
}
func (c *testLogCore) With([]zapcore.Field) zapcore.Core {
	return c
}
func (c *testLogCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, c)
}
func (c *testLogCore) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	c.notify(ent, fs)
	return nil
}
func (*testLogCore) Sync() error {
	return nil
}

func TestContext_Logger(t *testing.T) {
	mc := gomock.NewController(t)
	tf := test.NewMockFuncCall(mc)

	notifyCore := &testLogCore{}
	log := zaplog.NewLogger(zaplog.DEV, zap.New(notifyCore))
	ctx := &wrapContext{
		log: log,
	}

	fc := zap.Int("value", 1)

	// Normal
	tf.EXPECT().Call(fc)
	notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
		assert.EqualValues(t, "msg", ent.Message, "compare msg")
		assert.Equal(t, len(fields), int(3), "compare size")
		assert.Equal(t, zaplog.ZapLevelFieldDev, fields[0], "compare fields value")
		assert.Equal(t, "fname", fields[1].Key, "fname key check")
		assert.Equal(t, fc, fields[2], "compare fields value")
		tf.Call(fields[2])
	}

	le := ctx.NewEntry("funcName")
	le.Develop8("msg", fc)

	// If[Level]
	log.SetLogLevel(zaplog.DEBUG)
	tf.EXPECT().Call()
	notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
		assert.EqualValues(t, "msg", ent.Message, "compare msg")
		assert.Equal(t, len(fields), int(2), "compare size")
		assert.Equal(t, zaplog.ZapLevelFieldDebug, fields[0], "compare fields value")
		assert.Equal(t, "fname", fields[1].Key, "fname key check")
		//assert.Equal(t, fc, fields[2], "compare fields value")
		tf.Call()
	}

	le = ctx.NewEntry("funcName")
	le.IfDevelop8(fc)
	le.Debug7("msg")

	// WhenError 1
	log.SetLogLevel(zaplog.DEBUG)
	tf.EXPECT().Call()
	notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
		assert.EqualValues(t, "msg", ent.Message, "compare msg")
		assert.Equal(t, len(fields), int(2), "compare size")
		assert.Equal(t, zaplog.ZapLevelFieldDebug, fields[0], "compare fields value")
		assert.Equal(t, "fname", fields[1].Key, "fname key check")
		//assert.Equal(t, fc, fields[2], "compare fields value")
		tf.Call()
	}

	le = ctx.NewEntry("funcName")
	le.WhenErr(fc)
	le.Debug7("msg")

	// WhenError 2
	tf.EXPECT().Call()
	notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
		assert.EqualValues(t, "msg", ent.Message, "compare msg")
		assert.Equal(t, len(fields), int(3), "compare size")
		assert.Equal(t, zaplog.ZapLevelFieldErr, fields[0], "compare fields value")
		assert.Equal(t, "fname", fields[1].Key, "fname key check")
		assert.Equal(t, fc, fields[2], "compare fields value")
		tf.Call()
	}

	le = ctx.NewEntry("funcName")
	le.WhenErr(fc)
	le.Error3("msg")
}
