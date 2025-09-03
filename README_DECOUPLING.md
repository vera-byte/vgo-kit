# VGO-Kit 日志与Sentry解耦重构

## 概述

本次重构将 vgo-kit 中的日志系统和 Sentry 错误监控系统进行了解耦，提供了更灵活、可扩展的架构设计。

## 架构变更

### 重构前
- `sentry` 包同时包含日志和 Sentry 功能
- 日志系统与 Sentry 紧密耦合
- 难以独立使用日志或 Sentry 功能

### 重构后
- `logger` 包：专注于日志功能，提供接口抽象
- `sentry` 包：专注于错误监控，提供独立的客户端
- 两个模块完全解耦，可独立使用

## 新架构组件

### 1. Logger 模块 (`/logger`)

#### 核心接口
```go
type Logger interface {
    Debug(msg string, fields ...interface{})
    Info(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Fatal(msg string, fields ...interface{})
    WithFields(fields map[string]interface{}) Logger
    WithRequestID(requestID string) Logger
    Sync() error
    Close() error
}
```

#### 配置接口
```go
type Config interface {
    GetLevel() string
    GetDirectory() string
    GetFilename() string
    GetStdout() bool
    GetMaxSize() int
    GetMaxBackups() int
    GetMaxAge() int
    GetCompress() bool
}
```

#### 工厂接口
```go
type Factory interface {
    CreateLogger(config Config) (Logger, error)
}
```

### 2. Sentry 模块 (`/sentry`)

#### 核心接口
```go
type Client interface {
    CaptureException(err error) *sentry.EventID
    CaptureMessage(message string, level sentry.Level) *sentry.EventID
    WithContext(ctx context.Context) Client
    WithTag(key, value string) Client
    WithTags(tags map[string]string) Client
    WithExtra(key string, value interface{}) Client
    WithUser(user sentry.User) Client
    Flush(timeout time.Duration) bool
    Close()
}
```

#### 配置结构
```go
type SentryConfig struct {
    Enabled     bool    `yaml:"enabled" json:"enabled"`
    DSN         string  `yaml:"dsn" json:"dsn"`
    Environment string  `yaml:"environment" json:"environment"`
    Debug       bool    `yaml:"debug" json:"debug"`
    SampleRate  float64 `yaml:"sample_rate" json:"sample_rate"`
}
```

## 使用方法

### 1. 独立使用日志系统

```go
import "github.com/vera-byte/vgo-kit/logger"

// 使用默认配置
logFactory := logger.NewZapFactory()
log, err := logFactory.CreateLogger(nil)
if err != nil {
    panic(err)
}

// 记录日志
log.Info("应用启动")
log.Error("发生错误", map[string]interface{}{
    "error": err,
    "user_id": 123,
})
```

### 2. 独立使用Sentry

```go
import "github.com/vera-byte/vgo-kit/sentry"

// 创建Sentry客户端
config := &sentry.SentryConfig{
    Enabled:     true,
    DSN:         "your-sentry-dsn",
    Environment: "production",
    Debug:       false,
    SampleRate:  1.0,
}

client, err := sentry.NewClient(config)
if err != nil {
    panic(err)
}

// 捕获错误
client.CaptureException(err)
client.CaptureMessage("用户登录失败", sentry.LevelWarning)
```

### 3. 组合使用（推荐）

```go
import (
    "github.com/vera-byte/vgo-kit/logger"
    "github.com/vera-byte/vgo-kit/sentry"
)

// 初始化日志
logFactory := logger.NewZapFactory()
log, _ := logFactory.CreateLogger(nil)

// 初始化Sentry
sentryClient, _ := sentry.NewClient(&sentry.SentryConfig{
    Enabled: true,
    DSN:     "your-dsn",
})

// 记录错误日志并发送到Sentry
func handleError(err error, msg string) {
    // 记录到日志文件
    log.Error(msg, map[string]interface{}{"error": err})
    
    // 发送到Sentry
    sentryClient.CaptureException(err)
}
```

## 配置文件示例

### config.yaml
```yaml
# 日志配置
log:
  level: "info"
  directory: "./logs"
  filename: "app.log"
  stdout: true
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true

# Sentry配置
sentry:
  enabled: true
  dsn: "https://your-sentry-dsn@sentry.io/project-id"
  environment: "production"
  debug: false
  sample_rate: 1.0
```

## 迁移指南

### 从旧版本迁移

1. **更新导入**：
   ```go
   // 旧版本
   import "github.com/vera-byte/vgo-kit/sentry"
   
   // 新版本
   import (
       "github.com/vera-byte/vgo-kit/logger"
       "github.com/vera-byte/vgo-kit/sentry"
   )
   ```

2. **更新全局变量使用**：
   ```go
   // 旧版本
   vgokit.Log.Error("错误消息")
   
   // 新版本
   vgokit.Log.Error("错误消息", nil)
   // 如需发送到Sentry
   vgokit.SentryClient.CaptureMessage("错误消息", sentry.LevelError)
   ```

3. **更新配置文件**：
   - 将原来的嵌套配置分离为独立的 `log` 和 `sentry` 配置块

## 优势

1. **解耦设计**：日志和错误监控功能完全分离
2. **灵活配置**：可独立配置和使用各个组件
3. **接口抽象**：便于扩展和测试
4. **向后兼容**：通过适配器模式保持API兼容性
5. **性能优化**：减少不必要的依赖和初始化开销

## 测试

运行测试：
```bash
# 测试日志模块
go test ./logger/...

# 测试Sentry模块
go test ./sentry/...
```

## 注意事项

1. **配置验证**：确保Sentry DSN配置正确
2. **错误处理**：合理处理初始化失败的情况
3. **资源清理**：应用退出时调用 `Close()` 方法清理资源
4. **性能考虑**：在高并发场景下注意日志和Sentry的性能影响

## 后续计划

- [ ] 添加更多日志后端支持（如 logrus、zerolog）
- [ ] 支持更多错误监控服务（如 Rollbar、Bugsnag）
- [ ] 添加链路追踪集成
- [ ] 提供更多配置选项和优化