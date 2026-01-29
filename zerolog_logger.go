package zllog

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ============================================================================
// ZerologLogger - 默认的 Zerolog 实现
// ============================================================================

// ZerologLogger 基于 Zerolog 的 Logger 接口实现
type ZerologLogger struct {
	logger *zerolog.Logger
}

// NewZerologLogger 创建 Zerolog Logger 实例
func NewZerologLogger(logger *zerolog.Logger) *ZerologLogger {
	return &ZerologLogger{logger: logger}
}

// addFields 将自定义字段添加到日志事件
func (l *ZerologLogger) addFields(event *zerolog.Event, fields ...Field) *zerolog.Event {
	for _, field := range fields {
		switch v := field.Value.(type) {
		case string:
			event = event.Str(field.Key, v)
		case int:
			event = event.Int(field.Key, v)
		case int8:
			event = event.Int8(field.Key, v)
		case int16:
			event = event.Int16(field.Key, v)
		case int32:
			event = event.Int32(field.Key, v)
		case int64:
			event = event.Int64(field.Key, v)
		case uint:
			event = event.Uint(field.Key, v)
		case uint8:
			event = event.Uint8(field.Key, v)
		case uint16:
			event = event.Uint16(field.Key, v)
		case uint32:
			event = event.Uint32(field.Key, v)
		case uint64:
			event = event.Uint64(field.Key, v)
		case float32:
			event = event.Float32(field.Key, v)
		case float64:
			event = event.Float64(field.Key, v)
		case bool:
			event = event.Bool(field.Key, v)
		case time.Time:
			event = event.Time(field.Key, v)
		case time.Duration:
			event = event.Dur(field.Key, v)
		case error:
			event = event.Err(v)
		case []byte:
			event = event.RawJSON(field.Key, v)
		case []Field:
			// 处理 Dict 和 Array 类型
			if len(v) > 0 {
				// 判断是 Dict 还是 Array
				// 这里简化处理，默认使用 Array
				// 如果需要 Dict，可以使用 zerolog.Dict()
				event = event.Array(field.Key, zerolog.Arr())
			}
		default:
			event = event.Interface(field.Key, v)
		}
	}
	return event
}

// Debug logs a message at DEBUG level
func (l *ZerologLogger) Debug(ctx context.Context, module, message string, fields ...Field) {
	event := l.logger.Debug()
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// Info logs a message at INFO level
func (l *ZerologLogger) Info(ctx context.Context, module, message string, fields ...Field) {
	event := l.logger.Info()
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// Warn logs a message at WARN level
func (l *ZerologLogger) Warn(ctx context.Context, module, message string, fields ...Field) {
	event := l.logger.Warn()
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// Error logs a message at ERROR level with error info
func (l *ZerologLogger) Error(ctx context.Context, module, message string, err error, fields ...Field) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// ErrorWithCode logs a message at ERROR level with error code
func (l *ZerologLogger) ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...Field) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	event = event.Str("error_code", errorCode)
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// Fatal logs a message at FATAL level and exits
func (l *ZerologLogger) Fatal(ctx context.Context, module, message string, err error, fields ...Field) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
	os.Exit(1)
}

// InfoWithRequest INFO日志 + request_id + cost_ms
func (l *ZerologLogger) InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...Field) {
	event := l.logger.Info()
	if requestID != "" {
		event = event.Str("request_id", requestID)
	}
	if costMs > 0 {
		event = event.Int64("cost_ms", costMs)
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// ErrorWithRequest ERROR日志 + request_id + cost_ms
func (l *ZerologLogger) ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...Field) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	if requestID != "" {
		event = event.Str("request_id", requestID)
	}
	if costMs > 0 {
		event = event.Int64("cost_ms", costMs)
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}
