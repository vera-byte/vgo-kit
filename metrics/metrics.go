package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// gRPC相关指标
	RecordGRPCRequest(method string, err error)
	RecordGRPCDuration(method string, duration time.Duration)
	
	// 数据库相关指标
	UpdateDBConnections(active, idle, total int)
	
	// 业务指标
	RecordBusinessMetric(metricType string)
	
	// 错误指标
	RecordError(errorType, errorCode string)
	
	// 认证指标
	RecordAuthAttempt(success bool)
	
	// 获取gRPC拦截器
	GetGRPCInterceptor() grpc.UnaryServerInterceptor
	
	// 获取HTTP处理器
	GetHTTPHandler() http.Handler
}

// DefaultMetrics 默认指标收集器实现
type DefaultMetrics struct {
	// gRPC 请求计数器
	grpcRequestsTotal *prometheus.CounterVec
	// gRPC 请求持续时间
	grpcRequestDuration *prometheus.HistogramVec
	// 数据库连接池指标
	dbConnectionsActive prometheus.Gauge
	dbConnectionsIdle   prometheus.Gauge
	dbConnectionsTotal  prometheus.Gauge
	// 业务指标
	businessMetrics *prometheus.CounterVec
	// 错误指标
	errorsTotal *prometheus.CounterVec
	// 认证指标
	authAttemptsTotal *prometheus.CounterVec
	// Prometheus注册器
	registry *prometheus.Registry
}

// NewMetrics 创建新的指标收集器
func NewMetrics(namespace string) *DefaultMetrics {
	registry := prometheus.NewRegistry()
	
	m := &DefaultMetrics{
		registry: registry,
		grpcRequestsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "grpc_requests_total",
				Help:      "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		),
		grpcRequestDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "grpc_request_duration_seconds",
				Help:      "Duration of gRPC requests in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		dbConnectionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_active",
				Help:      "Number of active database connections",
			},
		),
		dbConnectionsIdle: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Number of idle database connections",
			},
		),
		dbConnectionsTotal: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_total",
				Help:      "Total number of database connections",
			},
		),
		businessMetrics: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "business_operations_total",
				Help:      "Total number of business operations",
			},
			[]string{"operation_type"},
		),
		errorsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors by type",
			},
			[]string{"error_type", "error_code"},
		),
		authAttemptsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_attempts_total",
				Help:      "Total number of authentication attempts",
			},
			[]string{"result"},
		),
	}
	
	// 注册默认的Go运行时指标
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	
	return m
}

// RecordGRPCRequest 记录gRPC请求
func (m *DefaultMetrics) RecordGRPCRequest(method string, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.grpcRequestsTotal.WithLabelValues(method, status).Inc()
}

// RecordGRPCDuration 记录gRPC请求持续时间
func (m *DefaultMetrics) RecordGRPCDuration(method string, duration time.Duration) {
	m.grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// UpdateDBConnections 更新数据库连接指标
func (m *DefaultMetrics) UpdateDBConnections(active, idle, total int) {
	m.dbConnectionsActive.Set(float64(active))
	m.dbConnectionsIdle.Set(float64(idle))
	m.dbConnectionsTotal.Set(float64(total))
}

// RecordBusinessMetric 记录业务指标
func (m *DefaultMetrics) RecordBusinessMetric(metricType string) {
	m.businessMetrics.WithLabelValues(metricType).Inc()
}

// RecordError 记录错误
func (m *DefaultMetrics) RecordError(errorType, errorCode string) {
	m.errorsTotal.WithLabelValues(errorType, errorCode).Inc()
}

// RecordAuthAttempt 记录认证尝试
func (m *DefaultMetrics) RecordAuthAttempt(success bool) {
	result := "failure"
	if success {
		result = "success"
	}
	m.authAttemptsTotal.WithLabelValues(result).Inc()
}

// GetGRPCInterceptor 获取gRPC指标拦截器
func (m *DefaultMetrics) GetGRPCInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		resp, err := handler(ctx, req)
		
		duration := time.Since(start)
		m.RecordGRPCDuration(info.FullMethod, duration)
		m.RecordGRPCRequest(info.FullMethod, err)
		
		// 记录错误详情
		if err != nil {
			st := status.Convert(err)
			m.RecordError("grpc", st.Code().String())
		}
		
		return resp, err
	}
}

// GetHTTPHandler 获取Prometheus HTTP处理器
func (m *DefaultMetrics) GetHTTPHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Registry: m.registry,
	})
}

// 全局指标收集器实例
var globalMetrics MetricsCollector

// InitGlobalMetrics 初始化全局指标收集器
func InitGlobalMetrics(namespace string) {
	globalMetrics = NewMetrics(namespace)
}

// GetGlobalMetrics 获取全局指标收集器
func GetGlobalMetrics() MetricsCollector {
	if globalMetrics == nil {
		InitGlobalMetrics("vgo")
	}
	return globalMetrics
}

// 便捷函数

// RecordBusinessMetric 记录业务指标的便捷函数
func RecordBusinessMetric(metricType string) {
	GetGlobalMetrics().RecordBusinessMetric(metricType)
}

// RecordBusinessError 记录业务错误的便捷函数
func RecordBusinessError(errorType, errorCode string) {
	GetGlobalMetrics().RecordError(errorType, errorCode)
}

// RecordAuth 记录认证结果的便捷函数
func RecordAuth(success bool) {
	GetGlobalMetrics().RecordAuthAttempt(success)
}

// UpdateDBStats 更新数据库统计信息的便捷函数
func UpdateDBStats(active, idle, total int) {
	GetGlobalMetrics().UpdateDBConnections(active, idle, total)
}