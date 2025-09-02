package i18n

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// InterceptorConfig 国际化拦截器配置
type InterceptorConfig struct {
	// Translator 翻译器实例
	Translator Translator
	// DefaultLanguage 默认语言
	DefaultLanguage SupportedLanguage
	// LanguageHeader 语言头名称
	LanguageHeader string
}

// DefaultInterceptorConfig 默认拦截器配置
// 返回值:
//   - *InterceptorConfig: 默认配置
func DefaultInterceptorConfig() *InterceptorConfig {
	return &InterceptorConfig{
		Translator:      NewTranslator(DefaultLanguage),
		DefaultLanguage: DefaultLanguage,
		LanguageHeader:  "accept-language",
	}
}

// UnaryServerInterceptor 一元服务器拦截器
// 参数:
//   - config: 拦截器配置
//
// 返回值:
//   - grpc.UnaryServerInterceptor: gRPC一元服务器拦截器
func UnaryServerInterceptor(config *InterceptorConfig) grpc.UnaryServerInterceptor {
	if config == nil {
		config = DefaultInterceptorConfig()
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从metadata中提取语言信息
		lang := extractLanguageFromMetadata(ctx, config)

		// 将语言信息添加到上下文中
		ctx = SetLanguageToContext(ctx, lang)

		// 设置翻译器的当前语言
		config.Translator.SetLanguage(lang)

		// 调用下一个处理器
		return handler(ctx, req)
	}
}

// StreamServerInterceptor 流服务器拦截器
// 参数:
//   - config: 拦截器配置
//
// 返回值:
//   - grpc.StreamServerInterceptor: gRPC流服务器拦截器
func StreamServerInterceptor(config *InterceptorConfig) grpc.StreamServerInterceptor {
	if config == nil {
		config = DefaultInterceptorConfig()
	}

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 从metadata中提取语言信息
		lang := extractLanguageFromMetadata(ss.Context(), config)

		// 创建包装的流，将语言信息添加到上下文中
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          SetLanguageToContext(ss.Context(), lang),
		}

		// 设置翻译器的当前语言
		config.Translator.SetLanguage(lang)

		// 调用下一个处理器
		return handler(srv, wrappedStream)
	}
}

// wrappedServerStream 包装的服务器流
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回流的上下文
// 返回值:
//   - context.Context: 上下文
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// extractLanguageFromMetadata 从metadata中提取语言信息
// 参数:
//   - ctx: 上下文
//   - config: 拦截器配置
//
// 返回值:
//   - SupportedLanguage: 提取的语言
func extractLanguageFromMetadata(ctx context.Context, config *InterceptorConfig) SupportedLanguage {
	// 从gRPC metadata中获取语言信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return config.DefaultLanguage
	}

	// 检查语言头
	langHeaders := md.Get(config.LanguageHeader)
	if len(langHeaders) == 0 {
		// 尝试其他常见的语言头
		langHeaders = md.Get("language")
		if len(langHeaders) == 0 {
			langHeaders = md.Get("lang")
		}
	}

	if len(langHeaders) > 0 {
		// 解析Accept-Language格式
		return ParseAcceptLanguage(langHeaders[0])
	}

	return config.DefaultLanguage
}

// UnaryClientInterceptor 一元客户端拦截器
// 参数:
//   - lang: 要设置的语言
//
// 返回值:
//   - grpc.UnaryClientInterceptor: gRPC一元客户端拦截器
func UnaryClientInterceptor(lang SupportedLanguage) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// 将语言信息添加到outgoing metadata中
		ctx = metadata.AppendToOutgoingContext(ctx, "accept-language", string(lang))
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor 流客户端拦截器
// 参数:
//   - lang: 要设置的语言
//
// 返回值:
//   - grpc.StreamClientInterceptor: gRPC流客户端拦截器
func StreamClientInterceptor(lang SupportedLanguage) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// 将语言信息添加到outgoing metadata中
		ctx = metadata.AppendToOutgoingContext(ctx, "accept-language", string(lang))
		return streamer(ctx, desc, cc, method, opts...)
	}
}
