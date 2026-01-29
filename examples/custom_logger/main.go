package main

import (
	"context"
	"fmt"
	"os"

	"github.com/zlxdbj/zllog"
)

// ============================================================================
// 自定义 Logger 实现示例
// ============================================================================

// CustomLogger 自定义日志实现示例
// 可以实现为将日志发送到远程服务、写入自定义格式等
type CustomLogger struct {
	// 可以添加自定义字段，如：数据库连接、API 地址等
	serviceName string
}

// Debug 实现 Logger 接口的 Debug 方法
func (l *CustomLogger) Debug(ctx context.Context, module, message string, fields ...zllog.Field) {
	l.log("DEBUG", ctx, module, message, fields...)
}

// Info 实现 Logger 接口的 Info 方法
func (l *CustomLogger) Info(ctx context.Context, module, message string, fields ...zllog.Field) {
	l.log("INFO", ctx, module, message, fields...)
}

// Warn 实现 Logger 接口的 Warn 方法
func (l *CustomLogger) Warn(ctx context.Context, module, message string, fields ...zllog.Field) {
	l.log("WARN", ctx, module, message, fields...)
}

// Error 实现 Logger 接口的 Error 方法
func (l *CustomLogger) Error(ctx context.Context, module, message string, err error, fields ...zllog.Field) {
	msg := fmt.Sprintf("[%s] %s: %s", module, message, err)
	if len(fields) > 0 {
		msg += " " + formatFields(fields)
	}
	fmt.Printf("ERROR: %s\n", msg)
}

// ErrorWithCode 实现 Logger 接口的 ErrorWithCode 方法
func (l *CustomLogger) ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...zllog.Field) {
	msg := fmt.Sprintf("[%s] %s [code:%s]: %s", module, message, errorCode, err)
	if len(fields) > 0 {
		msg += " " + formatFields(fields)
	}
	fmt.Printf("ERROR: %s\n", msg)
}

// Fatal 实现 Logger 接口的 Fatal 方法
func (l *CustomLogger) Fatal(ctx context.Context, module, message string, err error, fields ...zllog.Field) {
	msg := fmt.Sprintf("[%s] %s: %s", module, message, err)
	if len(fields) > 0 {
		msg += " " + formatFields(fields)
	}
	fmt.Printf("FATAL: %s\n", msg)
	os.Exit(1)
}

// InfoWithRequest 实现 Logger 接口的 InfoWithRequest 方法
func (l *CustomLogger) InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...zllog.Field) {
	msg := fmt.Sprintf("[%s] %s [request:%s] [cost:%dms]", module, message, requestID, costMs)
	if len(fields) > 0 {
		msg += " " + formatFields(fields)
	}
	fmt.Printf("INFO: %s\n", msg)
}

// ErrorWithRequest 实现 Logger 接口的 ErrorWithRequest 方法
func (l *CustomLogger) ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...zllog.Field) {
	msg := fmt.Sprintf("[%s] %s [request:%s] [cost:%dms]: %s", module, message, requestID, costMs, err)
	if len(fields) > 0 {
		msg += " " + formatFields(fields)
	}
	fmt.Printf("ERROR: %s\n", msg)
}

// log 内部日志方法
func (l *CustomLogger) log(level string, ctx context.Context, module, message string, fields ...zllog.Field) {
	msg := fmt.Sprintf("[%s] %s: %s", module, level, message)
	if len(fields) > 0 {
		msg += " " + formatFields(fields)
	}
	fmt.Printf("%s: %s\n", level, msg)
}

// formatFields 格式化字段
func formatFields(fields []zllog.Field) string {
	var result string
	for _, field := range fields {
		result += fmt.Sprintf("%s=%v ", field.Key, field.Value)
	}
	return result
}

// ============================================================================
// 使用示例
// ============================================================================

func main() {
	// 方式1：使用默认的 Zerolog Logger
	if err := zllog.InitLogger(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	zllog.Info(ctx, "main", "Using default Zerolog logger",
		zllog.String("service", "zllog"))

	fmt.Println("--- 切换到自定义 Logger ---")

	// 方式2：使用自定义 Logger
	customLogger := &CustomLogger{
		serviceName: "my-custom-service",
	}
	zllog.SetLogger(customLogger)

	// 现在所有日志调用都会使用自定义 Logger
	zllog.Info(ctx, "main", "Using custom logger",
		zllog.String("service", "zllog"))

	zllog.Debug(ctx, "database", "Query executed",
		zllog.String("sql", "SELECT * FROM users"),
		zllog.Int("rows", 10))

	zllog.Warn(ctx, "api", "Rate limit approaching",
		zllog.Int("requests", 990),
		zllog.Int("limit", 1000))

	// 测试带请求追踪的日志
	zllog.InfoWithRequest(ctx, "api", "Request processed", "req-123", 150,
		zllog.String("path", "/api/users"),
		zllog.String("method", "GET"))

	// 测试错误日志
	zllog.Error(ctx, "database", "Connection failed",
		fmt.Errorf("connection timeout"),
		zllog.String("host", "localhost:3306"),
		zllog.Int("port", 3306))

	// 测试带错误码的日志
	zllog.ErrorWithCode(ctx, "api", "Authentication failed", "AUTH_001",
		fmt.Errorf("invalid token"),
		zllog.String("user_id", "12345"))
}
