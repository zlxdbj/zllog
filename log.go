package zllog

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ============================================================================
// 中林日志组件（ZLLog）- 基于 Zerolog 的结构化日志
// 支持复制到其他项目直接使用
// ============================================================================

var (
	globalLogger zerolog.Logger
	onceInit     sync.Once
	serviceName  string
	envName      string
	hostName     string

	// ✅ 全局 TraceID Provider（解耦追踪系统）
	globalTraceIDProvider TraceIDProvider

	// ✅ 全局 Logger 接口（支持自定义实现）
	globalLoggerImpl Logger
)

// ============================================================================
// TraceIDProvider 接口 - 解耦日志和追踪系统
// ============================================================================

// TraceIDProvider 定义 trace_id 提供者接口
// 任何追踪系统（SkyWalking、Jaeger、OpenTelemetry）都可以实现此接口
type TraceIDProvider interface {
	// GetTraceID 从 context 中提取 trace_id
	// 如果 context 中没有 trace 信息，返回空字符串
	GetTraceID(ctx context.Context) string

	// Name 返回追踪系统的名称（用于日志记录）
	Name() string
}

// RegisterTraceIDProvider 注册 trace_id 提供者
// 可以在运行时动态注册不同的追踪系统
func RegisterTraceIDProvider(provider TraceIDProvider) {
	globalTraceIDProvider = provider
}

// GetTraceIDProvider 获取当前注册的 trace_id 提供者
func GetTraceIDProvider() TraceIDProvider {
	return globalTraceIDProvider
}

// ============================================================================
// Logger 接口 - 支持自定义日志实现
// ============================================================================

// Logger 定义日志接口，用户可以实现此接口来使用自定义的日志库
// 默认提供基于 Zerolog 的实现（ZerologLogger）
type Logger interface {
	// Debug logs a message at DEBUG level
	Debug(ctx context.Context, module, message string, fields ...Field)

	// Info logs a message at INFO level
	Info(ctx context.Context, module, message string, fields ...Field)

	// Warn logs a message at WARN level
	Warn(ctx context.Context, module, message string, fields ...Field)

	// Error logs a message at ERROR level with error info
	Error(ctx context.Context, module, message string, err error, fields ...Field)

	// ErrorWithCode logs a message at ERROR level with error code
	ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...Field)

	// Fatal logs a message at FATAL level and exits
	Fatal(ctx context.Context, module, message string, err error, fields ...Field)

	// InfoWithRequest INFO日志 + request_id + cost_ms
	InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...Field)

	// ErrorWithRequest ERROR日志 + request_id + cost_ms
	ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...Field)
}

// SetLogger 设置自定义 Logger 实现
// 允许用户在运行时替换默认的日志实现
//
// 用法示例：
//   // 自定义 Logger 实现
//   type MyLogger struct {}
//   func (l *MyLogger) Debug(ctx context.Context, module, message string, fields ...zllog.Field) {
//       // 自定义实现
//   }
//   // ... 实现其他方法
//
//   // 注册自定义 Logger
//   zllog.SetLogger(&MyLogger{})
func SetLogger(logger Logger) {
	globalLoggerImpl = logger
}

// GetLogger 获取当前使用的 Logger 实现
func GetLogger() Logger {
	return globalLoggerImpl
}

// ============================================================================
// 日志配置
// ============================================================================

// LogConfig 日志配置
type LogConfig struct {
	// 必须字段
	ServiceName string // 服务名称
	Env         string // 环境：dev/test/prod
	LogLevel    string // 日志级别：DEBUG/INFO/WARN/ERROR/FATAL

	// 日志文件配置
	LogDir     string // 日志目录
	MaxSize    int    // 单个日志文件最大大小（MB）
	MaxBackups int    // 保留的历史日志文件个数
	MaxAge     int    // 保留历史日志文件的最大天数
	Compress   bool   // 是否压缩历史日志文件

	// 日期滚动配置
	EnableDailyRoll bool // 是否启用日期滚动（默认true）

	// 控制台输出配置
	EnableConsole     bool // 是否输出到控制台（开发环境建议true）
	ConsoleJSONFormat bool // 控制台是否使用JSON格式（false时使用彩色文本）
}

