package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/zlxdbj/zllog"
)

// ============================================================================
// 远程日志 Logger 示例 - 将日志发送到 HTTP 服务
// ============================================================================

// RemoteLoggerConfig 远程日志配置
type RemoteLoggerConfig struct {
	Endpoint   string        // 日志接收端点
	BatchSize  int           // 批量发送大小
	Timeout    time.Duration // 请求超时
	MaxRetries int           // 最大重试次数
}

// RemoteLogEntry 远程日志条目
type RemoteLogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Module    string                 `json:"module"`
	Message   string                 `json:"message"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	CostMs    int64                  `json:"cost_ms,omitempty"`
	ErrorCode string                 `json:"error_code,omitempty"`
}

// RemoteLogger 将日志发送到远程 HTTP 服务的 Logger 实现
type RemoteLogger struct {
	config  RemoteLoggerConfig
	client  *http.Client
	buffer  []RemoteLogEntry
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	service string
	env     string
}

// NewRemoteLogger 创建远程日志实例
func NewRemoteLogger(config RemoteLoggerConfig) *RemoteLogger {
	ctx, cancel := context.WithCancel(context.Background())

	return &RemoteLogger{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		buffer: make([]RemoteLogEntry, 0, config.BatchSize),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动后台发送协程
func (l *RemoteLogger) Start() {
	l.wg.Add(1)
	go l.flushLoop()
}

// Stop 停止并发送剩余日志
func (l *RemoteLogger) Stop() {
	l.cancel()
	l.wg.Wait()
	l.flush() // 发送剩余日志
}

// flushLoop 定期刷新日志
func (l *RemoteLogger) flushLoop() {
	defer l.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.flush()
		}
	}
}

// flush 发送缓冲区中的日志
func (l *RemoteLogger) flush() {
	l.mu.Lock()
	if len(l.buffer) == 0 {
		l.mu.Unlock()
		return
	}

	// 复制缓冲区并清空
	entries := make([]RemoteLogEntry, len(l.buffer))
	copy(entries, l.buffer)
	l.buffer = l.buffer[:0]
	l.mu.Unlock()

	// 发送日志
	if err := l.send(entries); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send logs: %v\n", err)
		// 失败重试
		for i := 0; i < l.config.MaxRetries; i++ {
			time.Sleep(time.Second * time.Duration(i+1))
			if err := l.send(entries); err == nil {
				break
			}
		}
	}
}

// send 发送日志到远程服务
func (l *RemoteLogger) send(entries []RemoteLogEntry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(l.ctx, "POST", l.config.Endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "zllog-remote-logger/1.0")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// add 添加日志到缓冲区
func (l *RemoteLogger) add(ctx context.Context, entry RemoteLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.Timestamp = time.Now().Format(time.RFC3339Nano)
	entry.TraceID = zllog.GetOrCreateTraceID(ctx)
	l.buffer = append(l.buffer, entry)

	// 如果达到批量大小，立即发送
	if len(l.buffer) >= l.config.BatchSize {
		// 在新协程中发送，避免阻塞
		go l.flush()
	}
}

// Debug 实现 Logger 接口的 Debug 方法
func (l *RemoteLogger) Debug(ctx context.Context, module, message string, fields ...zllog.Field) {
	l.add(ctx, RemoteLogEntry{
		Level:   "DEBUG",
		Module:  module,
		Message: message,
		Fields:  fieldsToMap(fields),
	})
}

// Info 实现 Logger 接口的 Info 方法
func (l *RemoteLogger) Info(ctx context.Context, module, message string, fields ...zllog.Field) {
	l.add(ctx, RemoteLogEntry{
		Level:   "INFO",
		Module:  module,
		Message: message,
		Fields:  fieldsToMap(fields),
	})
}

// Warn 实现 Logger 接口的 Warn 方法
func (l *RemoteLogger) Warn(ctx context.Context, module, message string, fields ...zllog.Field) {
	l.add(ctx, RemoteLogEntry{
		Level:   "WARN",
		Module:  module,
		Message: message,
		Fields:  fieldsToMap(fields),
	})
}

// Error 实现 Logger 接口的 Error 方法
func (l *RemoteLogger) Error(ctx context.Context, module, message string, err error, fields ...zllog.Field) {
	entry := RemoteLogEntry{
		Level:   "ERROR",
		Module:  module,
		Message: message,
		Fields:  fieldsToMap(fields),
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.add(ctx, entry)
}

// ErrorWithCode 实现 Logger 接口的 ErrorWithCode 方法
func (l *RemoteLogger) ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...zllog.Field) {
	entry := RemoteLogEntry{
		Level:     "ERROR",
		Module:    module,
		Message:   message,
		ErrorCode: errorCode,
		Fields:    fieldsToMap(fields),
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.add(ctx, entry)
}

// Fatal 实现 Logger 接口的 Fatal 方法
func (l *RemoteLogger) Fatal(ctx context.Context, module, message string, err error, fields ...zllog.Field) {
	entry := RemoteLogEntry{
		Level:   "FATAL",
		Module:  module,
		Message: message,
		Fields:  fieldsToMap(fields),
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.add(ctx, entry)
	l.flush() // 立即发送
	os.Exit(1)
}

// InfoWithRequest 实现 Logger 接口的 InfoWithRequest 方法
func (l *RemoteLogger) InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...zllog.Field) {
	l.add(ctx, RemoteLogEntry{
		Level:     "INFO",
		Module:    module,
		Message:   message,
		RequestID: requestID,
		CostMs:    costMs,
		Fields:    fieldsToMap(fields),
	})
}

// ErrorWithRequest 实现 Logger 接口的 ErrorWithRequest 方法
func (l *RemoteLogger) ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...zllog.Field) {
	entry := RemoteLogEntry{
		Level:     "ERROR",
		Module:    module,
		Message:   message,
		RequestID: requestID,
		CostMs:    costMs,
		Fields:    fieldsToMap(fields),
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.add(ctx, entry)
}

// fieldsToMap 将字段数组转换为 map
func fieldsToMap(fields []zllog.Field) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}

	m := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		m[field.Key] = field.Value
	}
	return m
}

// ============================================================================
// 使用示例
// ============================================================================

func main() {
	// 创建远程日志实例
	remoteLogger := NewRemoteLogger(RemoteLoggerConfig{
		Endpoint:   "http://localhost:8080/api/logs", // 替换为实际的日志接收端点
		BatchSize:  10,
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	})

	// 启动后台发送协程
	remoteLogger.Start()
	defer remoteLogger.Stop()

	// 注册到 zllog
	zllog.SetLogger(remoteLogger)

	// 现在所有日志都会发送到远程服务
	ctx := context.Background()

	// 测试各种日志
	zllog.Debug(ctx, "app", "Application starting",
		zllog.String("version", "1.0.0"))

	zllog.Info(ctx, "database", "Connected to database",
		zllog.String("host", "localhost"),
		zllog.String("database", "mydb"))

	zllog.Warn(ctx, "cache", "Cache miss rate increasing",
		zllog.Float64("rate", 0.35))

	// 模拟一些日志，触发批量发送
	for i := 0; i < 15; i++ {
		zllog.Info(ctx, "test", fmt.Sprintf("Test message %d", i),
			zllog.Int("index", i))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Logs sent to remote server")

	// 等待所有日志发送完成
	time.Sleep(2 * time.Second)
}
