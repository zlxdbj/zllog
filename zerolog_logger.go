package zllog

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// ============================================================================
// ZerologLogger - 默认的 Zerolog 实现
// ============================================================================

// ZerologLogger 基于 Zerolog 的 Logger 接口实现
type ZerologLogger struct {
	logger        *zerolog.Logger
	enableCaller  bool
}

// NewZerologLogger 创建 Zerolog Logger 实例
func NewZerologLogger(logger *zerolog.Logger) *ZerologLogger {
	return &ZerologLogger{
		logger:       logger,
		enableCaller: true, // 默认启用 caller，后续可通过配置控制
	}
}

// getCaller 获取调用者位置信息（跳过库内部的调用帧）
// 返回格式：filename:line
func getCaller() string {
	// 尝试不同的调用栈深度
	for skip := 3; skip <= 6; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			continue
		}

		// 通过 pc 获取函数名
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// 跳过 zllog 包内部的调用
		funcName := fn.Name()
		// 如果函数名包含 "zllog."，说明还在库内部，继续查找
		if contains(funcName, "zllog.") {
			continue
		}

		// 获取文件名（不包含完整路径）
		shortFile := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' || file[i] == '\\' {
				shortFile = file[i+1:]
				break
			}
		}

		return fmt.Sprintf("%s:%d", shortFile, line)
	}

	return "unknown:0"
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// Info logs a message at INFO level
func (l *ZerologLogger) Info(ctx context.Context, module, message string, fields ...Field) {
	event := l.logger.Info()
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// Warn logs a message at WARN level
func (l *ZerologLogger) Warn(ctx context.Context, module, message string, fields ...Field) {
	event := l.logger.Warn()
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event = l.addFields(event, fields...)
	event.Msg(message)
}

// ============================================================================
// 格式化日志方法实现
// ============================================================================

// Debugf logs a formatted message at DEBUG level
func (l *ZerologLogger) Debugf(ctx context.Context, module, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Debug()
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}

// Infof logs a formatted message at INFO level
func (l *ZerologLogger) Infof(ctx context.Context, module, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Info()
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}

// Warnf logs a formatted message at WARN level
func (l *ZerologLogger) Warnf(ctx context.Context, module, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Warn()
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}

// Errorf logs a formatted message at ERROR level with error info
func (l *ZerologLogger) Errorf(ctx context.Context, module, format string, err error, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}

// ErrorWithCodef logs a formatted message at ERROR level with error code
func (l *ZerologLogger) ErrorWithCodef(ctx context.Context, module, format string, errorCode string, err error, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("error_code", errorCode)
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}

// Fatalf logs a formatted message at FATAL level and exits
func (l *ZerologLogger) Fatalf(ctx context.Context, module, format string, err error, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
	os.Exit(1)
}

// InfoWithRequestf INFO日志 + request_id + cost_ms (formatted)
func (l *ZerologLogger) InfoWithRequestf(ctx context.Context, module, format string, requestID string, costMs int64, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	event := l.logger.Info()
	if requestID != "" {
		event = event.Str("request_id", requestID)
	}
	if costMs > 0 {
		event = event.Int64("cost_ms", costMs)
	}
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}

// ErrorWithRequestf ERROR日志 + request_id + cost_ms (formatted)
func (l *ZerologLogger) ErrorWithRequestf(ctx context.Context, module, format string, requestID string, err error, costMs int64, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
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
	if l.enableCaller {
		event = event.Str("caller", getCaller())
	}
	event = event.Str("trace_id", GetOrCreateTraceID(ctx))
	event = event.Str("module", module)
	event.Msg(message)
}
