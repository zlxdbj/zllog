package zllog

import "time"

// ============================================================================
// 结构化日志字段（完整支持 Zerolog 所有类型）
// ============================================================================

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// ============================================================================
// 基础类型
// ============================================================================

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// ============================================================================
// 整数类型
// ============================================================================

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int8 创建 int8 字段
func Int8(key string, value int8) Field {
	return Field{Key: key, Value: value}
}

// Int16 创建 int16 字段
func Int16(key string, value int16) Field {
	return Field{Key: key, Value: value}
}

// Int32 创建 int32 字段
func Int32(key string, value int32) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建 int64 字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// ============================================================================
// 无符号整数类型
// ============================================================================

// Uint 创建 uint 字段
func Uint(key string, value uint) Field {
	return Field{Key: key, Value: value}
}

// Uint8 创建 uint8 字段
func Uint8(key string, value uint8) Field {
	return Field{Key: key, Value: value}
}

// Uint16 创建 uint16 字段
func Uint16(key string, value uint16) Field {
	return Field{Key: key, Value: value}
}

// Uint32 创建 uint32 字段
func Uint32(key string, value uint32) Field {
	return Field{Key: key, Value: value}
}

// Uint64 创建 uint64 字段
func Uint64(key string, value uint64) Field {
	return Field{Key: key, Value: value}
}

// ============================================================================
// 浮点类型
// ============================================================================

// Float32 创建 float32 字段
func Float32(key string, value float32) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建 float64 字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// ============================================================================
// 时间类型
// ============================================================================

// Time 创建时间字段
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Dur 创建时间间隔字段（duration）
func Dur(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// ============================================================================
// 错误与接口
// ============================================================================

// Err 创建错误字段
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// NamedErr 创建命名错误字段
func NamedErr(key string, err error) Field {
	return Field{Key: key, Value: err}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// ============================================================================
// 高级类型
// ============================================================================

// Interface 创建接口类型字段（与 Any 相同）
func Interface(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// RawJSON 创建原始 JSON 字段（不会再次序列化）
func RawJSON(key string, b []byte) Field {
	return Field{Key: key, Value: b}
}

// Dict 创建字典字段（用于嵌套对象）
func Dict(key string, f ...Field) Field {
	return Field{Key: key, Value: f}
}

// Array 创建数组字段
func Array(key string, f ...Field) Field {
	return Field{Key: key, Value: f}
}
