package sentry

// DefaultSentryConfig 返回默认的 Sentry 配置
// 返回: *SentryConfig 默认配置实例
func DefaultSentryConfig() *SentryConfig {
	return &SentryConfig{
		Enabled:     false,
		DSN:         "",
		Environment: "development",
		Debug:       false,
		SampleRate:  1.0,
	}
}
