package sentry

func DefaultSentryConfig() *Config {
	// 配置
	config := Config{
		Log: LogConfig{
			Level:      "info",
			Directory:  "./logs",
			Filename:   "vgo.log",
			Stdout:     true,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		Sentry: SentryConfig{
			Enabled:     false,
			DSN:         "your-sentry-dsn-here",
			Environment: "development",
			Debug:       false,
		},
	}

	return &config
}
