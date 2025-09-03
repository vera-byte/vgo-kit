package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger zap日志实现
// 实现了Logger接口，基于zap提供高性能的结构化日志
type ZapLogger struct {
	*zap.Logger
}

// ZapFactory zap日志工厂
// 实现了Factory接口，用于创建ZapLogger实例
type ZapFactory struct{}

// CreateLogger 创建zap日志实例
// cfg: 日志配置
// 返回: Logger 日志实例, error 创建过程中的错误
func (f *ZapFactory) CreateLogger(cfg Config) (Logger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var cores []zapcore.Core

	// 日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.GetLevel())); err != nil {
		level = zapcore.InfoLevel
	}

	// 输出到文件
	if cfg.GetDirectory() != "" && cfg.GetFilename() != "" {
		if err := os.MkdirAll(cfg.GetDirectory(), 0750); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		logfile := filepath.Join(cfg.GetDirectory(), cfg.GetFilename())

		// 使用 lumberjack 进行日志轮转
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    getMaxSize(cfg.GetMaxSize()),
			MaxBackups: getMaxBackups(cfg.GetMaxBackups()),
			MaxAge:     getMaxAge(cfg.GetMaxAge()),
			Compress:   cfg.IsCompressEnabled(),
		})

		// 文件输出配置
		fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)
		cores = append(cores, fileCore)
	}

	// 输出到控制台
	if cfg.IsStdoutEnabled() {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 如果没有配置任何输出，默认输出到控制台
	if len(cores) == 0 {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 创建核心和日志器
	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{Logger: zapLogger}, nil
}

// Debug 记录调试级别日志
// msg: 日志消息
// fields: 结构化字段
func (l *ZapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info 记录信息级别日志
// msg: 日志消息
// fields: 结构化字段
func (l *ZapLogger) Info(msg string, fields ...zapcore.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn 记录警告级别日志
// msg: 日志消息
// fields: 结构化字段
func (l *ZapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error 记录错误级别日志
// msg: 日志消息
// fields: 结构化字段
// 返回: error 用于链式调用
func (l *ZapLogger) Error(msg string, fields ...zapcore.Field) error {
	l.Logger.Error(msg, fields...)
	return fmt.Errorf("%s", msg)
}

// Fatal 记录致命错误级别日志并退出程序
// msg: 日志消息
// fields: 结构化字段
func (l *ZapLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.Logger.Fatal(msg, fields...)
}

// WithFields 创建带有预设字段的新日志实例
// fields: 预设的结构化字段
// 返回: Logger 新的日志实例
func (l *ZapLogger) WithFields(fields ...zapcore.Field) Logger {
	return &ZapLogger{Logger: l.Logger.With(fields...)}
}

// WithRequestID 创建带有请求ID的新日志实例
// reqID: 请求ID
// 返回: Logger 新的日志实例
func (l *ZapLogger) WithRequestID(reqID string) Logger {
	return &ZapLogger{Logger: l.Logger.With(zap.String("request_id", reqID))}
}

// Sync 同步日志缓冲区
// 返回: error 同步过程中的错误
func (l *ZapLogger) Sync() error {
	return l.Logger.Sync()
}

// Close 关闭日志系统，释放资源
// 返回: error 关闭过程中的错误
func (l *ZapLogger) Close() error {
	return l.Logger.Sync()
}

// getMaxSize 获取最大文件大小，提供默认值
// size: 配置的大小
// 返回: int 实际使用的大小
func getMaxSize(size int) int {
	if size <= 0 {
		return 100 // 默认100MB
	}
	return size
}

// getMaxBackups 获取最大备份数量，提供默认值
// backups: 配置的备份数量
// 返回: int 实际使用的备份数量
func getMaxBackups(backups int) int {
	if backups <= 0 {
		return 3 // 默认保留3个备份
	}
	return backups
}

// getMaxAge 获取最大保留天数，提供默认值
// age: 配置的保留天数
// 返回: int 实际使用的保留天数
func getMaxAge(age int) int {
	if age <= 0 {
		return 28 // 默认保留28天
	}
	return age
}