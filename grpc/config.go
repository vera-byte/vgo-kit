package grpc

import (
	"time"
)

// Config gRPC 配置结构
type Config struct {
	// Server 服务端配置
	Server ServerConfig `mapstructure:"server" yaml:"server"`
	// Clients 客户端配置
	Clients map[string]ClientConfig `mapstructure:"clients" yaml:"clients"`
}

// ServerConfig gRPC 服务端配置
type ServerConfig struct {
	// Host 监听地址
	Host string `mapstructure:"host" yaml:"host"`
	// Port 监听端口
	Port int `mapstructure:"port" yaml:"port"`
	// MaxRecvMsgSize 最大接收消息大小 (字节)
	MaxRecvMsgSize int `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	// MaxSendMsgSize 最大发送消息大小 (字节)
	MaxSendMsgSize int `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size"`
	// ConnectionTimeout 连接超时时间
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout" yaml:"connection_timeout"`
	// KeepAlive 保活配置
	KeepAlive KeepAliveConfig `mapstructure:"keepalive" yaml:"keepalive"`
	// TLS TLS配置
	TLS *TLSConfig `mapstructure:"tls" yaml:"tls"`
}

// ClientConfig gRPC 客户端配置
type ClientConfig struct {
	// Target 目标地址 (host:port)
	Target string `mapstructure:"target" yaml:"target"`
	// MaxRecvMsgSize 最大接收消息大小 (字节)
	MaxRecvMsgSize int `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	// MaxSendMsgSize 最大发送消息大小 (字节)
	MaxSendMsgSize int `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size"`
	// ConnectionTimeout 连接超时时间
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout" yaml:"connection_timeout"`
	// KeepAlive 保活配置
	KeepAlive KeepAliveConfig `mapstructure:"keepalive" yaml:"keepalive"`
	// TLS TLS配置
	TLS *TLSConfig `mapstructure:"tls" yaml:"tls"`
	// Retry 重试配置
	Retry RetryConfig `mapstructure:"retry" yaml:"retry"`
}

// KeepAliveConfig 保活配置
type KeepAliveConfig struct {
	// Time 保活时间间隔
	Time time.Duration `mapstructure:"time" yaml:"time"`
	// Timeout 保活超时时间
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
	// PermitWithoutStream 是否允许在没有活跃流时发送保活
	PermitWithoutStream bool `mapstructure:"permit_without_stream" yaml:"permit_without_stream"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	// Enabled 是否启用TLS
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	// CertFile 证书文件路径
	CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
	// KeyFile 私钥文件路径
	KeyFile string `mapstructure:"key_file" yaml:"key_file"`
	// CAFile CA证书文件路径
	CAFile string `mapstructure:"ca_file" yaml:"ca_file"`
	// ServerName 服务器名称 (用于证书验证)
	ServerName string `mapstructure:"server_name" yaml:"server_name"`
	// InsecureSkipVerify 是否跳过证书验证
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	// MaxAttempts 最大重试次数
	MaxAttempts int `mapstructure:"max_attempts" yaml:"max_attempts"`
	// InitialBackoff 初始退避时间
	InitialBackoff time.Duration `mapstructure:"initial_backoff" yaml:"initial_backoff"`
	// MaxBackoff 最大退避时间
	MaxBackoff time.Duration `mapstructure:"max_backoff" yaml:"max_backoff"`
	// BackoffMultiplier 退避倍数
	BackoffMultiplier float64 `mapstructure:"backoff_multiplier" yaml:"backoff_multiplier"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:              "0.0.0.0",
			Port:              9000,
			MaxRecvMsgSize:    4 * 1024 * 1024, // 4MB
			MaxSendMsgSize:    4 * 1024 * 1024, // 4MB
			ConnectionTimeout: 5 * time.Second,
			KeepAlive: KeepAliveConfig{
				Time:                30 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			},
		},
		Clients: make(map[string]ClientConfig),
	}
}

// DefaultClientConfig 返回默认客户端配置
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		MaxRecvMsgSize:    4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:    4 * 1024 * 1024, // 4MB
		ConnectionTimeout: 5 * time.Second,
		KeepAlive: KeepAliveConfig{
			Time:                30 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		},
		Retry: RetryConfig{
			MaxAttempts:       3,
			InitialBackoff:    100 * time.Millisecond,
			MaxBackoff:        30 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}
}