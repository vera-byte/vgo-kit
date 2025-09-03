package logger

// BaseConfig 基础日志配置实现
// 实现了Config接口，提供日志系统的基本配置
type BaseConfig struct {
	Level      string `yaml:"level" json:"level"`             // 日志级别: debug, info, warn, error, fatal
	Directory  string `yaml:"directory" json:"directory"`     // 日志文件目录
	Filename   string `yaml:"filename" json:"filename"`       // 日志文件名
	Stdout     bool   `yaml:"stdout" json:"stdout"`           // 是否输出到终端
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // 保留的旧日志文件最大数量
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // 保留旧日志文件的最大天数
	Compress   bool   `yaml:"compress" json:"compress"`       // 是否压缩旧日志文件
}

// GetLevel 获取日志级别
// 返回: string 日志级别字符串
func (c *BaseConfig) GetLevel() string {
	return c.Level
}

// GetDirectory 获取日志文件目录
// 返回: string 日志文件目录路径
func (c *BaseConfig) GetDirectory() string {
	return c.Directory
}

// GetFilename 获取日志文件名
// 返回: string 日志文件名
func (c *BaseConfig) GetFilename() string {
	return c.Filename
}

// IsStdoutEnabled 是否启用标准输出
// 返回: bool 是否输出到标准输出
func (c *BaseConfig) IsStdoutEnabled() bool {
	return c.Stdout
}

// GetMaxSize 获取单个日志文件最大大小(MB)
// 返回: int 最大文件大小
func (c *BaseConfig) GetMaxSize() int {
	return c.MaxSize
}

// GetMaxBackups 获取保留的旧日志文件最大数量
// 返回: int 最大备份数量
func (c *BaseConfig) GetMaxBackups() int {
	return c.MaxBackups
}

// GetMaxAge 获取保留旧日志文件的最大天数
// 返回: int 最大保留天数
func (c *BaseConfig) GetMaxAge() int {
	return c.MaxAge
}

// IsCompressEnabled 是否启用日志文件压缩
// 返回: bool 是否压缩旧日志文件
func (c *BaseConfig) IsCompressEnabled() bool {
	return c.Compress
}

// DefaultConfig 创建默认日志配置
// 返回: *BaseConfig 默认配置实例
func DefaultConfig() *BaseConfig {
	return &BaseConfig{
		Level:      "info",
		Directory:  "./logs",
		Filename:   "app.log",
		Stdout:     true,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
}