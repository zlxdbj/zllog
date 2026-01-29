# zllog - 中林日志组件

基于 **Zerolog** 的结构化日志组件，支持分布式追踪、自动日志轮转等功能。

## 特性

- ✅ **完全独立**：无项目特定依赖，可复制到任何项目使用
- ✅ **结构化日志**：基于 Zerolog，支持 JSON 格式输出
- 🔗 **分布式追踪解耦**：通过 `TraceIDProvider` 接口支持任意追踪系统
- 📁 **智能配置**：自动检测服务名、环境，支持多种配置来源
- 🔄 **日志轮转**：支持按大小和日期自动轮转、压缩（等保3合规）
- 🎨 **多格式输出**：控制台彩色文本 + 文件 JSON 格式
- 🚀 **高性能**：零内存分配，适合高并发场景
- 🌍 **环境感知**：根据环境自动调整配置

## 安装

```bash
go get github.com/zlxdbj/zllog
```

---

## 快速开始

### 1. 初始化日志系统

```go
import "github.com/zlxdbj/zllog"

func main() {
    // 自动从 resource/ 目录查找配置文件
    if err := zllog.InitLogger(); err != nil {
        panic(err)
    }

    // 使用日志
    zllog.Info(context.Background(), "main", "Application started")
}
```

### 2. 配置文件优先级

`InitLogger()` 按以下顺序查找配置：

1. `resource/log.yaml` - 独立日志配置
2. `resource/application.yaml` - 项目配置中的 `logger` 节点
3. `resource/application_{ENV}.yaml` - 环境特定配置
4. 默认配置

### 3. 环境变量

支持以下环境变量：

| 环境变量 | 说明 |
|---------|------|
| `SERVICE_NAME` 或 `APP_NAME` | 服务名称 |
| `ENV`、`APP_ENV`、`GO_ENV`、`MODE` | 环境名称（dev/test/prod） |

---

## 配置说明

### 方式1：独立配置文件 `log.yaml`

```yaml
# resource/log.yaml
service_name: my_service
env: dev
level: INFO
dir: ./logs
max_size: 100        # 单个文件最大大小（MB）
max_backups: 180     # 保留历史文件个数
max_age: 180         # 保留天数（等保3要求）
compress: true       # 压缩历史日志
daily_roll: true     # 按日期滚动
enable_console: true # 控制台输出
console_json: false  # false=彩色文本，true=JSON
```

### 方式2：项目配置文件 `application.yaml`

```yaml
# resource/application.yaml
logger:
  level: INFO
  dir: ./logs
  max_size: 100
  max_backups: 180
  max_age: 180
  compress: true
  daily_roll: true
  enable_console: true
  console_json: false
```

### 方式3：使用配置对象

```go
config := &zllog.LogConfig{
    ServiceName: "my_service",
    Env:         "dev",
    LogLevel:    "INFO",
    LogDir:      "./logs",
    MaxSize:     100,
    MaxBackups:  180,
    MaxAge:      180,
    Compress:    true,
    EnableDailyRoll: true,
    EnableConsole:   true,
}
zllog.InitLoggerWithConfig(config)
```

### 方式4：从指定文件加载

```go
// 从指定文件加载
zllog.InitLoggerFromFile("./config/log.yaml")

// 从指定目录查找
zllog.InitLoggerWithConfigDir("./resource")
```

---

## 使用示例

### 基础日志

```go
// Debug 日志
zllog.Debug(ctx, "module", "debug message")

// Info 日志
zllog.Info(ctx, "module", "info message")

// Warn 日志
zllog.Warn(ctx, "module", "warning message")

// Error 日志
zllog.Error(ctx, "module", "error message", err)

// Fatal 日志（会退出程序）
zllog.Fatal(ctx, "module", "fatal error", err)
```

### 带结构化字段

```go
zllog.Info(ctx, "database",
    "User login successful",
    zllog.String("user_id", "12345"),
    zllog.String("username", "john"),
    zllog.Int("age", 30),
    zllog.Bool("verified", true),
    zllog.Float("score", 99.5),
)
```

### 带请求追踪

```go
// 带请求ID和耗时的日志
zllog.InfoWithRequest(ctx, "api",
    "Request processed",
    requestID,      // 请求ID
    costMs,         // 耗时（毫秒）
    zllog.String("path", "/api/users"),
)

zllog.ErrorWithRequest(ctx, "api",
    "Request failed",
    requestID,
    err,
    costMs,
)
```

### 与追踪系统集成

