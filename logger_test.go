package zllog

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
)

// MockLogger 用于测试的 Mock 实现
type MockLogger struct {
	mu        sync.Mutex
	calls     []string
	lastLevel string
}

func (m *MockLogger) Debug(ctx context.Context, module, message string, fields ...Field) {
	m.record("DEBUG", module, message)
}

func (m *MockLogger) Info(ctx context.Context, module, message string, fields ...Field) {
	m.record("INFO", module, message)
}

func (m *MockLogger) Warn(ctx context.Context, module, message string, fields ...Field) {
	m.record("WARN", module, message)
}

func (m *MockLogger) Error(ctx context.Context, module, message string, err error, fields ...Field) {
	m.record("ERROR", module, message)
}

func (m *MockLogger) ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...Field) {
	m.record("ERROR_CODE", module, message)
}

func (m *MockLogger) Fatal(ctx context.Context, module, message string, err error, fields ...Field) {
	m.record("FATAL", module, message)
	os.Exit(0) // 测试中不真正退出
}

func (m *MockLogger) InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...Field) {
	m.record("INFO_REQUEST", module, message)
}

func (m *MockLogger) ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...Field) {
	m.record("ERROR_REQUEST", module, message)
}

func (m *MockLogger) record(level, module, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, fmt.Sprintf("[%s] %s: %s", level, module, message))
	m.lastLevel = level
}

func (m *MockLogger) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *MockLogger) getLastCall() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.calls) == 0 {
		return ""
	}
	return m.calls[len(m.calls)-1]
}

// TestLoggerInterface 测试 Logger 接口
func TestLoggerInterface(t *testing.T) {
	// 保存原始的 logger
	originalLogger := globalLoggerImpl
	defer func() {
		globalLoggerImpl = originalLogger
	}()

	// 创建 mock logger
	mock := &MockLogger{}
	SetLogger(mock)

	ctx := context.Background()

	// 测试 Debug
	Debug(ctx, "test", "debug message")
	if mock.getLastCall() != "[DEBUG] test: debug message" {
		t.Errorf("Debug call failed: %s", mock.getLastCall())
	}

	// 测试 Info
	Info(ctx, "test", "info message")
	if mock.getLastCall() != "[INFO] test: info message" {
		t.Errorf("Info call failed: %s", mock.getLastCall())
	}

	// 测试 Warn
	Warn(ctx, "test", "warn message")
	if mock.getLastCall() != "[WARN] test: warn message" {
		t.Errorf("Warn call failed: %s", mock.getLastCall())
	}

	// 测试 Error
	Error(ctx, "test", "error message", fmt.Errorf("test error"))
	if mock.getLastCall() != "[ERROR] test: error message" {
		t.Errorf("Error call failed: %s", mock.getLastCall())
	}

	// 测试 ErrorWithCode
	ErrorWithCode(ctx, "test", "error with code", "E001", fmt.Errorf("test error"))
	if mock.getLastCall() != "[ERROR_CODE] test: error with code" {
		t.Errorf("ErrorWithCode call failed: %s", mock.getLastCall())
	}

	// 测试 InfoWithRequest
	InfoWithRequest(ctx, "test", "info with request", "req-123", 100)
	if mock.getLastCall() != "[INFO_REQUEST] test: info with request" {
		t.Errorf("InfoWithRequest call failed: %s", mock.getLastCall())
	}

	// 测试 ErrorWithRequest
	ErrorWithRequest(ctx, "test", "error with request", "req-123", fmt.Errorf("test error"), 100)
	if mock.getLastCall() != "[ERROR_REQUEST] test: error with request" {
		t.Errorf("ErrorWithRequest call failed: %s", mock.getLastCall())
	}

	// 验证调用次数
	expectedCalls := 7
	if mock.getCallCount() != expectedCalls {
		t.Errorf("Expected %d calls, got %d", expectedCalls, mock.getCallCount())
	}
}

// TestSetLogger 测试 SetLogger 和 GetLogger
func TestSetLogger(t *testing.T) {
	// 保存原始的 logger
	originalLogger := globalLoggerImpl
	defer func() {
		globalLoggerImpl = originalLogger
	}()

	mock := &MockLogger{}
	SetLogger(mock)

	// 验证 GetLogger 返回正确的实现
	retrieved := GetLogger()
	if retrieved != mock {
		t.Error("GetLogger did not return the set logger")
	}
}

// TestZerologLogger 测试默认的 ZerologLogger 实现
func TestZerologLogger(t *testing.T) {
	// 需要先初始化 logger
	if err := InitLoggerWithConfig(DefaultConfig("test")); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	ctx := context.Background()

	// 这些测试只验证不会 panic，实际输出到文件/控制台
	Info(ctx, "test", "info message")
	Debug(ctx, "test", "debug message")
	Warn(ctx, "test", "warn message")
}
