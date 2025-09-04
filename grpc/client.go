package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientManager gRPC 客户端管理器
type ClientManager struct {
	connections map[string]*grpc.ClientConn
	configs     map[string]ClientConfig
	logger      *zap.Logger
	mu          sync.RWMutex
}

// NewClientManager 创建新的客户端管理器
// logger: 日志记录器
// 返回: ClientManager 实例
func NewClientManager(logger *zap.Logger) *ClientManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ClientManager{
		connections: make(map[string]*grpc.ClientConn),
		configs:     make(map[string]ClientConfig),
		logger:      logger,
	}
}

// AddClient 添加客户端配置
// name: 客户端名称
// config: 客户端配置
func (cm *ClientManager) AddClient(name string, config ClientConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.configs[name] = config
	cm.logger.Info("added gRPC client config", 
		zap.String("name", name),
		zap.String("target", config.Target),
	)
}

// GetClient 获取客户端连接
// name: 客户端名称
// 返回: gRPC 客户端连接和错误信息
func (cm *ClientManager) GetClient(name string) (*grpc.ClientConn, error) {
	cm.mu.RLock()
	conn, exists := cm.connections[name]
	config, configExists := cm.configs[name]
	cm.mu.RUnlock()

	if !configExists {
		return nil, fmt.Errorf("client config not found: %s", name)
	}

	if exists && conn != nil {
		// 检查连接状态
		if conn.GetState().String() != "SHUTDOWN" {
			return conn, nil
		}
		// 连接已关闭，需要重新创建
		cm.mu.Lock()
		delete(cm.connections, name)
		cm.mu.Unlock()
	}

	// 创建新连接
	conn, err := cm.createConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection for %s: %w", name, err)
	}

	cm.mu.Lock()
	cm.connections[name] = conn
	cm.mu.Unlock()

	cm.logger.Info("created gRPC client connection", 
		zap.String("name", name),
		zap.String("target", config.Target),
	)

	return conn, nil
}

// createConnection 创建 gRPC 连接
// config: 客户端配置
// 返回: gRPC 客户端连接和错误信息
func (cm *ClientManager) createConnection(config ClientConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	// 构建连接选项
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(config.MaxSendMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepAlive.Time,
			Timeout:             config.KeepAlive.Timeout,
			PermitWithoutStream: config.KeepAlive.PermitWithoutStream,
		}),
	}

	// 配置 TLS
	if config.TLS != nil && config.TLS.Enabled {
		tlsConfig, err := loadClientTLSConfig(config.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 创建连接
	conn, err := grpc.DialContext(ctx, config.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", config.Target, err)
	}

	return conn, nil
}

// CloseClient 关闭指定客户端连接
// name: 客户端名称
// 返回: 错误信息
func (cm *ClientManager) CloseClient(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[name]
	if !exists {
		return fmt.Errorf("client not found: %s", name)
	}

	err := conn.Close()
	delete(cm.connections, name)

	cm.logger.Info("closed gRPC client connection", zap.String("name", name))
	return err
}

// CloseAll 关闭所有客户端连接
// 返回: 错误信息
func (cm *ClientManager) CloseAll() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var lastErr error
	for name, conn := range cm.connections {
		if err := conn.Close(); err != nil {
			lastErr = err
			cm.logger.Error("failed to close client connection", 
				zap.String("client", name), 
				zap.Error(err),
			)
		} else {
			cm.logger.Info("closed client connection", zap.String("client", name))
		}
	}

	// 清空连接映射
	cm.connections = make(map[string]*grpc.ClientConn)
	cm.configs = make(map[string]ClientConfig)

	return lastErr
}

// ListClients 列出所有客户端名称
// 返回: 客户端名称列表
func (cm *ClientManager) ListClients() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.configs))
	for name := range cm.configs {
		names = append(names, name)
	}
	return names
}

// loadClientTLSConfig 加载客户端 TLS 配置
// config: TLS 配置
// 返回: TLS 配置和错误信息
func loadClientTLSConfig(config *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if config.ServerName != "" {
		tlsConfig.ServerName = config.ServerName
	}

	// 加载 CA 证书
	if config.CAFile != "" {
		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// 加载客户端证书
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}