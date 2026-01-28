package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/zlxdbj/zllog"

	"gorm.io/gorm/logger"
)

// ============================================================================
// GORM Logger 适配器 - 使用 zllog 输出 SQL 日志
// ============================================================================

// GormLogger GORM Logger 适配器，将 SQL 查询日志输出到 zllog
//
// 用法示例：
//   import "go_shield/zllog/adapter"
//
//   db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
//       Logger: adapter.NewGormLogger(),
//   })
//
// 特性：
//   - SQL 查询日志自动包含 trace_id（当在 HTTP 请求上下文中时）
//   - 错误日志使用 ERROR 级别，成功查询使用 INFO 级别
//   - 结构化字段：source=gorm, elapsed_ms, module=database
//   - 支持日志级别过滤
type GormLogger struct{}

// NewGormLogger 创建 GORM Logger 实例
func NewGormLogger() *GormLogger {
	return &GormLogger{}
}

// Format 格式化 SQL 日志
func (l *GormLogger) Format(level logger.LogLevel, msg string, tz *time.Location) string {
	// GORM 内部会传入完整的SQL日志信息
	return msg
}

// Print 输出日志到 zllog
func (l *GormLogger) Print(msg string) {
	// 创建一个后台context用于数据库日志（没有HTTP请求）
	ctx := context.Background()

	// 解析消息，提取关键信息
	// 格式示例：
	// [2.508ms] [rows:0] SELECT * FROM `t_ptg_shield_mq_fail` LIMIT 5
	// [2025/01/27 14:12:52.123] [rows:1] SELECT...

	// 直接记录为调试日志
	zllog.Debug(ctx, "database", msg,
		zllog.String("source", "gorm"))
}

// Error 输出错误日志到 zllog
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 将args格式化为字符串
	argsStr := fmt.Sprintf("%v", args)
	zllog.Error(ctx, "database", msg, fmt.Errorf("%s", argsStr),
		zllog.String("source", "gorm"))
}

// Info 输出信息日志到 zllog
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 将args作为字段添加
	zllog.Info(ctx, "database", msg,
		zllog.Any("args", fmt.Sprintf("%v", args)),
		zllog.String("source", "gorm"))
}

// Warn 输出警告日志到 zllog
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 将args作为字段添加
	zllog.Warn(ctx, "database", msg,
		zllog.Any("args", fmt.Sprintf("%v", args)),
		zllog.String("source", "gorm"))
}

// Log 根据level输出日志到 zllog
func (l *GormLogger) Log(level logger.LogLevel, msg string) {
	// GORM会将info级别以上的日志通过Print输出
	// 我们只关心慢查询，所以统一用Debug级别记录
	ctx := context.Background()
	zllog.Debug(ctx, "database", msg,
		zllog.String("source", "gorm"))
}

// LogMode 设置日志级别（GORM logger.Interface 要求）
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// 返回自身，支持链式调用
	return l
}

// Trace 追踪SQL执行（GORM logger.Interface 要求）
// 这个方法是核心，记录所有 SQL 查询的详细信息
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取SQL和执行时间
	sql, rows := fc()
	elapsed := time.Since(begin)

	// 构建日志消息
	msg := fmt.Sprintf("[%s] [rows:%d] %s", elapsed, rows, sql)

	if err != nil {
		// SQL执行错误，记录错误日志
		zllog.Error(ctx, "database", msg, err,
			zllog.String("source", "gorm"),
			zllog.Int64("elapsed_ms", elapsed.Milliseconds()))
	} else {
		// SQL执行成功，记录调试日志
		zllog.Debug(ctx, "database", msg,
			zllog.String("source", "gorm"),
			zllog.Int64("elapsed_ms", elapsed.Milliseconds()))
	}
}
