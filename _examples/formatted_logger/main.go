package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zlxdbj/zllog"
)

func main() {
	// 初始化日志系统
	config := &zllog.LogConfig{
		ServiceName:      "formatted_example",
		Env:             "dev",
		LogLevel:        "DEBUG",
		LogDir:          "./logs",
		MaxSize:         100,
		MaxBackups:      10,
		MaxAge:          30,
		Compress:        true,
		EnableDailyRoll: true,
		EnableConsole:   true,
		ConsoleJSONFormat: false, // 控制台使用彩色文本
		EnableCaller:    true,    // 启用调用位置信息
	}

	if err := zllog.InitLoggerWithConfig(config); err != nil {
		panic(err)
	}

	ctx := context.Background()

	fmt.Println("\n=== 格式化日志示例 ===\n")

	// ========================================================================
	// 1. 基础格式化日志 - 适合简单的业务数据打印
	// ========================================================================

	fmt.Println("1. 基础格式化日志")
	zllog.Infof(ctx, "api", "User %s logged in from %s", "john", "192.168.1.100")
	zllog.Debugf(ctx, "cache", "Cache hit rate: %.2f%%", 85.67)
	zllog.Warnf(ctx, "api", "Request timeout after %dms", 5000)

	// ========================================================================
	// 2. 带错误的格式化日志
	// ========================================================================

	fmt.Println("\n2. 带错误的格式化日志")
	err := fmt.Errorf("database connection failed")
	zllog.Errorf(ctx, "database", "Failed to execute query: %s", err, "SELECT * FROM users")
	zllog.ErrorWithCodef(ctx, "api", "Authentication failed for user %s", "AUTH_001", err, "john")

	// ========================================================================
	// 3. 带请求追踪的格式化日志
	// ========================================================================

	fmt.Println("\n3. 带请求追踪的格式化日志")
	requestID := "req-123456"
	startTime := time.Now()
	time.Sleep(100 * time.Millisecond)
	costMs := time.Since(startTime).Milliseconds()

	zllog.InfoWithRequestf(ctx, "api", "Processed %d items", requestID, costMs, 150)
	zllog.ErrorWithRequestf(ctx, "api", "Failed to process order %s", requestID, err, costMs, "ORD-12345")

	// ========================================================================
	// 4. 实际业务场景
	// ========================================================================

	fmt.Println("\n4. 实际业务场景")

	// 场景1: 用户注册
	zllog.Infof(ctx, "user", "New user registered: %s <%s>", "user-12345", "john@example.com")

	// 场景2: 订单处理
	zllog.Infof(ctx, "order", "Order %s created with total $%.2f", "ORD-67890", 299.99)

	// 场景3: 性能监控
	zllog.Debugf(ctx, "database", "Query executed in %dms: %s", 45, "SELECT * FROM users")

	// 场景4: API 响应
	zllog.Infof(ctx, "api", "API response: %s %s - %d %s", "GET", "/api/users/12345", 200, "OK")

	// ========================================================================
	// 5. 最佳实践 - 如何选择使用格式化还是结构化字段
	// ========================================================================

	fmt.Println("\n=== 最佳实践对比 ===\n")

	// ❌ 不好：字符串拼接（无法搜索）
	fmt.Println("❌ 不好：字符串拼接")
	zllog.Info(ctx, "api", "User john (ID: 12345) logged in from 192.168.1.100")

	// ✅ 好：结构化字段（可搜索、可分析）
	fmt.Println("\n✅ 好：结构化字段（可搜索）")
	zllog.Info(ctx, "api", "User logged in",
		zllog.String("username", "john"),
		zllog.String("user_id", "12345"),
		zllog.String("ip", "192.168.1.100"),
		zllog.Bool("verified", true),
	)

	// ✅ 好：格式化日志（直观、易读）
	fmt.Println("\n✅ 好：格式化日志（直观）")
	zllog.Infof(ctx, "api", "User john (ID: 12345) logged in from 192.168.1.100")

	// ========================================================================
	// 使用建议
	// ========================================================================

	fmt.Println("\n=== 使用建议 ===\n")
	fmt.Println("1. 临时调试、简单业务数据 → 使用格式化日志 (Infof)")
	fmt.Println("2. 生产环境、需要分析 → 使用结构化字段 (Info + Field)")
	fmt.Println("3. 关键错误 → 必须使用结构化字段 (Error + Field)")
	fmt.Println("")

	// 示例：临时调试
	zllog.Debugf(ctx, "debug", "Processing item %d/%d", 5, 100)

	// 示例：生产环境（可搜索）
	zllog.Info(ctx, "business", "Order processed",
		zllog.String("order_id", "ORD-12345"),
		zllog.String("status", "completed"),
		zllog.Float64("total", 299.99),
	)

	// 示例：关键错误（必须可搜索）
	zllog.Error(ctx, "database", "Connection failed", err,
		zllog.String("host", "localhost"),
		zllog.Int("port", 5432),
		zllog.String("database", "mydb"),
		zllog.Int("retry_count", 3),
	)

	fmt.Println("\n=== 示例完成 ===")
}
