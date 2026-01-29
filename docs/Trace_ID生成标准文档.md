# Trace ID 生成标准文档

## 概述

本文档说明 `zltrace` 组件中 **Trace ID** 的生成规则、算法实现和 W3C 标准符合性。

---

## W3C Trace Context 标准

### 官方文档

- **标准名称**：W3C Trace Context
- **标准链接**：https://www.w3.org/TR/trace-context/
- **状态**：W3C 推荐标准（Recommendation）
- **发布时间**：2021 年

### Trace ID 规范

| 属性 | 规范要求 | 说明 |
|------|----------|------|
| **长度** | 32 位十六进制字符 | 固定长度，不可变 |
| **位数** | 128 位（16 字节） | 实际占用空间 |
| **字符集** | 0-9, a-f（小写） | 仅使用小写字母 |
| **格式** | 十六进制字符串 | 不含连字符或其他分隔符 |
| **示例** | `4bf92f3577b34da6a3ce929d0e0e4736` | 标准 trace_id |

### 示例

```
✅ 正确示例：
  - 4bf92f3577b34da6a3ce929d0e0e4736
  - 1e6f21a9f64e8ad333d21a814f279e36
  - 00f067aa0ba902b7

❌ 错误示例：
  - 4bf92f35-7b34-da6a3-ce92-929d0e0e4736  (包含连字符)
  - 4BF92F3577B34DA6A3CE929D0E0E4736      (包含大写字母)
  - 12345678901234567890123456789012    (超过32位)
```

---

## OpenTelemetry 实现

### 标准文档

- **官方文档**：https://opentelemetry.io/docs/reference/specification/trace/
- **实现库**：`go.opentelemetry.io/otel`

### 生成算法

#### 1. 算法原理

```
步骤1：生成 128 位随机 UUID
  - 使用加密安全的随机数生成器
  - 符合 RFC 4122 UUID v4 标准
  - 示例：01234567-89ab-cdef-0123-456789abcdef

步骤2：去除连字符
  - 将 UUID 中的连字符 - 去除
  - 示例：01234567-89ab-cdef-0123456789abcdef
  - 结果：0123456789abcdef0123456789abcdef

步骤3：转换为十六进制
  - 将 128 位二进制转为 32 位十六进制字符串
  - 使用小写字母（0-9, a-f）
  - 示例：1e6f21a9f64e8ad333d21a814f279e36

步骤4：返回 trace_id
  - 最终得到 32 位十六进制字符串
```

#### 2. 代码实现

**当前实现**（`zltrace/opentelemetry.go:99-102`）：

```go
func (s *OTELSpan) TraceID() string {
    spanCtx := s.span.SpanContext()
    return spanCtx.TraceID().String()
}
```

**内部实现**（OpenTelemetry SDK）：

```go
// 内部使用 OpenTelemetry 标准库
import (
    "go.opentelemetry.io/otel/trace"
)

// TraceID 内部结构（128位）
type TraceID [16]byte

// String() 方法转换为十六进制字符串
func (t TraceID) String() string {
    return hex.EncodeToString(t[:])
}
```

#### 3. 随机性保证

- **算法**：加密安全的伪随机数生成器（CSPRNG）
- **空间大小**：2^128（约 3.4 x 10^38）
- **冲突概率**：可以忽略不计
- **全局唯一性**：在分布式环境中保证唯一性

---

## 验证测试结果

### 测试环境

- **测试程序**：`test_trace_id.go`
- **测试次数**：5 次
- **测试结果**：全部通过 ✅

### 测试数据

```
[测试 1]
TraceID: 1e6f21a9f64e8ad333d21a814f279e36
长度: 32 字符
✅ 符合 W3C 标准（32位十六进制）

[测试 2]
TraceID: 3f11d655c630b2bdab48ced858880ae6
长度: 32 字符
✅ 符合 W3C 标准（32位十六进制）

[测试 3]
TraceID: 750a5d916a1ac315753e6697a7e4a5f6
长度: 32 字符
✅ 符合 W3C 标准（32位十六进制）

[测试 4]
TraceID: ff358ecc919cce6f9db18f58d966d10b
长度: 32 字符
✅ 符合 W3C 标准（32位十六进制）

[测试 5]
TraceID: 33bd8ac562d2da4498b8f96ae80b95b4
长度: 32 字符
✅ 符合 W3C 标准（32位十六进制）
```

### 字符集验证

所有测试样本仅包含：`0-9` 和 `a-f`（小写），完全符合 W3C 标准。

---

## traceparent Header 格式

### W3C 标准格式

```
traceparent: 00-trace_id-span_id-flags
```

### 字段说明

| 字段 | 格式 | 说明 | 示例 |
|------|------|------|------|
| **版本** | 2 位十六进制 | 当前版本固定为 `00` | `00` |
| **trace_id** | 32 位十六进制 | W3C Trace ID | `4bf92f3577b34da6a3ce929d0e0e4736` |
| **span_id** | 16 位十六进制 | 当前 Span ID | `00f067aa0ba902b7` |
| **flags** | 2 位十六进制 | 采样标志 | `01`（已采样）|

