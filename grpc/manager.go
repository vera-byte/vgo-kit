package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Manager gRPC 管理器
type Manager struct {
	config        *Config
	server        *Server
	clientManager *ClientManager
	logger        *zap.Logger
}

// NewManager 创建新的 gRPC 管理器
// config: gRPC 配置
// logger: 日志记录器
// 返回: Manager 实例和错误信息
func NewManager(config *Config, logger *zap.Logger) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	// 创建客户端管理器
	clientManager := NewClientManager(logger)

	// 添加客户端配置
	for name, clientConfig := range config.Clients {
		clientManager.AddClient(name, clientConfig)
	}

	return &Manager{
		config:        config,
		clientManager: clientManager,
		logger:        logger,
	}, nil
}

// InitServer 初始化 gRPC 服务端
// interceptors: 可选的拦截器配置
// 返回: 错误信息
func (m *Manager) InitServer(interceptors ...InterceptorOption) error {
	if m.server != nil {
		return fmt.Errorf("server already initialized")
	}

	// 应用拦截器配置
	interceptorConfig := &InterceptorConfig{}
	for _, opt := range interceptors {
		opt(interceptorConfig)
	}

	// 创建服务端配置副本并应用拦截器
	serverConfig := m.config.Server
	server, err := NewServer(serverConfig, m.logger, interceptors...)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	m.server = server
	m.logger.Info("gRPC server initialized", 
		zap.String("address", fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)),
	)

	return nil
}

// GetServer 获取 gRPC 服务端
// 返回: Server 实例
func (m *Manager) GetServer() *Server {
	return m.server
}

// RegisterService 注册 gRPC 服务
// desc: 服务描述
// impl: 服务实现
// 返回: 错误信息
func (m *Manager) RegisterService(desc *grpc.ServiceDesc, impl interface{}) error {
	if m.server == nil {
		return fmt.Errorf("server not initialized, call InitServer first")
	}

	m.server.RegisterService(desc, impl)
	return nil
}

// StartServer 启动 gRPC 服务端
// 返回: 错误信息
func (m *Manager) StartServer() error {
	if m.server == nil {
		return fmt.Errorf("server not initialized, call InitServer first")
	}

	return m.server.Start()
}

// StopServer 停止 gRPC 服务端
// ctx: 上下文
// 返回: 错误信息
func (m *Manager) StopServer(ctx context.Context) error {
	if m.server == nil {
		return nil
	}

	return m.server.Stop(ctx)
}

// GetClient 获取 gRPC 客户端连接
// name: 客户端名称
// 返回: gRPC 客户端连接和错误信息
func (m *Manager) GetClient(name string) (*grpc.ClientConn, error) {
	return m.clientManager.GetClient(name)
}

// AddClient 添加客户端配置
// name: 客户端名称
// config: 客户端配置
func (m *Manager) AddClient(name string, config ClientConfig) {
	m.clientManager.AddClient(name, config)
	m.config.Clients[name] = config
}

// RemoveClient 移除客户端
// name: 客户端名称
// 返回: 错误信息
func (m *Manager) RemoveClient(name string) error {
	err := m.clientManager.CloseClient(name)
	delete(m.config.Clients, name)
	return err
}

// ListClients 列出所有客户端名称
// 返回: 客户端名称列表
func (m *Manager) ListClients() []string {
	return m.clientManager.ListClients()
}

// Close 关闭管理器，清理所有资源
// ctx: 上下文
// 返回: 错误信息
func (m *Manager) Close(ctx context.Context) error {
	var lastErr error

	// 停止服务端
	if m.server != nil {
		if err := m.server.Stop(ctx); err != nil {
			lastErr = err
			m.logger.Error("failed to stop gRPC server", zap.Error(err))
		}
	}

	// 关闭所有客户端连接
	if err := m.clientManager.CloseAll(); err != nil {
		lastErr = err
		m.logger.Error("failed to close gRPC clients", zap.Error(err))
	}

	m.logger.Info("gRPC manager closed")
	return lastErr
}

// InterceptorConfig 拦截器配置
type InterceptorConfig struct {
	LoggingInterceptor    *LoggingInterceptor
	RecoveryInterceptor   *RecoveryInterceptor
	CustomUnaryInterceptors []grpc.UnaryServerInterceptor
	CustomStreamInterceptors []grpc.StreamServerInterceptor
}

// InterceptorOption 拦截器选项
type InterceptorOption func(*InterceptorConfig)

// WithLoggingInterceptor 添加日志拦截器
// logger: 日志记录器
// 返回: 拦截器选项
func WithLoggingInterceptor(logger *zap.Logger) InterceptorOption {
	return func(config *InterceptorConfig) {
		config.LoggingInterceptor = NewLoggingInterceptor(logger)
	}
}

// WithRecoveryInterceptor 添加恢复拦截器
// logger: 日志记录器
// 返回: 拦截器选项
func WithRecoveryInterceptor(logger *zap.Logger) InterceptorOption {
	return func(config *InterceptorConfig) {
		config.RecoveryInterceptor = NewRecoveryInterceptor(logger)
	}
}

// WithCustomUnaryInterceptor 添加自定义一元拦截器
// interceptor: 自定义一元拦截器
// 返回: 拦截器选项
func WithCustomUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) InterceptorOption {
	return func(config *InterceptorConfig) {
		config.CustomUnaryInterceptors = append(config.CustomUnaryInterceptors, interceptor)
	}
}

// WithCustomStreamInterceptor 添加自定义流拦截器
// interceptor: 自定义流拦截器
// 返回: 拦截器选项
func WithCustomStreamInterceptor(interceptor grpc.StreamServerInterceptor) InterceptorOption {
	return func(config *InterceptorConfig) {
		config.CustomStreamInterceptors = append(config.CustomStreamInterceptors, interceptor)
	}
}