```go
import (
    "github.com/zlxdbj/zllog"
    "github.com/zlxdbj/zltrace"
)

func main() {
    // 1. 初始化日志
    zllog.InitLogger()

    // 2. 初始化追踪
    zltrace.InitTracer()

    // 3. 日志会自动获取 trace_id
    ctx := context.Background()
    zllog.Info(ctx, "main", "Application started")
    // 输出：{"trace_id": "abc123...", "module": "main", ...}
}
```

---

## API 参考

### 初始化函数

| 函数 | 说明 |
|------|------|
| `InitLogger()` | 自动查找配置并初始化 |
| `InitLoggerWithConfig(*LogConfig)` | 使用配置对象初始化 |
| `InitLoggerWithConfigDir(string)` | 从指定目录查找配置 |
| `InitLoggerFromFile(string)` | 从指定文件加载配置 |

### 日志函数

| 函数 | 说明 |
|------|------|
| `Debug(ctx, module, message, fields...)` | DEBUG 级别日志 |
| `Info(ctx, module, message, fields...)` | INFO 级别日志 |
| `Warn(ctx, module, message, fields...)` | WARN 级别日志 |
| `Error(ctx, module, message, err, fields...)` | ERROR 级别日志 |
| `Fatal(ctx, module, message, err, fields...)` | FATAL 级别日志（会退出） |
| `InfoWithRequest(ctx, module, message, requestID, costMs, fields...)` | 带请求追踪的 INFO |
| `ErrorWithRequest(ctx, module, message, requestID, err, costMs, fields...)` | 带请求追踪的 ERROR |

### 字段函数

| 函数 | 说明 |
|------|------|
| `String(key, value)` | 字符串字段 |
| `Int(key, value)` | 整数字段 |
| `Int64(key, value)` | int64 字段 |
| `Float(key, value)` | 浮点数字段 |
| `Bool(key, value)` | 布尔字段 |
| `Any(key, value)` | 任意类型字段 |

### 配置结构

```go
type LogConfig struct {
    ServiceName      string  // 服务名称
    Env              string  // 环境（dev/test/prod）
    LogLevel         string  // 日志级别（DEBUG/INFO/WARN/ERROR/FATAL）
    LogDir           string  // 日志目录
    MaxSize          int     // 单文件最大大小（MB）
    MaxBackups       int     // 保留历史文件数
    MaxAge           int     // 保留天数
    Compress         bool    // 是否压缩
    EnableDailyRoll  bool    // 是否按日期滚动
    EnableConsole    bool    // 是否输出到控制台
    ConsoleJSONFormat bool   // 控制台是否 JSON 格式
}
```

---

## 配置自动调整

### 环境检测

按以下优先级自动检测环境：
1. 环境变量 `ENV`
2. 环境变量 `APP_ENV`
3. 环境变量 `GO_ENV`
4. 环境变量 `MODE`
5. 默认 `dev`

### 配置自动调整

| 环境 | 日志级别 | 控制台输出 |
|------|----------|------------|
| `prod` | INFO | 关闭 |
| `test` | INFO | 开启 |
| `dev` | DEBUG | 开启 |

### 服务名检测

按以下优先级自动检测服务名：
1. 环境变量 `SERVICE_NAME`
2. 环境变量 `APP_NAME`
3. 可执行文件名
4. 当前目录名
5. 默认 `service`

---

## 日志输出示例

### 开发环境（彩色文本）

```
[INFO]  2025-01-28 10:30:45  service=my_service env=dev host=localhost  trace_id=abc123  module=main  message=Application started
```

### 生产环境（JSON）

```json
{
  "level": "info",
  "timestamp": "2025-01-28T10:30:45.123456789Z",
  "service": "my_service",
  "env": "prod",
  "host": "server-01",
  "trace_id": "abc123def456",
  "module": "main",
  "message": "Application started"
}
```

---

## 常见问题（FAQ）

### Q1: 为什么需要传递 context.Context？

**常见疑问**：为什么每个日志函数都要传 `context.Context`？直接用 `context.Background()` 不是更简单吗？

#### Go 语言的标准做法

在 Go 语言中，**context 是请求范围的元数据传递的标准方式**：

- `database/sql` 包：所有查询方法都接收 context
- `net/http` 包：Request 包含 context
- Go 官方推荐：所有接收请求的函数都应接收 context

#### trace_id 的传递

```go
// ❌ 错误：trace_id 链中断
func ProcessData(data string) {
    zllog.Info(context.Background(), "module", "处理数据")
    // 每次调用都是新的 trace_id，无法追踪！
}

// ✅ 正确：trace_id 贯穿调用链
func ProcessData(ctx context.Context, data string) {
    zllog.Info(ctx, "module", "处理数据")
    // trace_id 从上游传递过来，可以追踪完整流程
}
```

