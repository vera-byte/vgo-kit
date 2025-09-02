package sentry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 封装了 zap.Logger 和 sentry 集成功能
type Logger struct {
	*zap.Logger
	hub *sentry.Hub
}

// InitLogger 初始化日志系统，支持 Sentry 集成
func InitLogger(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = DefaultSentryConfig()
	}
	var cores []zapcore.Core

	// 日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 输出到文件
	if cfg.Log.Directory != "" && cfg.Log.Filename != "" {
		if err := os.MkdirAll(cfg.Log.Directory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		logfile := filepath.Join(cfg.Log.Directory, cfg.Log.Filename)

		// 使用 lumberjack 进行日志轮转
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    getMaxSize(cfg.Log.MaxSize),
			MaxBackups: getMaxBackups(cfg.Log.MaxBackups),
			MaxAge:     getMaxAge(cfg.Log.MaxAge),
			Compress:   cfg.Log.Compress,
		})

		// 文件使用 JSON encoder，配置自定义时间格式
		fileConfig := zap.NewProductionEncoderConfig()
		fileConfig.TimeKey = "ts"
		fileConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02:15:04:05:000")
		fileEncoder := zapcore.NewJSONEncoder(fileConfig)
		cores = append(cores, zapcore.NewCore(fileEncoder, fileWriter, level))
	}

	// 输出到终端
	if cfg.Log.Stdout {
		// 终端使用 Console encoder
		consoleConfig := zap.NewDevelopmentEncoderConfig()
		consoleConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level))
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no log output configured")
	}

	// 初始化 Sentry（如果启用）
	var hub *sentry.Hub
	if cfg.Sentry.Enabled && cfg.Sentry.DSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Sentry.DSN,
			Environment:      cfg.Sentry.Environment,
			Debug:            cfg.Sentry.Debug,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			return nil, fmt.Errorf("sentry initialization failed: %w", err)
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())
		hub = sentry.CurrentHub()
	}

	// 创建 zap logger
	zapLogger := zap.New(zapcore.NewTee(cores...))

	return &Logger{
		Logger: zapLogger,
		hub:    hub,
	}, nil
}

// WithFields 添加字段到日志
func (l *Logger) WithFields(fields ...zapcore.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		hub:    l.hub,
	}
}

// WithRequestID 添加请求ID到日志
func (l *Logger) WithRequestID(reqID string) *Logger {
	if reqID == "" {
		reqID = GenerateRequestID()
	}
	return l.WithFields(zap.String("request_id", reqID))
}

// Error 记录错误并发送到 Sentry（如果启用） 返回error
func (l *Logger) Error(msg string, fields ...zapcore.Field) error {
	l.Logger.Error(msg, fields...)

	if l.hub != nil {
		// 提取错误信息
		var errorObj error
		var extraFields = make(map[string]interface{})

		for _, field := range fields {
			if field.Type == zapcore.ErrorType {
				if err, ok := field.Interface.(error); ok {
					errorObj = err
				}
			} else {
				extraFields[field.Key] = field.Interface
			}
		}

		// 发送到 Sentry
		event := sentry.NewEvent()
		event.Level = sentry.LevelError
		event.Message = msg
		event.Extra = extraFields

		if errorObj != nil {
			event.Exception = []sentry.Exception{
				{
					Type:  fmt.Sprintf("%T", errorObj),
					Value: errorObj.Error(),
				},
			}
		}

		l.hub.CaptureEvent(event)
	}
	return errors.New(msg + fmt.Sprintf("%v", fields))
}

// Fatal 记录致命错误并发送到 Sentry（如果启用）
func (l *Logger) Fatal(msg string, fields ...zapcore.Field) {
	l.Logger.Fatal(msg, fields...)

	if l.hub != nil {
		// 提取错误信息
		var errorObj error
		var extraFields = make(map[string]interface{})

		for _, field := range fields {
			if field.Type == zapcore.ErrorType {
				if err, ok := field.Interface.(error); ok {
					errorObj = err
				}
			} else {
				extraFields[field.Key] = field.Interface
			}
		}

		// 发送到 Sentry
		event := sentry.NewEvent()
		event.Level = sentry.LevelFatal
		event.Message = msg
		event.Extra = extraFields

		if errorObj != nil {
			event.Exception = []sentry.Exception{
				{
					Type:  fmt.Sprintf("%T", errorObj),
					Value: errorObj.Error(),
				},
			}
		}

		l.hub.CaptureEvent(event)
		l.hub.Flush(2 * time.Second)
	}
}

// Warn 记录警告信息
func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	l.Logger.Warn(msg, fields...)

	if l.hub != nil {
		// 对于警告，也发送到 Sentry
		event := sentry.NewEvent()
		event.Level = sentry.LevelWarning
		event.Message = msg

		extraFields := make(map[string]interface{})
		for _, field := range fields {
			extraFields[field.Key] = field.Interface
		}
		event.Extra = extraFields

		l.hub.CaptureEvent(event)
	}
}

// Info 记录信息
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.Logger.Info(msg, fields...)
}

// Debug 记录调试信息
func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.Logger.Debug(msg, fields...)
}

// Sync 同步日志
func (l *Logger) Sync() error {
	if l.hub != nil {
		l.hub.Flush(2 * time.Second)
	}
	return l.Logger.Sync()
}

// Close 清理资源
func (l *Logger) Close() error {
	if l.hub != nil {
		l.hub.Flush(2 * time.Second)
	}
	return l.Sync()
}

// GenerateRequestID 生成请求ID
func GenerateRequestID() string {
	return fmt.Sprintf("req-%s-%d", uuid.New().String()[:8], time.Now().UnixNano()%1000000)
}

// 辅助函数：获取最大文件大小，使用默认值
func getMaxSize(size int) int {
	if size <= 0 {
		return 100 // 默认 100MB
	}
	return size
}

// 辅助函数：获取最大备份数量，使用默认值
func getMaxBackups(backups int) int {
	if backups <= 0 {
		return 10 // 默认 10个
	}
	return backups
}

// 辅助函数：获取最大保留天数，使用默认值
func getMaxAge(age int) int {
	if age <= 0 {
		return 30 // 默认 30天
	}
	return age
}
