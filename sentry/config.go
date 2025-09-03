package sentry

// SentryConfig Sentry 配置结构体
// 定义Sentry错误监控服务的配置参数
type SentryConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`         // 是否启用 Sentry
	DSN         string `yaml:"dsn" json:"dsn"`                 // Sentry DSN
	Environment string `yaml:"environment" json:"environment"` // 环境名称
	Debug       bool   `yaml:"debug" json:"debug"`             // 是否开启调试模式
	SampleRate  float64 `yaml:"sample_rate" json:"sample_rate"` // 采样率 (0.0-1.0)
}

// GetEnabled 获取是否启用Sentry
// 返回: bool 是否启用
func (c *SentryConfig) GetEnabled() bool {
	return c.Enabled
}

// GetDSN 获取Sentry DSN
// 返回: string DSN字符串
func (c *SentryConfig) GetDSN() string {
	return c.DSN
}

// GetEnvironment 获取环境名称
// 返回: string 环境名称
func (c *SentryConfig) GetEnvironment() string {
	return c.Environment
}

// GetDebug 获取是否开启调试模式
// 返回: bool 是否开启调试
func (c *SentryConfig) GetDebug() bool {
	return c.Debug
}

// GetSampleRate 获取采样率
// 返回: float64 采样率
func (c *SentryConfig) GetSampleRate() float64 {
	if c.SampleRate <= 0 || c.SampleRate > 1 {
		return 1.0 // 默认采样率
	}
	return c.SampleRate
}
