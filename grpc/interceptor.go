package grpc

import (
	"context"
	"time"

	vgologger "github.com/vera-byte/vgo-kit/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor 日志拦截器配置
type LoggingInterceptor struct {
	logger *zap.Logger
}

// NewLoggingInterceptor 创建新的日志拦截器
// logger: 日志记录器
// 返回: LoggingInterceptor 实例
func NewLoggingInterceptor(logger *zap.Logger) *LoggingInterceptor {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LoggingInterceptor{logger: logger}
}

// UnaryServerInterceptor 一元服务端拦截器
// 返回: gRPC 一元服务端拦截器
func (li *LoggingInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		// 提取请求元数据
		md, _ := metadata.FromIncomingContext(ctx)
		requestID := getRequestID(md)
		
		li.logger.Info("gRPC request started",
			zap.String("method", info.FullMethod),
			zap.String("request_id", requestID),
		)
		
		// 调用处理器
		resp, err := handler(ctx, req)
		
		// 记录响应
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			li.logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("code", st.Code().String()),
				zap.String("message", st.Message()),
			)
		} else {
			li.logger.Info("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
			)
		}
		
		return resp, err
	}
}

// StreamServerInterceptor 流式服务端拦截器
// 返回: gRPC 流式服务端拦截器
func (li *LoggingInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		
		// 提取请求元数据
		md, _ := metadata.FromIncomingContext(stream.Context())
		requestID := getRequestID(md)
		
		li.logger.Info("gRPC stream started",
			zap.String("method", info.FullMethod),
			zap.String("request_id", requestID),
			zap.Bool("client_stream", info.IsClientStream),
			zap.Bool("server_stream", info.IsServerStream),
		)
		
		// 调用处理器
		err := handler(srv, stream)
		
		// 记录响应
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			li.logger.Error("gRPC stream failed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("code", st.Code().String()),
				zap.String("message", st.Message()),
			)
		} else {
			li.logger.Info("gRPC stream completed",
				zap.String("method", info.FullMethod),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
			)
		}
		
		return err
	}
}

// UnaryClientInterceptor 一元客户端拦截器
// 返回: gRPC 一元客户端拦截器
func (li *LoggingInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		
		// 添加请求ID到元数据（使用UUID确保全局唯一）
		ctx = addRequestIDToContext(ctx)
		
		li.logger.Info("gRPC client request started",
			zap.String("method", method),
			zap.String("target", cc.Target()),
		)
		
		// 调用远程方法
		err := invoker(ctx, method, req, reply, cc, opts...)
		
		// 记录响应
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			li.logger.Error("gRPC client request failed",
				zap.String("method", method),
				zap.String("target", cc.Target()),
				zap.Duration("duration", duration),
				zap.String("code", st.Code().String()),
				zap.String("message", st.Message()),
			)
		} else {
			li.logger.Info("gRPC client request completed",
				zap.String("method", method),
				zap.String("target", cc.Target()),
				zap.Duration("duration", duration),
			)
		}
		
		return err
	}
}

// StreamClientInterceptor 流式客户端拦截器
// 返回: gRPC 流式客户端拦截器
func (li *LoggingInterceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()
		
		// 添加请求ID到元数据（使用UUID确保全局唯一）
		ctx = addRequestIDToContext(ctx)
		
		li.logger.Info("gRPC client stream started",
			zap.String("method", method),
			zap.String("target", cc.Target()),
			zap.Bool("client_stream", desc.ClientStreams),
			zap.Bool("server_stream", desc.ServerStreams),
		)
		
		// 创建流
		stream, err := streamer(ctx, desc, cc, method, opts...)
		
		// 记录结果
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			li.logger.Error("gRPC client stream creation failed",
				zap.String("method", method),
				zap.String("target", cc.Target()),
				zap.Duration("duration", duration),
				zap.String("code", st.Code().String()),
				zap.String("message", st.Message()),
			)
		} else {
			li.logger.Info("gRPC client stream created",
				zap.String("method", method),
				zap.String("target", cc.Target()),
				zap.Duration("duration", duration),
			)
		}
		
		return stream, err
	}
}

// RecoveryInterceptor 恢复拦截器
type RecoveryInterceptor struct {
	logger *zap.Logger
}

// NewRecoveryInterceptor 创建新的恢复拦截器
// logger: 日志记录器
// 返回: RecoveryInterceptor 实例
func NewRecoveryInterceptor(logger *zap.Logger) *RecoveryInterceptor {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecoveryInterceptor{logger: logger}
}

// UnaryServerInterceptor 一元服务端恢复拦截器
// 返回: gRPC 一元服务端拦截器
func (ri *RecoveryInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				ri.logger.Error("gRPC panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		
		return handler(ctx, req)
	}
}

// StreamServerInterceptor 流式服务端恢复拦截器
// 返回: gRPC 流式服务端拦截器
func (ri *RecoveryInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				ri.logger.Error("gRPC stream panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		
		return handler(srv, stream)
	}
}

// getRequestID 从元数据中获取请求ID
// md: 元数据
// 返回: 请求ID
func getRequestID(md metadata.MD) string {
	if values := md.Get("request-id"); len(values) > 0 {
		return values[0]
	}
	return "unknown"
}

// addRequestIDToContext 向上下文添加请求ID（使用UUID）
// ctx: 上下文
// 返回: 包含请求ID的上下文
func addRequestIDToContext(ctx context.Context) context.Context {
	// 使用 vgo-kit/logger 的 GenerateRequestID 生成UUID
	requestID := vgologger.GenerateRequestID()
	md := metadata.Pairs("request-id", requestID)
	return metadata.NewOutgoingContext(ctx, md)
}