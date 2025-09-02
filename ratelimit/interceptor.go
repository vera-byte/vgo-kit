package ratelimit

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// InterceptorConfig 拦截器配置
type InterceptorConfig struct {
	RateLimiter RateLimiter
	KeyFunc     KeyFunc
	SkipFunc    SkipFunc
}

// KeyFunc 生成限流key的函数
type KeyFunc func(ctx context.Context, info *grpc.UnaryServerInfo) string

// SkipFunc 判断是否跳过限流的函数
type SkipFunc func(ctx context.Context, info *grpc.UnaryServerInfo) bool

// DefaultKeyFunc 默认的key生成函数（基于IP地址）
func DefaultKeyFunc(ctx context.Context, info *grpc.UnaryServerInfo) string {
	// 尝试从peer获取客户端IP
	if p, ok := peer.FromContext(ctx); ok {
		if addr, ok := p.Addr.(*net.TCPAddr); ok {
			return fmt.Sprintf("ip:%s", addr.IP.String())
		}
	}

	// 尝试从metadata获取真实IP（通过代理时）
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if xForwardedFor := md.Get("x-forwarded-for"); len(xForwardedFor) > 0 {
			// 取第一个IP（原始客户端IP）
			ips := strings.Split(xForwardedFor[0], ",")
			if len(ips) > 0 {
				return fmt.Sprintf("ip:%s", strings.TrimSpace(ips[0]))
			}
		}
		if xRealIP := md.Get("x-real-ip"); len(xRealIP) > 0 {
			return fmt.Sprintf("ip:%s", xRealIP[0])
		}
	}

	return "unknown"
}

// UserKeyFunc 基于用户ID的key生成函数
func UserKeyFunc(ctx context.Context, info *grpc.UnaryServerInfo) string {
	// 尝试从metadata获取用户ID
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userID := md.Get("user-id"); len(userID) > 0 {
			return fmt.Sprintf("user:%s", userID[0])
		}
		if authorization := md.Get("authorization"); len(authorization) > 0 {
			// 这里可以解析JWT token获取用户ID
			// 为简化，直接使用token的hash
			return fmt.Sprintf("token:%s", authorization[0][:min(len(authorization[0]), 32)])
		}
	}

	// 回退到IP限流
	return DefaultKeyFunc(ctx, info)
}

// MethodKeyFunc 基于方法的key生成函数
func MethodKeyFunc(ctx context.Context, info *grpc.UnaryServerInfo) string {
	return fmt.Sprintf("method:%s", info.FullMethod)
}

// CombinedKeyFunc 组合多个维度的key生成函数
func CombinedKeyFunc(funcs ...KeyFunc) KeyFunc {
	return func(ctx context.Context, info *grpc.UnaryServerInfo) string {
		var parts []string
		for _, f := range funcs {
			parts = append(parts, f(ctx, info))
		}
		return strings.Join(parts, ":")
	}
}

// DefaultSkipFunc 默认的跳过函数（不跳过任何请求）
func DefaultSkipFunc(ctx context.Context, info *grpc.UnaryServerInfo) bool {
	return false
}

// HealthCheckSkipFunc 跳过健康检查的函数
func HealthCheckSkipFunc(ctx context.Context, info *grpc.UnaryServerInfo) bool {
	return strings.Contains(info.FullMethod, "Health")
}

// WhitelistSkipFunc 基于白名单的跳过函数
func WhitelistSkipFunc(methods []string) SkipFunc {
	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}
	return func(ctx context.Context, info *grpc.UnaryServerInfo) bool {
		return methodSet[info.FullMethod]
	}
}

// UnaryServerInterceptor 创建一元服务器拦截器
func UnaryServerInterceptor(config *InterceptorConfig) grpc.UnaryServerInterceptor {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	if config.SkipFunc == nil {
		config.SkipFunc = DefaultSkipFunc
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 检查是否跳过限流
		if config.SkipFunc(ctx, info) {
			return handler(ctx, req)
		}

		// 生成限流key
		key := config.KeyFunc(ctx, info)

		// 检查是否允许请求
		allowed, err := config.RateLimiter.Allow(ctx, key)
		if err != nil {
			// 限流器错误，记录日志但允许请求通过
			// 这里可以添加日志记录
			return handler(ctx, req)
		}

		if !allowed {
			// 请求被限流
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded for key: %s", key)
		}

		// 允许请求通过
		return handler(ctx, req)
	}
}

// StreamServerInterceptor 创建流服务器拦截器
func StreamServerInterceptor(config *InterceptorConfig) grpc.StreamServerInterceptor {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	if config.SkipFunc == nil {
		config.SkipFunc = DefaultSkipFunc
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// 将StreamServerInfo转换为UnaryServerInfo以复用逻辑
		unaryInfo := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: info.FullMethod,
		}

		// 检查是否跳过限流
		if config.SkipFunc(ctx, unaryInfo) {
			return handler(srv, ss)
		}

		// 生成限流key
		key := config.KeyFunc(ctx, unaryInfo)

		// 检查是否允许请求
		allowed, err := config.RateLimiter.Allow(ctx, key)
		if err != nil {
			// 限流器错误，记录日志但允许请求通过
			return handler(srv, ss)
		}

		if !allowed {
			// 请求被限流
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded for key: %s", key)
		}

		// 允许请求通过
		return handler(srv, ss)
	}
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
