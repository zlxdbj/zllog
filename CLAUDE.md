# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

zllog 是基于 Zerolog 的结构化日志组件，支持分布式追踪、自动日志轮转等功能。核心设计理念是通过接口解耦，允许用户替换日志实现。

**核心特性**：
- 完全独立，无项目特定依赖
- 通过 `TraceIDProvider` 接口与追踪系统解耦
- 通过 `Logger` 接口支持自定义日志实现
- 自动检测配置文件（支持多种来源和环境变量）
- 按大小和日期自动轮转日志（等保3合规）

## 开发命令

```bash
# 构建所有包（_examples 目录会被自动忽略）
go build ./...

# 运行所有测试
go test ./...

# 运行特定测试
go test -run TestLoggerInterface

# 运行示例
cd _examples/custom_logger && go run main.go
cd _examples/remote_logger && go run main.go

# 更新依赖
go mod tidy
```

## 核心架构

### 文件组织原则

项目采用接口与实现分离的架构：

- **log.go** - 核心接口、配置和全局 API（不包含具体实现）
- **field.go** - 结构化日志字段类型定义（完整支持 Zerolog 所有类型）
- **zerolog_logger.go** - Logger 接口的 Zerolog 默认实现
- **config.go** - 配置加载器
- **adapter/** - 第三方库适配器

### 关键接口

#### 1. TraceIDProvider 接口

用于从 context 提取 trace_id，支持任意追踪系统（SkyWalking、Jaeger、OpenTelemetry）。

```go
type TraceIDProvider interface {
    GetTraceID(ctx context.Context) string
    Name() string
}
```

用户需要调用 `RegisterTraceIDProvider()` 注册实现。

#### 2. Logger 接口

定义了所有日志方法，用户可以实现此接口来使用自定义日志库。

```go
type Logger interface {
    Debug(ctx, module, message string, fields ...Field)
    Info(ctx, module, message string, fields ...Field)
    Warn(ctx, module, message string, fields ...Field)
    Error(ctx, module, message string, err error, fields ...Field)
    ErrorWithCode(ctx, module, message, errorCode string, err error, fields ...Field)
    Fatal(ctx, module, message string, err error, fields ...Field)
    InfoWithRequest(ctx, module, message, requestID string, costMs int64, fields ...Field)
    ErrorWithRequest(ctx, module, message, requestID string, err error, costMs int64, fields ...Field)
}
```

默认实现：`ZerologLogger`

### 字段类型系统

所有字段类型定义在 **field.go**，支持完整的 Zerolog 类型：

- **基础**: String, Bool
- **整数**: Int, Int8, Int16, Int32, Int64
- **无符号**: Uint, Uint8, Uint16, Uint32, Uint64
- **浮点**: Float32, Float64
- **时间**: Time, Dur
- **错误**: Err, NamedErr
- **高级**: RawJSON, Dict, Array, Any, Interface

**重要**: 字段通过 type switch 处理，所有 Logger 实现必须正确处理这些类型。参考 `zerolog_logger.go` 的 `addFields()` 方法。

### 配置加载优先级

`InitLogger()` 按以下顺序查找配置：

1. `resource/log.yaml` - 独立日志配置
2. `resource/application.yaml` - 项目配置中的 `logger` 节点
3. `resource/application_{ENV}.yaml` - 环境特定配置
4. 默认配置

支持的环境变量（用于自动检测服务名和环境）：
- `SERVICE_NAME` 或 `APP_NAME`
- `ENV`、`APP_ENV`、`GO_ENV`、`MODE`

### trace_id 生成机制

`GetOrCreateTraceID(ctx)` 函数是公开的 API，工作流程：

1. 优先从 `TraceIDProvider` 获取（如果已注册）
2. 如果没有，自动生成符合 W3C 标准的 32 位十六进制 trace_id
3. 用于定时任务、初始化等没有 context 的场景

**注意**: 自定义 Logger 实现应该调用此函数来获取 trace_id。

## 重要设计决策

### 为什么使用 _examples/ 目录

Go 语言的约定：下划线前缀目录会被 `go build ./...` 和 `go test ./...` 自动忽略。示例代码不应该作为主包的一部分被编译。

### 代码重复消除

`zerolog_logger.go` 中的 `addFields()` 方法消除了 8 个方法中的重复 type switch 代码。如果添加新的 Logger 实现（如 zap_logger.go），也应该使用类似模式。

### 适配器模式

`adapter/gormadapter/` 展示了如何将第三方库的日志重定向到 zllog。这是推荐的模式，因为：
- 自动包含 trace_id
- 统一日志格式
- 便于追踪全链路

## 版本发布

项目使用语义化版本。Breaking changes 必须升级次版本号（如 v1.2.0 → v1.3.0）。

示例 breaking change：
- 重命名公开 API（如 `Float()` → `Float64()`）
- 修改接口签名

## 相关文档

- `README.md` - 用户指南和完整 API 文档
- `docs/代码结构说明.md` - 详细的代码组织说明
- `docs/Trace_ID生成标准文档.md` - trace_id 生成规范
- `docs/文件日志规范.md` - 日志轮转和存储规范
