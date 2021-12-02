package zaplog

import (
	"go.uber.org/zap"
)

// Logger 日志对象封装
type Logger struct {
	log *zap.Logger
}

// NewLogger 新建日志对象
func NewLogger(logger *zap.Logger) *Logger {
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return &Logger{log: logger}
}

// Logger 获取原始日志接口
func (log *Logger) Logger() *zap.Logger {
	return log.log
}

// With 新建日志接口，并附加到每次日志内
func (log *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		log: log.log.With(fields...),
	}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (log *Logger) Named(name string) *Logger {
	return &Logger{
		log: log.log.Named(name),
	}
}

// New 新建日志实体(默认不记录时间)
func (log *Logger) New(funcName string) *LogEntities {
	return &LogEntities{
		logger: log.log,
		fields: []zap.Field{zap.String("func", funcName)},
	}
}

// Logic 逻辑层默认日志接口
var Logic *Logger

// Frame 框架层默认日志接口
var Frame *Logger

func init() {
	pcfg := zap.NewProductionConfig()
	pcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	debug, _ := pcfg.Build()
	Logic = NewLogger(debug)

	errLog, _ := zap.NewProduction(zap.IncreaseLevel(zap.ErrorLevel))
	Frame = NewLogger(errLog)
}