// DefaultConfig 返回默认配置（符合等保3最低要求）
func DefaultConfig(serviceName string) *LogConfig {
	return &LogConfig{
		ServiceName:      serviceName,
		Env:             "dev",
		LogLevel:        "INFO",
		LogDir:          "./logs",
		MaxSize:         100, // 100MB（单个文件最大大小）
		MaxBackups:      180, // 保留180个历史文件（配合每日切割，可保留180天）
		MaxAge:          180, // 保留180天（6个月，符合等保3对ERROR日志的最低要求）
		Compress:        true, // 启用压缩（等保3要求）
		EnableDailyRoll: true, // 启用日期滚动（每天切割）
		EnableConsole:   true, // 开发环境默认开启控制台输出
		ConsoleJSONFormat: false, // 控制台使用彩色文本格式（更友好）
	}
}

// ============================================================================
// 日志系统初始化
// ============================================================================

// InitLogger 初始化日志系统
// 自动查找配置文件，按优先级查找：
//   1. resource/log.yaml（独立配置文件）
//   2. resource/application.yaml（项目配置文件）
//   3. resource/application_{ENV}.yaml（环境配置）
//   4. 默认配置
func InitLogger() error {
	return InitLoggerWithConfigDir("resource")
}

// InitLoggerWithConfigDir 从指定目录初始化日志系统
func InitLoggerWithConfigDir(configDir string) error {
	loader := NewConfigLoader()
	loader.SetConfigDir(configDir)
	config := loader.LoadConfig()
	return InitLoggerWithConfig(config)
}

// InitLoggerWithConfig 使用指定配置初始化日志系统
// 这是推荐的初始化方式，完全可控
func InitLoggerWithConfig(config *LogConfig) error {
	var initErr error
	onceInit.Do(func() {
		// 保存全局配置
		serviceName = config.ServiceName
		envName = config.Env
		if h, err := os.Hostname(); err == nil {
			hostName = h
		} else {
			hostName = "unknown"
		}

		// 解析日志级别
		level, err := parseLevel(config.LogLevel)
		if err != nil {
			initErr = fmt.Errorf("invalid log level: %s", config.LogLevel)
			return
		}
		zerolog.SetGlobalLevel(level)

		// 设置时间格式为纳秒精度（更适合日志分析和高并发场景）
		zerolog.TimeFieldFormat = time.RFC3339Nano

		// 创建输出writers
		var writers []io.Writer

		// 文件输出
		logFile := createLogFileWriter(config)
		writers = append(writers, logFile)

		// 控制台输出
		if config.EnableConsole {
			consoleWriter := createConsoleWriter(config)
			writers = append(writers, consoleWriter)
		}

		// 多路输出（文件 + 控制台）
		multiWriter := zerolog.MultiLevelWriter(writers...)

		// 创建全局logger（添加基础字段）
		globalLogger = zerolog.New(multiWriter).
			Level(level).
			With().
			Timestamp().
			Str("service", serviceName).
			Str("env", config.Env).
			Str("host", hostName).
			Logger()

		// ✅ 创建默认的 ZerologLogger 实现
		globalLoggerImpl = NewZerologLogger(&globalLogger)

		// 打印初始化成功信息
		globalLogger.Info().
			Str("service", serviceName).
			Str("env", config.Env).
			Str("level", config.LogLevel).
			Str("dir", config.LogDir).
			Send()
	})

	return initErr
}

// InitLoggerFromFile 从指定文件初始化日志系统
// 支持两种格式：
//   1. log.yaml（直接格式）：service_name, env, level, dir...
//   2. application.yaml（嵌套格式）：logger.level, logger.dir...
func InitLoggerFromFile(filename string) error {
	loader := NewConfigLoader()

	// 判断文件类型
	if strings.Contains(filename, "log.yaml") {
		// 独立的 log.yaml
		v := viper.New()
		v.SetConfigFile(filename)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		config := loader.parseLogConfig(v)
		return InitLoggerWithConfig(config)
	} else {
		// application.yaml 或 application_{ENV}.yaml
		v := viper.New()
		v.SetConfigFile(filename)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		config := loader.parseLoggerConfig(v)
		return InitLoggerWithConfig(config)
	}
}

