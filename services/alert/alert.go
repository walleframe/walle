package alert

import (
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

type AlertLevel int8

const (
	AlertLevelNone AlertLevel = iota
	// AlertLevelWarning 警告信息 不是错误信息
	AlertLevelWarning
	// AlertLevelError 错误告警 需要处理，但不是很紧急
	AlertLevelError
	// AlertLevelCritical 关键信息告警
	AlertLevelCritical
	// AlertLevelEmergency 紧急告警 必须马上处理
	AlertLevelEmergency
)

type AlertSupport interface {
	Alert(lv AlertLevel, id int64, service, module, title, msg string, ext ...string)
}

var (
	backend AlertSupport
	service string
)

// SetupAlert 安装告警后端支持
func SetupAlert(svc string, alert AlertSupport) {
	backend = alert
	service = svc
}

func Alert(lv AlertLevel, id int64, module, title, msg string, ext ...string) {
	if backend == nil {
		zaplog.GetFrameLogger().New("alert.Alert").Debug("alert not setup",
			zap.Int8("lv", int8(lv)), zap.Int64("id", id), zap.String("service", service),
			zap.String("module", module), zap.String("title", title), zap.String("msg", msg),
			zap.Strings("ext", ext),
		)
		return
	}
	backend.Alert(lv, id, service, module, title, msg, ext...)
}

// Warning 警告信息 不是错误信息
func Warning(module, title, msg string) {
	if backend == nil {
		zaplog.GetFrameLogger().New("alert.Warning").Debug("alert not setup",
			zap.String("service", service), zap.String("module", module),
			zap.String("title", title), zap.String("msg", msg),
		)
		return
	}
	backend.Alert(AlertLevelWarning, 0, service, module, title, msg)
}

// Error 错误告警 需要处理，但不是很紧急
func Error(module, title, msg string) {
	if backend == nil {
		zaplog.GetFrameLogger().New("alert.Error").Debug("alert not setup",
			zap.String("service", service), zap.String("module", module),
			zap.String("title", title), zap.String("msg", msg),
		)
		return
	}
	backend.Alert(AlertLevelError, 0, service, module, title, msg)
}

// Critical 关键信息告警
func Critical(module, title, msg string) {
	if backend == nil {
		zaplog.GetFrameLogger().New("alert.Critical").Debug("alert not setup",
			zap.String("service", service), zap.String("module", module),
			zap.String("title", title), zap.String("msg", msg),
		)
		return
	}
	backend.Alert(AlertLevelCritical, 0, service, module, title, msg)
}

// Emergency 紧急告警 必须马上处理
func Emergency(module, title, msg string) {
	if backend == nil {
		zaplog.GetFrameLogger().New("alert.Emergency").Debug("alert not setup",
			zap.String("service", service), zap.String("module", module),
			zap.String("title", title), zap.String("msg", msg),
		)
		return
	}
	backend.Alert(AlertLevelEmergency, 0, service, module, title, msg)
}
