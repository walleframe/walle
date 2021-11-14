package zaplog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level new log level
type Level int8

const (
	// 0 EMERG （紧急）：会导致主机系统不可用的情况
	// 0 EMERG (emergency): a situation that will cause the host system to be unavailable
	EMERG Level = iota // zap.ERROR
	// 1 ALERT （警告）：必须马上采取措施解决的问题
	// 1 ALERT: problems that must be resolved immediately
	ALERT // zap.ERROR
	// 2 CRIT （严重）：比较严重的情况.程序异常. 影响正常处理业务. 必须马上处理.
	// 2 CRIT (serious): a more serious situation
	// The program is abnormal. It affects the normal processing business.
	// It must be dealt with immediately.
	CRIT // zap.ERROR
	// 3: 错误日志. 程序处理出错.
	// Error log. Program processing error.
	ERR // zap.ERROR
	// 4 WARNING（提醒）：可能会影响系统功能的事件
	// 4 WARNING: events that may affect the function of the system
	// 4: 异常信息. 不符合预期的日志. 但是不影响业务流程. -- 开发阶段需要注意的.线上阶段可忽略的.
	// Abnormal information. Logs that do not meet expectations.
	// But do not affect the business process. - Need to pay attention to during the
	// development phase. Can be ignored in the online phase.
	WARNING // zap.WARNING
	// 5: 关键日志. 重要信息,关键操作日志.  -- 可以理解成需要入库的日志
	// 5: Critical log. Important information, critical operation log.
	// - It can be understood as a log that needs to be stored
	INFO // zap.INFO
	// 6 NOTICE （注意）：不会影响系统但值得注意
	// 6 NOTICE: will not affect the system but it is worth noting
	NOTICE // zap.INFO
	// 7: 流水日志,调试信息. 用于程序调试.
	// Flow log, debugging information. Used for program debugging.
	DEBUG // zap.DEBUG
	// 8: 日志明细. 所有操作日志,组件操作日志
	// Log details. All operation logs, component operation logs
	DEV // zap.DEBUG
)

// ModifyZapConfig must set level key empty.
func ModifyZapConfig(cfg *zap.Config) {
	cfg.EncoderConfig.LevelKey = ""
	//if cfg.EncoderConfig.EncodeTime == nil {
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//}
	return
}

// Default default logger
var Default *Logger

func init() {
	log, err := NewLoggerWithCfg(DEBUG, zap.NewProductionConfig(), zap.AddStacktrace(zap.WarnLevel))
	if err != nil {
		panic(err)
	}
	Default = log
}

// Logger Custom Level Logger
type Logger struct {
	zaplog *zap.Logger
	lv     *Level
}

// NewLoggerWithCfg new zap.Logger by Config
func NewLoggerWithCfg(lv Level, cfg zap.Config, opts ...zap.Option) (log *Logger, err error) {
	ModifyZapConfig(&cfg)
	logger, err := cfg.Build(opts...)
	if err != nil {
		return
	}
	logger = logger.WithOptions(
		zap.AddCallerSkip(1),
	)
	log = &Logger{
		zaplog: logger,
		lv:     &lv,
	}
	return
}

// NewLogger new Logger by exists zap.Logger. you must set EncoderConfig.LevelKey empty
func NewLogger(lv Level, logger *zap.Logger) (log *Logger) {
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return &Logger{
		zaplog: logger,
		lv:     &lv,
	}
}

func (log *Logger) SetLogLevel(lv Level) {
	*log.lv = lv
}

func (log *Logger) GetLogLevel() Level {
	return *log.lv
}

func (log *Logger) Enabled(lv Level) bool {
	if *log.lv < lv {
		return false
	}
	switch lv {
	case EMERG, ALERT, CRIT, ERR:
		return log.zaplog.Core().Enabled(zap.ErrorLevel)
	case WARNING:
		return log.zaplog.Core().Enabled(zap.WarnLevel)
	case INFO, NOTICE:
		return log.zaplog.Core().Enabled(zap.InfoLevel)
	case DEBUG, DEV:
		return log.zaplog.Core().Enabled(zap.DebugLevel)
	}
	return false
}

func (log *Logger) Named(name string) *Logger {
	nlog := &Logger{
		zaplog: log.zaplog.Named(name),
		lv:     log.lv,
	}
	return nlog
}

func (log *Logger) With(fields ...zap.Field) *Logger {
	nlog := &Logger{
		zaplog: log.zaplog.With(fields...),
		lv:     log.lv,
	}
	return nlog
}

func (log *Logger) WithOptions(opts ...zap.Option) *Logger {
	nlog := &Logger{
		zaplog: log.zaplog.WithOptions(opts...),
		lv:     log.lv,
	}
	return nlog
}