// parseLevel 解析日志级别字符串
func parseLevel(levelStr string) (zerolog.Level, error) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return zerolog.DebugLevel, nil
	case "INFO":
		return zerolog.InfoLevel, nil
	case "WARN", "WARNING":
		return zerolog.WarnLevel, nil
	case "ERROR":
		return zerolog.ErrorLevel, nil
	case "FATAL":
		return zerolog.FatalLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", levelStr)
	}
}

// createLogFileWriter 创建日志文件输出writer
func createLogFileWriter(config *LogConfig) io.Writer {
	// 确保日志目录存在
	os.MkdirAll(config.LogDir, 0755)

	// 日志文件路径
	logFilePath := filepath.Join(config.LogDir, "app.log")

	// 使用lumberjack进行日志轮转
	return &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    config.MaxSize,    // MB
		MaxBackups: config.MaxBackups, // 保留历史文件数
		MaxAge:     config.MaxAge,     // 天数
		Compress:   config.Compress,   // 压缩
	}
}

// createConsoleWriter 创建控制台输出writer
func createConsoleWriter(config *LogConfig) io.Writer {
	if config.ConsoleJSONFormat {
		// JSON格式（适合生产环境日志采集）
		return os.Stdout
	}

	// 彩色文本格式（开发环境友好）
	return zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    false,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			return fmt.Sprintf("[%s]", strings.ToUpper(i.(string)))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
	}
}

// GetGlobalLogger 获取全局logger实例
func GetGlobalLogger() *zerolog.Logger {
	return &globalLogger
}

// GetServiceName 获取服务名称
func GetServiceName() string {
	return serviceName
}

// GetEnvName 获取环境名称
func GetEnvName() string {
	return envName
}

// ============================================================================
// Trace ID 工具函数
// ============================================================================

// GetOrCreateTraceID 获取或创建 trace_id
// 1. 尝试从 context 获取 trace_id
// 2. 如果没有，自动生成一个新的 trace_id（用于定时任务、初始化等场景）
// 3. 生成的 trace_id 符合 W3C Trace Context 标准（32位十六进制字符）
func GetOrCreateTraceID(ctx context.Context) string {
	// 1. 尝试从 context 获取 trace_id
	if globalTraceIDProvider != nil {
		if traceID := globalTraceIDProvider.GetTraceID(ctx); traceID != "" {
			return traceID
		}
	}

	// 2. 如果没有 trace_id，自动生成一个符合 W3C 标准的 trace_id
	// 使用 hex 编码，性能优于 strings.Replace
	traceID := uuid.New()
	return hex.EncodeToString(traceID[:])
}

// ============================================================================
// 公共日志方法（必须传 Context）
// ============================================================================

// getLogger 获取当前 logger 实现（如果未设置则使用默认实现）
func getLogger() Logger {
	if globalLoggerImpl == nil {
		// 如果没有设置自定义实现，使用默认的 ZerologLogger
		return NewZerologLogger(&globalLogger)
	}
	return globalLoggerImpl
}

// Debug logs a message at DEBUG level
func Debug(ctx context.Context, module, message string, fields ...Field) {
	getLogger().Debug(ctx, module, message, fields...)
}

// Info logs a message at INFO level
func Info(ctx context.Context, module, message string, fields ...Field) {
	getLogger().Info(ctx, module, message, fields...)
}

// Warn logs a message at WARN level
func Warn(ctx context.Context, module, message string, fields ...Field) {
	getLogger().Warn(ctx, module, message, fields...)
}

// Error logs a message at ERROR level with error info
func Error(ctx context.Context, module, message string, err error, fields ...Field) {
	getLogger().Error(ctx, module, message, err, fields...)
}

// ErrorWithCode logs a message at ERROR level with error code
func ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...Field) {
	getLogger().ErrorWithCode(ctx, module, message, errorCode, err, fields...)
}

// Fatal logs a message at FATAL level and exits
func Fatal(ctx context.Context, module, message string, err error, fields ...Field) {
	getLogger().Fatal(ctx, module, message, err, fields...)
}

// ============================================================================
// 带请求追踪的日志方法
// ============================================================================

// InfoWithRequest INFO日志 + request_id + cost_ms
func InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...Field) {
	getLogger().InfoWithRequest(ctx, module, message, requestID, costMs, fields...)
}

// ErrorWithRequest ERROR日志 + request_id + cost_ms
func ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...Field) {
	getLogger().ErrorWithRequest(ctx, module, message, requestID, err, costMs, fields...)
}

