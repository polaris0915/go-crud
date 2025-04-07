# 日志组件 (Log Component)

## 简介

本日志组件是 CRUD 框架的核心日志管理系统，基于 Zap 日志库构建，提供灵活、高性能的日志记录解决方案。作为框架内置组件，无需额外安装，开箱即用。

## 功能特性

### 1. 日志级别
支持四种标准日志级别：
- `Debug`：调试信息
- `Info`：常规信息
- `Warn`：警告信息
- `Error`：错误信息

### 2. 日志输出配置
- 支持控制台输出
- 支持文件存储
- 可选 JSON 或控制台格式输出

### 3. 日志文件管理
- 按日期自动创建日志文件
- 单文件大小限制
- 日志文件保留策略
- 自动清理过期日志文件

### 4. 上下文信息
从 Gin 框架上下文中自动提取用户信息：
- `user_id`：用户标识
- `user_role`：用户角色
- `user_ip`：用户 IP 地址

## 使用方法

### 基本初始化

```go
import (
    "your-project/log"
    "go.uber.org/zap"
)

// 使用默认配置
log.InitLogger()

// 使用自定义配置
config := log.Config{
    Level:          "info",
    Format:         "json",
    Path:           "/path/to/logs",
    MaxSize:        100,  // 单个日志文件最大 100MB
    MaxAge:         30,   // 日志保留30天
    Compress:       false,
    ConsoleLogging: true,
}
log.InitLogger(config)
```

### 日志记录

#### 基本日志记录
```go
// 不带上下文
log.Debug("调试信息", zap.String("key", "value"))
log.Info("普通信息")
log.Warn("警告信息")
log.Error("错误信息", zap.Error(err))

// 带上下文（需要 Gin Context）
log.DebugWithContext(ctx, "带上下文的调试信息", zap.String("key", "value"))
log.InfoWithContext(ctx, "带上下文的普通信息")
```

### 配置方法

#### 设置日志级别
```go
// 动态设置日志级别
log.SetLogLevel("debug")
```

#### 设置调用者跳过层数
```go
// 调整日志输出的调用者信息
log.SetCallerSkip(1)
```

## 日志路径解析

日志路径按以下优先级解析：
1. 配置中指定的路径
2. 环境变量 `CRUD_LOG_PATH`
3. 程序执行目录下的 `log/logs` 目录

## 环境变量

- `CRUD_LOG_PATH`：自定义日志存储路径

## 注意事项

1. 并发场景下谨慎使用 `SetLogLevel()`
2. 日志组件会自动处理日志文件滚动
3. 默认日志级别为 `Info`

## 未来改进计划

- [ ] 改进错误处理机制
- [ ] 优化日志轮转性能
- [ ] 添加磁盘空间检查
- [ ] 增强并发安全性

## 技术细节

- 基于 Zap 日志库
- 支持灵活配置
- 高性能日志记录
- 上下文信息自动提取