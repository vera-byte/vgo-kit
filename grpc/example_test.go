package grpc_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	grpckit "github.com/vera-byte/vgo-kit/grpc"
)

// ExampleServer 演示如何使用 gRPC 服务端
func ExampleServer() {
	// 创建日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建 gRPC 配置
	config := &grpckit.Config{
		Server: grpckit.ServerConfig{
			Host:               "localhost",
			Port:               50051,
			MaxRecvMsgSize:     4 * 1024 * 1024, // 4MB
			MaxSendMsgSize:     4 * 1024 * 1024, // 4MB
			ConnectionTimeout:  30 * time.Second,
			KeepAlive: grpckit.KeepAliveConfig{
				Time:                30 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			},
		},
		Clients: make(map[string]grpckit.ClientConfig),
	}

	// 创建 gRPC 管理器
	manager, err := grpckit.NewManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC manager: %v", err)
	}

	// 初始化服务端，添加日志和恢复拦截器
	err = manager.InitServer(
		grpckit.WithLoggingInterceptor(logger),
		grpckit.WithRecoveryInterceptor(logger),
	)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// 注册服务（这里需要你的实际服务实现）
	// manager.RegisterService(&pb.YourServiceDesc, &yourServiceImpl{})

	// 启动服务端
	go func() {
		if err := manager.StartServer(); err != nil {
			logger.Error("Server failed", zap.Error(err))
		}
	}()

	// 等待一段时间后停止服务端
	time.Sleep(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := manager.StopServer(ctx); err != nil {
		logger.Error("Failed to stop server", zap.Error(err))
	}

	fmt.Println("Server example completed")
}

// ExampleClient 演示如何使用 gRPC 客户端
func ExampleClient() {
	// 创建日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建 gRPC 配置
	config := &grpckit.Config{
		Server: grpckit.ServerConfig{}, // 客户端不需要服务端配置
		Clients: map[string]grpckit.ClientConfig{
			"user-service": {
				Target:            "localhost:50051",
				ConnectionTimeout: 10 * time.Second,
				KeepAlive: grpckit.KeepAliveConfig{
					Time:                30 * time.Second,
					Timeout:             5 * time.Second,
					PermitWithoutStream: true,
				},
				Retry: grpckit.RetryConfig{
					MaxAttempts:       3,
					InitialBackoff:    1 * time.Second,
					MaxBackoff:        30 * time.Second,
					BackoffMultiplier: 2.0,
				},
			},
			"order-service": {
				Target:            "localhost:50052",
				ConnectionTimeout: 10 * time.Second,
				KeepAlive: grpckit.KeepAliveConfig{
					Time:                30 * time.Second,
					Timeout:             5 * time.Second,
					PermitWithoutStream: true,
				},
			},
		},
	}

	// 创建 gRPC 管理器
	manager, err := grpckit.NewManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC manager: %v", err)
	}

	// 获取用户服务客户端连接
	_, err = manager.GetClient("user-service")
	if err != nil {
		log.Fatalf("Failed to get user service client: %v", err)
	}

	// 使用客户端连接创建服务客户端
	// userClient := pb.NewUserServiceClient(userConn)

	// 获取订单服务客户端连接
	_, err = manager.GetClient("order-service")
	if err != nil {
		log.Fatalf("Failed to get order service client: %v", err)
	}

	// 使用客户端连接创建服务客户端
	// orderClient := pb.NewOrderServiceClient(orderConn)

	// 动态添加新的客户端
	manager.AddClient("payment-service", grpckit.ClientConfig{
		Target:            "localhost:50053",
		ConnectionTimeout: 10 * time.Second,
	})

	// 列出所有客户端
	clients := manager.ListClients()
	fmt.Printf("Available clients: %v\n", clients)

	// 关闭管理器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := manager.Close(ctx); err != nil {
		logger.Error("Failed to close manager", zap.Error(err))
	}

	fmt.Println("Client example completed")

	// 输出示例:
	// Available clients: [user-service order-service payment-service]
	// Client example completed
}

// ExampleTLS 演示如何使用 TLS 配置
func ExampleTLS() {
	// 创建日志记录器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建带 TLS 的 gRPC 配置
	config := &grpckit.Config{
		Server: grpckit.ServerConfig{
			Host: "localhost",
			Port: 50051,
			TLS: &grpckit.TLSConfig{
				Enabled:  true,
				CertFile: "/path/to/server.crt",
				KeyFile:  "/path/to/server.key",
				CAFile:   "/path/to/ca.crt",
			},
		},
		Clients: map[string]grpckit.ClientConfig{
			"secure-service": {
				Target: "localhost:50051",
				TLS: &grpckit.TLSConfig{
					Enabled:            true,
					CAFile:             "/path/to/ca.crt",
					InsecureSkipVerify: false, // 生产环境应设为 false
					ServerName:         "your-server-name",
				},
			},
		},
	}

	// 创建 gRPC 管理器
	_, err := grpckit.NewManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC manager: %v", err)
	}

	fmt.Println("TLS configuration example completed")

	// 注意: 实际使用时需要提供真实的证书文件路径
}

// ExampleConfiguration 演示配置文件的使用
func ExampleConfiguration() {
	// 示例配置文件内容 (config.yaml):
	/*
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
	      address: "user-service:50051"
	      connection_timeout: "10s"
	      keepalive:
	        time: "30s"
	        timeout: "5s"
	        permit_without_stream: true
	      retry:
	        max_attempts: 3
	        backoff: "1s"
	    order-service:
	      address: "order-service:50052"
	      connection_timeout: "10s"
	*/

	fmt.Println("Configuration example completed")
}