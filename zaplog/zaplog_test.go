package zaplog

import (
	"fmt"
	"testing"

	"github.com/aggronmagi/walle/internal/util/test"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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


func GetLogFuncs(log *LogEntities) (arr []func(msg string, fields ...zap.Field)) {
	arr = append(arr,
		log.Debug,
		log.Info,
		log.Warn,
		log.Error,
	)
	return
}

func TestLogger_Write(t *testing.T) {

	notifyCore := &testLogCore{}
	log := NewLogger(zap.New(notifyCore))

	lfs := GetLogFuncs(log.New("test"))
	fc := zap.Int("value", 1)
	ff := zap.String("func", "test")
	for k, w := range lfs {
		t.Run(fmt.Sprintf("level-%d", k), func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			tf := test.NewMockFuncCall(mc)
			tf.EXPECT().Call()
			notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
				assert.EqualValues(t, "msg", ent.Message, "compare msg")
				assert.Equal(t, len(fields), int(2), "compare size")
				assert.EqualValues(t, ff, fields[1], "compare func fields")
				assert.Equal(t, fc, fields[0], "compare fields value")
				tf.Call()
			}
			w("msg", fc)
		})
	}
}
