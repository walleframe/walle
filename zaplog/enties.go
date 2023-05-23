package zaplog

import (
	"time"

	"go.uber.org/zap"
)

// LogEntities 日志条目
type LogEntities struct {
	logger *zap.Logger
	fields []zap.Field
	start  *time.Time
	err    *LogEntities
}

// ResetTime 记录时间,重置时间
func (entity *LogEntities) ResetTime() *LogEntities {
	now := time.Now()
	entity.start = &now
	return entity
}

// ClearTime 清空时间，不再记录时间
func (entity *LogEntities) ClearTime() *LogEntities {
	entity.start = nil 
	return entity
}

func (entity *LogEntities) Fields() []zap.Field {
	return entity.fields
}

// Debug 输出一条日志
func (entity *LogEntities) Debug(msg string, fields ...zap.Field) {
	if ce := entity.logger.Check(zap.DebugLevel, msg); ce != nil {
		if entity.start != nil {
			fields = append(fields,
				zap.Int64("usems", time.Since(*entity.start).Milliseconds()),
				// zap.Duration("use", time.Since(*entity.start)),
			)
		}
		fields = append(fields, entity.fields...)
		ce.Write(fields...)
	}
}

// Info 输出一条日志
func (entity *LogEntities) Info(msg string, fields ...zap.Field) {
	if ce := entity.logger.Check(zap.InfoLevel, msg); ce != nil {
		if entity.start != nil {
			fields = append(fields,
				zap.Int64("usems", time.Since(*entity.start).Milliseconds()),
				// zap.Duration("use", time.Since(*entity.start)),
			)
		}
		fields = append(fields, entity.fields...)
		ce.Write(fields...)
	}
}

// Warn 输出一条日志
func (entity *LogEntities) Warn(msg string, fields ...zap.Field) {
	if ce := entity.logger.Check(zap.WarnLevel, msg); ce != nil {
		if entity.start != nil {
			fields = append(fields,
				zap.Int64("usems", time.Since(*entity.start).Milliseconds()),
				// zap.Duration("use", time.Since(*entity.start)),
			)
		}
		fields = append(fields, entity.fields...)
		ce.Write(fields...)
	}
}

// Error 输出一条日志
func (entity *LogEntities) Error(msg string, fields ...zap.Field) {
	if ce := entity.logger.Check(zap.ErrorLevel, msg); ce != nil {
		if entity.start != nil {
			fields = append(fields,
				zap.Int64("usems", time.Since(*entity.start).Milliseconds()),
				// zap.Duration("use", time.Since(*entity.start)),
			)
		}
		fields = append(fields, entity.fields...)
		// 追加错误日志
		if entity.err != nil && len(entity.err.fields) > 0 {
			fields = append(fields, entity.err.fields...)
		}
		// ce.Write(entity.fields...)
		ce.Write(fields...)
	}
}

// Check 检测错误
func (entity *LogEntities) Check(err error, msg string, fields ...zap.Field) {
	if err == nil {
		return
	}
	entity.Error(msg, append(fields, zap.Error(err))...)
}

// Must 日志打印一定会追加的字段
func (entity *LogEntities) Must() *LogFields {
	return &LogFields{
		entity: entity,
	}
}

// IfDebug Debug级日志
func (entity *LogEntities) IfDebug() *LogFields {
	if !entity.logger.Core().Enabled(zap.DebugLevel) {
		return nil
	}
	return &LogFields{
		entity: entity,
	}
}

// IfInfo Info级日志
func (entity *LogEntities) IfInfo() *LogFields {
	if !entity.logger.Core().Enabled(zap.InfoLevel) {
		return nil
	}
	return &LogFields{
		entity: entity,
	}
}

// IfWarn Warn级日志
func (entity *LogEntities) IfWarn() *LogFields {
	if !entity.logger.Core().Enabled(zap.WarnLevel) {
		return nil
	}
	return &LogFields{
		entity: entity,
	}
}

// IfError Error级日志
func (entity *LogEntities) IfError() *LogFields {
	if !entity.logger.Core().Enabled(zap.ErrorLevel) {
		return nil
	}
	return &LogFields{
		entity: entity,
	}
}

// WhenErr 如果发生错误才打印的错误日志
func (entity *LogEntities) WhenErr() *LogFields {
	if entity.err == nil {
		entity.err = &LogEntities{}
	}
	return &LogFields{
		entity: entity.err,
	}
}

