package testpkg

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TestLogCore struct {
	Notify func(ent zapcore.Entry, fields []zap.Field)
}

func (c *TestLogCore) Enabled(zapcore.Level) bool {
	return true
}
func (c *TestLogCore) With([]zapcore.Field) zapcore.Core {
	return c
}
func (c *TestLogCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, c)
}
func (c *TestLogCore) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	c.Notify(ent, fs)
	return nil
}
func (*TestLogCore) Sync() error {
	return nil
}



type EmtpyLogWriter struct {
}

func (EmtpyLogWriter) Write(in []byte) (int, error) {
	return len(in), nil
}

func (EmtpyLogWriter) Sync() error {
	return nil
}

func (EmtpyLogWriter) Close() error {
	return nil
}