#### 生产环境影响

| 方案 | 代码简洁性 | 可追踪性 | 生产环境适用性 |
|------|------------|----------|--------------|
| 所有函数传递 context | 较复杂 | ⭐⭐⭐⭐⭐ | ✅ 推荐 |
| 使用 `context.Background()` | 简单 | ⭐ | ❌ 不推荐 |

**结论**：虽然传递 context 让代码稍微复杂一点，但这是 **Go 语言的规约**，也是 **分布式系统的标准做法**。生产环境的可观测性比开发便利性更重要。

### Q2: 如何在项目中使用 zllog？

**方式1：作为项目子模块**（当前项目）
```go
import "github.com/zlxdbj/zllog"
zllog.InitLogger()
```

**方式2：复制到其他项目**
```bash
# 复制整个 zllog 目录到其他项目
cp -r zllog /path/to/other/project/zllog

# 在其他项目中使用
import "otherproject/zllog"
zllog.InitLogger()
```

**方式3：发布到 GitHub**（未来计划）
```bash
go get github.com/yourorg/zllog
```

### Q3: 如何切换环境？

**方式1：环境变量**
```bash
export ENV=prod
./go_shield
```

**方式2：配置文件**
```bash
# 使用指定的配置文件
./go_shield --config resource/application_prod.yaml
```

**方式3：启动参数**
```bash
MODE=prod ./go_shield
```

### Q4: 日志文件如何管理？

- **日志轮转**：自动按大小和日期切割
- **压缩**：历史日志自动压缩（gzip）
- **清理**：超过 `max_age` 天的日志自动删除
- **等保3合规**：默认保留 180 天

日志文件示例：
```
logs/
  ├── app.log                    # 当前日志
  ├── app-2025-01-27.log.gz     # 昨天的日志（已压缩）
  ├── app-2025-01-26.log.gz     # 前天的日志
  └── ...
```

### Q5: 如何集成到 GORM？

```go
import (
    "gorm.io/gorm"
    "github.com/zlxdbj/zllog/adapter/gormadapter"
)

db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: gormadapter.NewGormLogger(),
})
```

详见 `adapter/gormadapter/gorm.go`

### Q6: 如何实现自定义 Logger？

zllog 支持通过接口自定义日志实现，适用于以下场景：
- 将日志发送到远程日志服务（如 ELK、Loki）
- 使用其他日志库（如 logrus、zap）
- 实现特殊的日志格式或存储方式

**步骤1：实现 Logger 接口**

```go
import "github.com/zlxdbj/zllog"

type CustomLogger struct{}

func (l *CustomLogger) Debug(ctx context.Context, module, message string, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) Info(ctx context.Context, module, message string, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) Warn(ctx context.Context, module, message string, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) Error(ctx context.Context, module, message string, err error, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) ErrorWithCode(ctx context.Context, module, message, errorCode string, err error, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) Fatal(ctx context.Context, module, message string, err error, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) InfoWithRequest(ctx context.Context, module, message, requestID string, costMs int64, fields ...zllog.Field) {
    // 自定义实现
}

func (l *CustomLogger) ErrorWithRequest(ctx context.Context, module, message, requestID string, err error, costMs int64, fields ...zllog.Field) {
    // 自定义实现
}
```

**步骤2：注册自定义 Logger**

```go
// 注册自定义 Logger（会替换默认的 Zerolog 实现）
zllog.SetLogger(&CustomLogger{})

// 所有日志调用都会使用自定义实现
zllog.Info(ctx, "module", "message")
```

**完整示例**：参见 `examples/custom_logger_example.go`

---

## 最佳实践

### 1. 使用 module 字段分类

```go
// ✅ 好：清晰的模块划分
zllog.Error(ctx, "database", "Query failed", err)
zllog.Error(ctx, "api", "Handler failed", err)
zllog.Error(ctx, "kafka", "Consume failed", err)

// ❌ 不好：所有日志都用同一个 module
zllog.Error(ctx, "app", "Something went wrong", err)
```

### 2. 结构化字段优于字符串拼接

```go
// ❌ 不好：字符串拼接
zllog.Info(ctx, "api", "User "+userID+" logged in from "+ip)

// ✅ 好：结构化字段
zllog.Info(ctx, "api", "User logged in",
    zllog.String("user_id", userID),
    zllog.String("ip", ip),
)
```

### 3. Context 传递规范

