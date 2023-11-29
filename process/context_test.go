package process

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/walleframe/walle/testpkg"
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
			tf := testpkg.NewMockFuncCall(mc)
			ctx := &WrapContext{}
			for k := 0; k < data.num; k++ {
				index := k
				if data.abort >= 0 && k > data.abort {
					ctx.Handlers = append(ctx.Handlers, func(ctx Context) {})
					continue
				}
				tf.EXPECT().Call(index)
				ctx.Handlers = append(ctx.Handlers, func(ctx Context) {
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
