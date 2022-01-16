package zaplog

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"testing"

	"github.com/aggronmagi/walle/testpkg"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getLogFuncs(log *LogEntities) (arr []func(msg string, fields ...zap.Field)) {
	arr = append(arr,
		log.Debug,
		log.Info,
		log.Warn,
		log.Error,
	)
	return
}

func TestLogger_Write(t *testing.T) {
	notifyCore := &testpkg.TestLogCore{}
	log := NewLogger(zap.New(notifyCore))

	lfs := getLogFuncs(log.New("test"))
	fc := zap.Int("value", 1)
	ff := zap.String("func", "test")
	for k, w := range lfs {
		t.Run(fmt.Sprintf("level-%d", k), func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			tf := testpkg.NewMockFuncCall(mc)
			tf.EXPECT().Call()
			notifyCore.Notify = func(ent zapcore.Entry, fields []zap.Field) {
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

func TestZapLogger(t *testing.T) {
	runtime.MemProfileRate = 1
	zl, _ := zap.NewDevelopment()
	zs := zl.Sugar()
	zl.Debug("abc", zap.String("a", "v"), zap.Int("b", 10))
	zl.Debug("abc", zap.String("a", "v"), zap.Int("b", 10))

	zs.Debug("abc", "v", 10)
	zs.Debug("abc", "v", 10)
	zs.Debugf("abc %s %d", "v", 10)
	zs.Debugf("abc %s %d", "v", 10)

	println("------------------------------")
	ll := logrus.New()
	ll.Out = os.Stdout
	ll.Level = logrus.DebugLevel
	ll.Debug("abc ", "v", 10)
	ll.Debug("abc ", "v", 10)
	ll.Debugf("abc %s %d", "v", 10)
	ll.Debugf("abc %s %d", "v", 10)
}

func BenchmarkLoggers(b *testing.B) {

	zap.RegisterSink("empty", func(u *url.URL) (zap.Sink, error) {
		return testpkg.EmtpyLogWriter{}, nil
	})
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"empty://xx"}
	cfg.ErrorOutputPaths = []string{"empty://xx"}
	zlog, _ := cfg.Build()

	llog := logrus.New()
	llog.Out = testpkg.EmtpyLogWriter{}
	llog.Level = logrus.DebugLevel
	llog.ReportCaller = false

	b.Run("logrus     ", func(b *testing.B) {
		log := llog
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Debugf("abc %s %d", "v", 10)
		}
	})
	b.Run("logrus2     ", func(b *testing.B) {
		log := llog
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Debug("abc", "v", 10)
		}
	})
	b.Run("zap-sugar   ", func(b *testing.B) {
		log := zlog.Sugar()
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Infof("abc %s %d", "v", 10)
		}
	})
	b.Run("zap-sugar2   ", func(b *testing.B) {
		log := zlog.Sugar()
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Info("abc", "v", 10)
		}
	})

	b.Run("zap-struct   ", func(b *testing.B) {
		log := zlog
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Info("abc", zap.String("a", "v"), zap.Int("d", 10))
		}
	})
	cfg2 := zap.NewDevelopmentConfig()
	cfg2.OutputPaths = []string{"empty://xx"}
	cfg2.ErrorOutputPaths = []string{"empty://xx"}
	zlog2, _ := cfg2.Build()
	b.Run("zap-sugar(dev)", func(b *testing.B) {
		log := zlog2.Sugar()
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Infof("abc %s %d", "v", 10)
		}
	})
	b.Run("zap-sugar2(dev)", func(b *testing.B) {
		log := zlog2.Sugar()
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Info("abc", "v", 10)
		}
	})

	b.Run("zap-struct(dev)", func(b *testing.B) {
		log := zlog2
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Info("abc", zap.String("a", "v"), zap.Int("d", 10))
		}
	})
}
