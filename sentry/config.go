package sentry

// LogConfig 日志配置结构体
type LogConfig struct {
	Level      string `yaml:"level" json:"level"`             // 日志级别: debug, info, warn, error, fatal
	Directory  string `yaml:"directory" json:"directory"`     // 日志文件目录
	Filename   string `yaml:"filename" json:"filename"`       // 日志文件名
	Stdout     bool   `yaml:"stdout" json:"stdout"`           // 是否输出到终端
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // 保留的旧日志文件最大数量
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // 保留旧日志文件的最大天数
	Compress   bool   `yaml:"compress" json:"compress"`       // 是否压缩旧日志文件
}

// SentryConfig Sentry 配置结构体
type SentryConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`         // 是否启用 Sentry
	DSN         string `yaml:"dsn" json:"dsn"`                 // Sentry DSN
	Environment string `yaml:"environment" json:"environment"` // 环境名称
	Debug       bool   `yaml:"debug" json:"debug"`             // 是否开启调试模式
}

// Config 完整的日志和 Sentry 配置
type Config struct {
	Log    LogConfig    `yaml:"log" json:"log"`
	Sentry SentryConfig `yaml:"sentry" json:"sentry"`
}
