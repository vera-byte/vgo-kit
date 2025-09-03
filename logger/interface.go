package logger

import (
	"go.uber.org/zap/zapcore"
)

// Logger 定义日志接口，用于解耦具体的日志实现
// 提供统一的日志记录方法，支持结构化日志
type Logger interface {
	// Debug 记录调试级别日志
	// msg: 日志消息
	// fields: 结构化字段
	Debug(msg string, fields ...zapcore.Field)

	// Info 记录信息级别日志
	// msg: 日志消息
	// fields: 结构化字段
	Info(msg string, fields ...zapcore.Field)

	// Warn 记录警告级别日志
	// msg: 日志消息
	// fields: 结构化字段
	Warn(msg string, fields ...zapcore.Field)

	// Error 记录错误级别日志
	// msg: 日志消息
	// fields: 结构化字段
	// 返回: error 用于链式调用
	Error(msg string, fields ...zapcore.Field) error

	// Fatal 记录致命错误级别日志并退出程序
	// msg: 日志消息
	// fields: 结构化字段
	Fatal(msg string, fields ...zapcore.Field)

	// WithFields 创建带有预设字段的新日志实例
	// fields: 预设的结构化字段
	// 返回: Logger 新的日志实例
	WithFields(fields ...zapcore.Field) Logger

	// WithRequestID 创建带有请求ID的新日志实例
	// reqID: 请求ID
	// 返回: Logger 新的日志实例
	WithRequestID(reqID string) Logger

	// Sync 同步日志缓冲区
	// 返回: error 同步过程中的错误
	Sync() error

	// Close 关闭日志系统，释放资源
	// 返回: error 关闭过程中的错误
	Close() error
}

// Config 日志配置接口
// 定义日志系统的基本配置结构
type Config interface {
	// GetLevel 获取日志级别
	// 返回: string 日志级别字符串
	GetLevel() string

	// GetDirectory 获取日志文件目录
	// 返回: string 日志文件目录路径
	GetDirectory() string

	// GetFilename 获取日志文件名
	// 返回: string 日志文件名
	GetFilename() string

	// IsStdoutEnabled 是否启用标准输出
	// 返回: bool 是否输出到标准输出
	IsStdoutEnabled() bool

	// GetMaxSize 获取单个日志文件最大大小(MB)
	// 返回: int 最大文件大小
	GetMaxSize() int

	// GetMaxBackups 获取保留的旧日志文件最大数量
	// 返回: int 最大备份数量
	GetMaxBackups() int

	// GetMaxAge 获取保留旧日志文件的最大天数
	// 返回: int 最大保留天数
	GetMaxAge() int

	// IsCompressEnabled 是否启用日志文件压缩
	// 返回: bool 是否压缩旧日志文件
	IsCompressEnabled() bool
}

// Factory 日志工厂接口
// 用于创建不同类型的日志实例
type Factory interface {
	// CreateLogger 创建日志实例
	// cfg: 日志配置
	// 返回: Logger 日志实例, error 创建过程中的错误
	CreateLogger(cfg Config) (Logger, error)
}