```go
// ✅ HTTP Handler
func Handler(c *gin.Context) {
    ctx := c.Request.Context()
    zllog.Info(ctx, "api", "Processing request")
}

// ✅ 业务函数
func ProcessOrder(ctx context.Context, orderID string) error {
    zllog.Info(ctx, "order", "Processing order",
        zllog.String("order_id", orderID))
    // ...
}

// ❌ 避免：业务函数不接收 context
func ProcessOrder(orderID string) error {
    zllog.Info(context.Background(), "order", "Processing")
    // trace_id 链中断！
}
```

### 4. 生产环境配置建议

```yaml
env: prod
level: INFO
enable_console: false      # 关闭控制台输出
console_json: true         # 使用 JSON 格式
compress: true             # 启用压缩
max_age: 180               # 保留180天（等保3要求）
max_backups: 180           # 保留180个历史文件
```

### 5. 错误日志规范

```go
// ✅ 记录完整的错误上下文
zllog.Error(ctx, "database",
    "Failed to query user",
    err,
    zllog.String("query", sql),
    zllog.Int("attempt", retryCount),
    zllog.String("user_id", userID),
)

// ❌ 不好：缺少上下文
zllog.Error(ctx, "database", "Query failed", err)
```

---

## 与 zltrace 集成

zllog 通过 `TraceIDProvider` 接口与 zltrace 解耦：

```go
// 1. zltrace 实现 TraceIDProvider 接口
type OTELProvider struct {
    tracer *OTELTracer
    name   string
}

func (p *OTELProvider) GetTraceID(ctx context.Context) string {
    span := SpanFromContext(ctx)
    if span == nil {
        return ""
    }
    return span.TraceID()
}

// 2. 注册到 zllog
zllog.RegisterTraceIDProvider(&OTELProvider{...})

// 3. zllog 自动获取 trace_id
zllog.Info(ctx, "module", "message")
// 输出包含 trace_id: "abc123..."
```

**优势**：
- ✅ 完全解耦：zllog 不依赖 zltrace
- ✅ 灵活切换：可以使用不同的追踪系统
- ✅ 自动集成：无需手动传递 trace_id

---

## 性能说明

### 高性能设计

- **零内存分配**：基于 Zerolog，避免频繁的内存分配
- **结构化日志**：JSON 格式，便于日志分析系统处理
- **异步写入**：通过 lumberjack 实现日志文件的异步写入
- **批量刷新**：支持批量刷新到磁盘

### 性能对比

| 日志库 | 内存分配 | 性能 |
|--------|---------|------|
| zllog (Zerolog) | 零分配 | ⭐⭐⭐⭐⭐ |
| logrus | 有分配 | ⭐⭐⭐ |
| zap | 零分配 | ⭐⭐⭐⭐⭐ |
| 标准库 log | 有分配 | ⭐⭐ |

---

## 依赖说明

zllog 只依赖标准的第三方库：

```go
require (
    github.com/rs/zerolog v1.31.0         // 日志核心
    github.com/spf13/viper v1.18.2         // 配置加载
    gopkg.in/natefinch/lumberjack.v2 v2.2.1 // 日志轮转
    github.com/google/uuid v1.6.0           // UUID 生成
)
```

**无项目特定依赖**，可以复制到任何项目使用。

---

## 参考文档

- [Go Context 官方文档](https://golang.org/pkg/context/)
- [Go 数据库最佳实践](https://go.dev/doc/database/context)
- [分布式追踪标准](https://opentelemetry.io/docs/reference/specification/)
- [Zerolog 官方文档](https://github.com/rs/zerolog)
- [Context 传递规范](../Context传递规范.md)

---

## 更新日志

### v1.1.0 (2025-01-29)
- ✨ **新增 Logger 接口**：支持自定义日志实现
- ✨ 提供基于 Zerolog 的默认实现（ZerologLogger）
- ✨ 新增 `SetLogger()` 和 `GetLogger()` 方法
- ✨ 添加自定义 Logger 示例代码
- 🔄 重构：将 adapter/gorm.go 移至 adapter/gormadapter/gorm.go
- 📝 完善文档，添加接口使用指南

### v1.0.0 (2025-01-28)
- ✅ 完全独立，移除对项目特定代码的依赖
- ✅ 支持多种配置加载方式
- ✅ 支持独立的 log.yaml 配置
- ✅ 支持从 application.yaml 加载配置
- ✅ 支持直接传入配置对象
- ✅ 支持默认配置（无配置文件时）
- ✅ 环境感知，自动调整配置

---

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！