// Emerg  (emergency): a situation that will cause the host system to be unavailable
func (log *Logger) Emerg0(msg string, fields ...zap.Field) {
	if !log.Enabled(EMERG) {
		return
	}
	ce := log.zaplog.Check(zap.ErrorLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// Alert problems that must be resolved immediately
func (log *Logger) Alert1(msg string, fields ...zap.Field) {
	if !log.Enabled(ALERT) {
		return
	}
	ce := log.zaplog.Check(zap.ErrorLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// CRIT (serious): a more serious situation
func (log *Logger) Crit2(msg string, fields ...zap.Field) {
	if !log.Enabled(CRIT) {
		return
	}
	ce := log.zaplog.Check(zap.ErrorLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

func (log *Logger) Error3(msg string, fields ...zap.Field) {
	if !log.Enabled(ERR) {
		return
	}
	ce := log.zaplog.Check(zap.ErrorLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// WARNING: events that may affect the function of the system
func (log *Logger) Warn4(msg string, fields ...zap.Field) {
	if !log.Enabled(WARNING) {
		return
	}
	ce := log.zaplog.Check(zap.WarnLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

func (log *Logger) Info5(msg string, fields ...zap.Field) {
	if !log.Enabled(INFO) {
		return
	}
	ce := log.zaplog.Check(zap.InfoLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// NOTICE: will not affect the system but it is worth noting
func (log *Logger) Notice6(msg string, fields ...zap.Field) {
	if !log.Enabled(NOTICE) {
		return
	}
	ce := log.zaplog.Check(zap.InfoLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// Debug7 Flow log, debugging information. Used for program debugging.
func (log *Logger) Debug7(msg string, fields ...zap.Field) {
	if !log.Enabled(DEBUG) {
		return
	}
	ce := log.zaplog.Check(zap.DebugLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// Develop8 Log details. All operation logs, component operation logs
func (log *Logger) Develop8(msg string, fields ...zap.Field) {
	if !log.Enabled(DEV) {
		return
	}
	ce := log.zaplog.Check(zap.DebugLevel, msg)
	if ce == nil {
		return
	}
	ce.Write(fields...)
	return
}

// func (log *Logger) check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
// 	// check must always be called directly by a method in the Logger interface
// 	// (e.g., Check, Info, Fatal).
// 	const callerSkipOffset = 2

// 	// Check the level first to reduce the cost of disabled log calls.
// 	// Since Panic and higher may exit, we skip the optimization for those levels.
// 	if lvl < zapcore.DPanicLevel && !log.core.Enabled(lvl) {
// 		return nil
// 	}

// 	// Create basic checked entry thru the core; this will be non-nil if the
// 	// log message will actually be written somewhere.
// 	ent := zapcore.Entry{
// 		LoggerName: log.name,
// 		Time:       log.clock.Now(),
// 		Level:      lvl,
// 		Message:    msg,
// 	}
// 	ce := log.core.Check(ent, nil)
// 	willWrite := ce != nil

// 	// Set up any required terminal behavior.
// 	switch ent.Level {
// 	case zapcore.PanicLevel:
// 		ce = ce.Should(ent, zapcore.WriteThenPanic)
// 	case zapcore.FatalLevel:
// 		onFatal := log.onFatal
// 		// Noop is the default value for CheckWriteAction, and it leads to
// 		// continued execution after a Fatal which is unexpected.
// 		if onFatal == zapcore.WriteThenNoop {
// 			onFatal = zapcore.WriteThenFatal
// 		}
// 		ce = ce.Should(ent, onFatal)
// 	case zapcore.DPanicLevel:
// 		if log.development {
// 			ce = ce.Should(ent, zapcore.WriteThenPanic)
// 		}
// 	}

// 	// Only do further annotation if we're going to write this message; checked
// 	// entries that exist only for terminal behavior don't benefit from
// 	// annotation.
// 	if !willWrite {
// 		return ce
// 	}

// 	// Thread the error output through to the CheckedEntry.
// 	ce.ErrorOutput = log.errorOutput
// 	if log.addCaller {
// 		frame, defined := getCallerFrame(log.callerSkip + callerSkipOffset)
// 		if !defined {
// 			fmt.Fprintf(log.errorOutput, "%v Logger.check error: failed to get caller\n", ent.Time.UTC())
// 			log.errorOutput.Sync()
// 		}

// 		ce.Entry.Caller = zapcore.EntryCaller{
// 			Defined:  defined,
// 			PC:       frame.PC,
// 			File:     frame.File,
// 			Line:     frame.Line,
// 			Function: frame.Function,
// 		}
// 	}
// 	if log.addStack.Enabled(ce.Entry.Level) {
// 		ce.Entry.Stack = StackSkip("", log.callerSkip+callerSkipOffset).String
// 	}

// 	return ce
// }

func GetLevelField(lv Level) zap.Field {
	switch lv {
	case EMERG:
		return ZapLevelFieldEmerg
	case ALERT:
		return ZapLevelFieldAlert
	case CRIT:
		return ZapLevelFieldCrit
	case ERR:
		return ZapLevelFieldErr
	case WARNING:
		return ZapLevelFieldWarning
	case INFO:
		return ZapLevelFieldInfo
	case NOTICE:
		return ZapLevelFieldNotice
	case DEBUG:
		return ZapLevelFieldDebug
	case DEV:
		return ZapLevelFieldDev
	default:
		if lv > DEV {
			return ZapLevelFieldDev
		} else {
			return ZapLevelFieldEmerg
		}
	}
}

var (
	_level_key           = "log_lv"
	ZapLevelFieldEmerg   = zap.Int8(_level_key, int8(EMERG))
	ZapLevelFieldAlert   = zap.Int8(_level_key, int8(ALERT))
	ZapLevelFieldCrit    = zap.Int8(_level_key, int8(CRIT))
	ZapLevelFieldErr     = zap.Int8(_level_key, int8(ERR))
	ZapLevelFieldWarning = zap.Int8(_level_key, int8(WARNING))
	ZapLevelFieldInfo    = zap.Int8(_level_key, int8(INFO))
	ZapLevelFieldNotice  = zap.Int8(_level_key, int8(NOTICE))
	ZapLevelFieldDebug   = zap.Int8(_level_key, int8(DEBUG))
	ZapLevelFieldDev     = zap.Int8(_level_key, int8(DEV))
)
