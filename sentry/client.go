package sentry

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

// Client Sentry客户端接口
// 定义Sentry错误监控的核心功能
type Client interface {
	// CaptureException 捕获异常
	// 参数: err error 错误对象
	// 返回: *sentry.EventID 事件ID
	CaptureException(err error) *sentry.EventID
	
	// CaptureMessage 捕获消息
	// 参数: message string 消息内容, level sentry.Level 日志级别
	// 返回: *sentry.EventID 事件ID
	CaptureMessage(message string, level sentry.Level) *sentry.EventID
	
	// WithContext 设置上下文
	// 参数: ctx context.Context 上下文
	// 返回: Client 带上下文的客户端
	WithContext(ctx context.Context) Client
	
	// WithTag 添加标签
	// 参数: key string 标签键, value string 标签值
	// 返回: Client 带标签的客户端
	WithTag(key, value string) Client
	
	// WithTags 添加多个标签
	// 参数: tags map[string]string 标签映射
	// 返回: Client 带标签的客户端
	WithTags(tags map[string]string) Client
	
	// WithExtra 添加额外信息
	// 参数: key string 键, value interface{} 值
	// 返回: Client 带额外信息的客户端
	WithExtra(key string, value interface{}) Client
	
	// WithUser 设置用户信息
	// 参数: user sentry.User 用户信息
	// 返回: Client 带用户信息的客户端
	WithUser(user sentry.User) Client
	
	// Flush 刷新缓冲区
	// 参数: timeout time.Duration 超时时间
	// 返回: bool 是否成功
	Flush(timeout time.Duration) bool
	
	// Close 关闭客户端
	Close()
}

// SentryClient Sentry客户端实现
type SentryClient struct {
	hub *sentry.Hub
}

// NewClient 创建新的Sentry客户端
// 参数: config *SentryConfig Sentry配置
// 返回: Client Sentry客户端, error 错误信息
func NewClient(config *SentryConfig) (Client, error) {
	if config == nil {
		config = DefaultSentryConfig()
	}

	if !config.GetEnabled() {
		return &NoOpClient{}, nil
	}

	if config.GetDSN() == "" {
		return nil, fmt.Errorf("sentry DSN is required when enabled")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         config.GetDSN(),
		Environment: config.GetEnvironment(),
		Debug:       config.GetDebug(),
		SampleRate:  config.GetSampleRate(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize sentry: %w", err)
	}

	return &SentryClient{
		hub: sentry.CurrentHub().Clone(),
	}, nil
}

// CaptureException 捕获异常
// 参数: err error 错误对象
// 返回: *sentry.EventID 事件ID
func (c *SentryClient) CaptureException(err error) *sentry.EventID {
	return c.hub.CaptureException(err)
}

// CaptureMessage 捕获消息
// 参数: message string 消息内容, level sentry.Level 日志级别
// 返回: *sentry.EventID 事件ID
func (c *SentryClient) CaptureMessage(message string, level sentry.Level) *sentry.EventID {
	c.hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
	})
	return c.hub.CaptureMessage(message)
}

// WithContext 设置上下文
// 参数: ctx context.Context 上下文
// 返回: Client 带上下文的客户端
func (c *SentryClient) WithContext(ctx context.Context) Client {
	newHub := c.hub.Clone()
	return &SentryClient{
		hub: newHub,
	}
}

// WithTag 添加标签
// 参数: key string 标签键, value string 标签值
// 返回: Client 带标签的客户端
func (c *SentryClient) WithTag(key, value string) Client {
	newHub := c.hub.Clone()
	newHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
	return &SentryClient{hub: newHub}
}

// WithTags 添加多个标签
// 参数: tags map[string]string 标签映射
// 返回: Client 带标签的客户端
func (c *SentryClient) WithTags(tags map[string]string) Client {
	newHub := c.hub.Clone()
	newHub.ConfigureScope(func(scope *sentry.Scope) {
		for key, value := range tags {
			scope.SetTag(key, value)
		}
	})
	return &SentryClient{hub: newHub}
}

// WithExtra 添加额外信息
// 参数: key string 键, value interface{} 值
// 返回: Client 带额外信息的客户端
func (c *SentryClient) WithExtra(key string, value interface{}) Client {
	newHub := c.hub.Clone()
	newHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetExtra(key, value)
	})
	return &SentryClient{hub: newHub}
}

// WithUser 设置用户信息
// 参数: user sentry.User 用户信息
// 返回: Client 带用户信息的客户端
func (c *SentryClient) WithUser(user sentry.User) Client {
	newHub := c.hub.Clone()
	newHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(user)
	})
	return &SentryClient{hub: newHub}
}

// Flush 刷新缓冲区
// 参数: timeout time.Duration 超时时间
// 返回: bool 是否成功
func (c *SentryClient) Flush(timeout time.Duration) bool {
	return c.hub.Flush(timeout)
}

// Close 关闭客户端
func (c *SentryClient) Close() {
	c.hub.Flush(2 * time.Second)
}

// NoOpClient 空操作客户端，用于禁用Sentry时
type NoOpClient struct{}

// CaptureException 空操作捕获异常
func (n *NoOpClient) CaptureException(err error) *sentry.EventID {
	return nil
}

// CaptureMessage 空操作捕获消息
func (n *NoOpClient) CaptureMessage(message string, level sentry.Level) *sentry.EventID {
	return nil
}

// WithContext 空操作设置上下文
func (n *NoOpClient) WithContext(ctx context.Context) Client {
	return n
}

// WithTag 空操作添加标签
func (n *NoOpClient) WithTag(key, value string) Client {
	return n
}

// WithTags 空操作添加多个标签
func (n *NoOpClient) WithTags(tags map[string]string) Client {
	return n
}

// WithExtra 空操作添加额外信息
func (n *NoOpClient) WithExtra(key string, value interface{}) Client {
	return n
}

// WithUser 空操作设置用户信息
func (n *NoOpClient) WithUser(user sentry.User) Client {
	return n
}

// Flush 空操作刷新缓冲区
func (n *NoOpClient) Flush(timeout time.Duration) bool {
	return true
}

// Close 空操作关闭客户端
func (n *NoOpClient) Close() {
	// 无操作
}