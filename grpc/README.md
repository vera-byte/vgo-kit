# gRPC 组件

vgo-kit 的 gRPC 组件提供了完整的 gRPC 服务端和客户端解决方案，支持配置化管理、拦截器、TLS 加密等功能。

## 功能特性

- 🚀 **统一管理**: 通过 Manager 统一管理服务端和客户端
- ⚙️ **配置化**: 支持 YAML 配置文件和代码配置
- 🔒 **安全**: 支持 TLS 加密和证书验证
- 📊 **监控**: 内置日志和恢复拦截器
- 🔄 **重试**: 客户端支持自动重试机制
- 🔗 **连接池**: 自动管理客户端连接
- 💪 **高性能**: 支持 Keep-Alive 和连接复用

## 快速开始

### 1. 基本使用

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go.uber.org/zap"
    grpckit "github.com/vera-byte/vgo-kit/grpc"
)

func main() {
    // 创建日志记录器
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()
    
    // 创建配置
    config := grpckit.DefaultConfig()
    config.Server.Host = "localhost"
    config.Server.Port = 50051
    
    // 创建管理器
    manager, err := grpckit.NewManager(config, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // 初始化服务端
    err = manager.InitServer(
        grpckit.WithLoggingInterceptor(logger),
        grpckit.WithRecoveryInterceptor(logger),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 注册服务
    // manager.RegisterService(&pb.YourServiceDesc, &yourServiceImpl{})
    
    // 启动服务端
    go func() {
        if err := manager.StartServer(); err != nil {
            logger.Error("Server failed", zap.Error(err))
        }
    }()
    
    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    defer manager.Close(ctx)
    
    // 保持运行
    select {}
}
```

### 2. 客户端使用

```go
// 添加客户端配置
manager.AddClient("user-service", grpckit.ClientConfig{
    Target: "localhost:50051",
    ConnectionTimeout: 10 * time.Second,
    Retry: grpckit.RetryConfig{
        MaxAttempts: 3,
        InitialBackoff: 1 * time.Second,
        MaxBackoff: 30 * time.Second,
        BackoffMultiplier: 2.0,
    },
})

// 获取客户端连接
conn, err := manager.GetClient("user-service")
if err != nil {
    log.Fatal(err)
}

// 使用连接创建服务客户端
// userClient := pb.NewUserServiceClient(conn)
```

### 3. 配置文件使用

在 `config.yaml` 中添加 gRPC 配置：

```yaml
grpc:
  server:
    host: "0.0.0.0"
    port: 50051
    max_recv_msg_size: 4194304  # 4MB
    max_send_msg_size: 4194304  # 4MB
    connection_timeout: "30s"
    keepalive:
      time: "30s"
      timeout: "5s"
      permit_without_stream: true
    tls:
      enabled: false
      cert_file: "/path/to/server.crt"
      key_file: "/path/to/server.key"
      ca_file: "/path/to/ca.crt"
  clients:
    user-service:
      target: "user-service:50051"
      connection_timeout: "10s"
      keepalive:
        time: "30s"
        timeout: "5s"
        permit_without_stream: true
      retry:
        max_attempts: 3
        initial_backoff: "1s"
        max_backoff: "30s"
        backoff_multiplier: 2.0
```

然后在代码中加载配置：

```go
import "github.com/vera-byte/vgo-kit/config"

type AppConfig struct {
    GRPC grpckit.Config `mapstructure:"grpc" yaml:"grpc"`
}

var cfg AppConfig
if err := config.LoadConfig(&cfg); err != nil {
    log.Fatal(err)
}

manager, err := grpckit.NewManager(&cfg.GRPC, logger)
```

## 高级功能

### TLS 配置

```go
// 服务端 TLS
config.Server.TLS = &grpckit.TLSConfig{
    Enabled:  true,
    CertFile: "/path/to/server.crt",
    KeyFile:  "/path/to/server.key",
    CAFile:   "/path/to/ca.crt",
}

// 客户端 TLS
clientConfig := grpckit.ClientConfig{
    Target: "secure-service:50051",
    TLS: &grpckit.TLSConfig{
        Enabled:            true,
        CAFile:             "/path/to/ca.crt",
        InsecureSkipVerify: false,
        ServerName:         "your-server-name",
    },
}
```

### 自定义拦截器

```go
// 创建自定义拦截器
func customInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 自定义逻辑
        return handler(ctx, req)
    }
}

// 在服务器选项中添加
opts := []grpc.ServerOption{
    grpc.UnaryInterceptor(customInterceptor()),
}
```

### 连接管理

```go
// 列出所有客户端
clients := manager.ListClients()
fmt.Printf("Available clients: %v\n", clients)

// 动态添加客户端
manager.AddClient("new-service", grpckit.ClientConfig{
    Target: "new-service:50053",
})

// 移除客户端
manager.RemoveClient("old-service")
```

## API 参考

### Manager

- `NewManager(config *Config, logger *zap.Logger) (*Manager, error)`: 创建管理器
- `InitServer(interceptors ...InterceptorOption) error`: 初始化服务端
- `StartServer() error`: 启动服务端
- `StopServer(ctx context.Context) error`: 停止服务端
- `RegisterService(desc *grpc.ServiceDesc, impl interface{}) error`: 注册服务
- `GetClient(name string) (*grpc.ClientConn, error)`: 获取客户端连接
- `AddClient(name string, config ClientConfig)`: 添加客户端
- `RemoveClient(name string) error`: 移除客户端
- `ListClients() []string`: 列出所有客户端
- `Close(ctx context.Context) error`: 关闭管理器

### 拦截器选项

- `WithLoggingInterceptor(logger *zap.Logger)`: 添加日志拦截器
- `WithRecoveryInterceptor(logger *zap.Logger)`: 添加恢复拦截器

### 配置结构

- `Config`: 主配置结构
- `ServerConfig`: 服务端配置
- `ClientConfig`: 客户端配置
- `KeepAliveConfig`: Keep-Alive 配置
- `TLSConfig`: TLS 配置
- `RetryConfig`: 重试配置

## 最佳实践

1. **使用配置文件**: 将 gRPC 配置放在配置文件中，便于不同环境的管理
2. **启用拦截器**: 使用日志和恢复拦截器提高系统的可观测性和稳定性
3. **配置重试**: 为客户端配置合理的重试策略
4. **TLS 加密**: 在生产环境中启用 TLS 加密
5. **连接复用**: 复用客户端连接，避免频繁创建和销毁
6. **优雅关闭**: 使用 context 控制服务的优雅关闭
7. **监控日志**: 关注连接状态和错误日志

## 注意事项

- 服务端必须先调用 `InitServer()` 再调用 `StartServer()`
- 客户端连接会自动管理，无需手动关闭
- TLS 证书路径必须是绝对路径
- 重试配置只对客户端有效
- 使用完毕后记得调用 `Close()` 方法清理资源

## 示例代码

更多示例代码请参考 `example_test.go` 文件。