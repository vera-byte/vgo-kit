package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Server gRPC 服务端
type Server struct {
	config   ServerConfig
	server   *grpc.Server
	listener net.Listener
	logger   *zap.Logger
}

// NewServer 创建新的 gRPC 服务端
// config: 服务端配置
// logger: 日志记录器
// interceptors: 可选的拦截器选项
// 返回: Server 实例和错误信息
func NewServer(config ServerConfig, logger *zap.Logger, interceptors ...InterceptorOption) (*Server, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// 构建 gRPC 服务器选项
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
		grpc.ConnectionTimeout(config.ConnectionTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.KeepAlive.Time,
			Timeout: config.KeepAlive.Timeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.KeepAlive.Time,
			PermitWithoutStream: config.KeepAlive.PermitWithoutStream,
		}),
	}

	// 应用拦截器
	interceptorConfig := &InterceptorConfig{}
	for _, opt := range interceptors {
		opt(interceptorConfig)
	}

	// 添加拦截器到服务器选项
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	if interceptorConfig.RecoveryInterceptor != nil {
		unaryInterceptors = append(unaryInterceptors, interceptorConfig.RecoveryInterceptor.UnaryServerInterceptor())
		streamInterceptors = append(streamInterceptors, interceptorConfig.RecoveryInterceptor.StreamServerInterceptor())
	}

	if interceptorConfig.LoggingInterceptor != nil {
		unaryInterceptors = append(unaryInterceptors, interceptorConfig.LoggingInterceptor.UnaryServerInterceptor())
		streamInterceptors = append(streamInterceptors, interceptorConfig.LoggingInterceptor.StreamServerInterceptor())
	}

	// 添加自定义拦截器
	unaryInterceptors = append(unaryInterceptors, interceptorConfig.CustomUnaryInterceptors...)
	streamInterceptors = append(streamInterceptors, interceptorConfig.CustomStreamInterceptors...)

	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	// 配置 TLS
	if config.TLS != nil && config.TLS.Enabled {
		tlsConfig, err := loadServerTLSConfig(config.TLS)
		if err != nil {
			listener.Close()
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.Creds(creds))
	}

	// 创建 gRPC 服务器
	server := grpc.NewServer(opts...)

	return &Server{
		config:   config,
		server:   server,
		listener: listener,
		logger:   logger,
	}, nil
}

// GetServer 获取底层的 gRPC 服务器实例
// 返回: gRPC 服务器实例
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

// RegisterService 注册 gRPC 服务
// desc: 服务描述
// impl: 服务实现
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
	s.logger.Info("registered gRPC service", zap.String("service", desc.ServiceName))
}

// Start 启动 gRPC 服务端
// 返回: 错误信息
func (s *Server) Start() error {
	s.logger.Info("starting gRPC server", 
		zap.String("address", s.listener.Addr().String()),
		zap.Bool("tls_enabled", s.config.TLS != nil && s.config.TLS.Enabled),
	)

	return s.server.Serve(s.listener)
}

// Stop 停止 gRPC 服务端
// ctx: 上下文
// 返回: 错误信息
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping gRPC server")

	// 创建一个通道来接收停止完成信号
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	// 等待优雅停止或上下文超时
	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Warn("gRPC server stop timeout, forcing stop")
		s.server.Stop()
		return ctx.Err()
	}
}

// GetListener 获取监听器
// 返回: 网络监听器
func (s *Server) GetListener() net.Listener {
	return s.listener
}

// loadServerTLSConfig 加载服务端 TLS 配置
// config: TLS 配置
// 返回: TLS 配置和错误信息
func loadServerTLSConfig(config *TLSConfig) (*tls.Config, error) {
	if config.CertFile == "" || config.KeyFile == "" {
		return nil, fmt.Errorf("cert_file and key_file are required for server TLS")
	}

	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if config.ServerName != "" {
		tlsConfig.ServerName = config.ServerName
	}

	return tlsConfig, nil
}