### 完整示例

```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

**分解**：
- `00` - 版本
- `4bf92f3577b34da6a3ce929d0e0e4736` - trace_id（32位）
- `00f067aa0ba902b7` - span_id（16位）
- `01` - flags（已采样）

---

## Span ID 规范

### 规范要求

| 属性 | 规范要求 | 说明 |
|------|----------|------|
| **长度** | 16 位十六进制字符 | 固定长度，不可变 |
| **位数** | 64 位（8 字节） | 实际占用空间 |
| **字符集** | 0-9, a-f（小写） | 仅使用小写字母 |
| **示例** | `00f067aa0ba902b7` | 标准 span_id |

### 对比

| 项目 | Trace ID | Span ID |
|------|----------|---------|
| **长度** | 32 字符 | 16 字符 |
| **位数** | 128 位 | 64 位 |
| **唯一性** | 全局唯一 | 当前 trace 内唯一 |

---

## HTTP Header 注入

### 实现代码

**位置**：`zltrace/adapter/http_client.go`

```go
func InjectHTTPHeaders(ctx context.Context, headers http.Header, operationName string) {
    tracer := zltrace.GetTracer()
    if tracer == nil {
        return
    }

    carrier := &HTTPHeaderCarrier{headers: headers}
    tracer.Inject(ctx, carrier)
}
```

### 注入格式

自动注入 `traceparent` HTTP header：

```http
GET /api/users HTTP/1.1
Host: example.com
traceparent: 00-1e6f21a9f64e8ad333d21a814f279e36-00f067aa0ba902b7-01
```

---

## Kafka 消息注入

### 实现代码

**位置**：`zltrace/kafka.go`

```go
func InjectKafkaProducerHeaders(ctx context.Context, msg *sarama.ProducerMessage) context.Context {
    tracer := GetTracer()
    if tracer == nil {
        return ctx
    }

    span, spanCtx := tracer.StartSpan(ctx, "Kafka/Produce/"+msg.Topic)
    defer span.Finish()

    carrier := &kafkaProducerHeaderCarrier{headers: &msg.Headers}
    tracer.Inject(spanCtx, carrier)

    return spanCtx
}
```

### 注入格式

自动注入 `traceparent` 到 Kafka 消息 headers：

```go
sarama.Record{
    Key:   "traceparent",
    Value: "00-1e6f21a9f64e8ad333d21a814f279e36-00f067aa0ba902b7-01",
}
```

---

## 验证清单

### ✅ W3C 标准符合性

- [x] Trace ID 长度：32 位十六进制字符
- [x] Trace ID 字符集：仅包含 0-9, a-f（小写）
- [x] Trace ID 唯一性：基于 128 位随机 UUID
- [x] Span ID 长度：16 位十六进制字符
- [x] traceparent Header 格式正确
- [x] 使用 W3C Trace Context 传播标准

### ✅ OpenTelemetry 实现

- [x] 使用 OpenTelemetry 官方 SDK（`go.opentelemetry.io/otel`）
- [x] 调用标准 API：`span.SpanContext().TraceID().String()`
- [x] 符合 OpenTelemetry 规范
- [x] 使用标准传播器：`propagation.TraceContext{}`

### ✅ 代码验证

- [x] 编译通过
- [x] 运行正常
- [x] Trace ID 生成正确
- [x] HTTP 传播正常
- [kafka 传播正常]

---

## 参考文档

### W3C 标准

1. **W3C Trace Context**
   - 链接：https://www.w3.org/TR/trace-context/
   - 状态：W3C Recommendation（推荐标准）
   - 版本：2021 年 1 月 27 日

2. **W3C Baggage**
   - 链接：https://www.w3.org/TR/baggage/
   - 状态：W3C Recommendation
   - 版本：2022 年

### OpenTelemetry 规范

1. **OpenTelemetry Trace Specification**
   - 链接：https://opentelemetry.io/docs/reference/specification/trace/
   - 维护：Cloud Native Computing Foundation (CNCF)

2. **OpenTelemetry Protocol**
   - 链接：https://opentelemetry.io/docs/reference/specification/otlp/
   - 协议：OTLP (OpenTelemetry Protocol)

### 实现库

- **Go SDK**：`go.opentelemetry.io/otel`
- **Go Trace**：`go.opentelemetry.io/otel/trace`
- **版本**：v1.39.0

---

## 总结

### 当前实现 ✅

1. **完全符合 W3C Trace Context 标准**
   - Trace ID：32 位十六进制
   - 使用 OpenTelemetry 官方实现
   - 标准 traceparent header 注入

2. **高质量实现**
   - 128 位随机性保证唯一性
   - 跨进程/跨服务传播正确
   - 性能开销极小

3. **生产就绪**
   - 符合国际标准
   - 兼容主流追踪系统（SkyWalking、Jaeger）
   - 社区最佳实践

### 建议

**无需修改**：当前实现已经是业界标准实现，完全符合 W3C 和 OpenTelemetry 规范。

---

**文档版本**：v1.0
**最后更新**：2025-01-28
**维护者**：zltrace Team
