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

func TestLogger_New_1(t *testing.T) {
	tf := test.NewMockFuncCall(gomock.NewController(t))
	tf.EXPECT().Call()
	notifyCore := &testLogCore{}
	log := NewLogger(DEV, zap.New(notifyCore))
	fc := zap.Int("value", 1)
	notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
		assert.EqualValues(t, "msg", ent.Message, "compare msg")
		assert.Equal(t, len(fields), int(1), "compare size")
		assert.Equal(t, fc, fields[0], "compare fields value")
		tf.Call()
	}
	log.Develop8("msg", fc)
}

func TestLogger_New_2(t *testing.T) {
	tf := test.NewMockFuncCall(gomock.NewController(t))
	tf.EXPECT().Call()
	notifyCore := &testLogCore{}
	log, err := NewLoggerWithCfg(DEV, zap.NewDevelopmentConfig(),
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return notifyCore
		}),
	)
	assert.Nil(t, err, "create logger")
	fc := zap.Int("value", 1)
	notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
		assert.EqualValues(t, "msg", ent.Message, "compare msg")
		assert.Equal(t, len(fields), int(1), "compare size")
		assert.Equal(t, fc, fields[0], "compare fields value")
		tf.Call()
	}
	log.Develop8("msg", fc)
}

func TestLogger_Level(t *testing.T) {
	notifyCore := &testLogCore{}
	log := NewLogger(DEV, zap.New(notifyCore))
	for lv := EMERG; lv <= DEV; lv++ {
		log.SetLogLevel(lv)
		for c := EMERG; c <= DEV; c++ {
			assert.Equal(t, c <= lv, log.Enabled(c), "log level check")
		}
	}
}

func GetLogFuncs(log *Logger) (arr []func(msg string, fields ...zap.Field)) {
	arr = append(arr,
		log.Emerg0,
		log.Alert1,
		log.Crit2,
		log.Error3,
		log.Warn4,
		log.Info5,
		log.Notice6,
		log.Debug7,
		log.Develop8,
	)
	return
}

func TestLogger_Write(t *testing.T) {

	notifyCore := &testLogCore{}
	log := NewLogger(DEV, zap.New(notifyCore))

	lfs := GetLogFuncs(log)
	fc := zap.Int("value", 1)
	for k, w := range lfs {
		t.Run(fmt.Sprintf("level-%d", k), func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			tf := test.NewMockFuncCall(mc)
			tf.EXPECT().Call()
			notifyCore.notify = func(ent zapcore.Entry, fields []zap.Field) {
				assert.EqualValues(t, "msg", ent.Message, "compare msg")
				assert.Equal(t, len(fields), int(1), "compare size")
				assert.Equal(t, fc, fields[0], "compare fields value")
				tf.Call()
			}
			w("msg", fc)
		})
	}
}
