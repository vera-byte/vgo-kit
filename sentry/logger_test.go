package sentry

import (
	"testing"
)

// TestInitLoggerWithNil 测试传入nil配置时的日志初始化
func TestInitLoggerWithNil(t *testing.T) {
	logger, err := InitLogger(nil)
	if err != nil {
		t.Fatalf("Failed to init logger with nil config: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// 测试日志输出
	logger.Info("Test log message")
	t.Log("Logger initialized successfully with nil config")
}

// TestInitLoggerWithDefaultConfig 测试使用默认配置的日志初始化
func TestInitLoggerWithDefaultConfig(t *testing.T) {
	cfg := DefaultSentryConfig()
	logger, err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to init logger with default config: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// 测试日志输出
	logger.Info("Test log message with default config")
	t.Log("Logger initialized successfully with default config")
}

// TestDefaultSentryConfig 测试默认配置
func TestDefaultSentryConfig(t *testing.T) {
	cfg := DefaultSentryConfig()
	if cfg == nil {
		t.Fatal("Default config should not be nil")
	}

	if cfg.Log.Directory == "" {
		t.Error("Log directory should not be empty")
	}

	if cfg.Log.Filename == "" {
		t.Error("Log filename should not be empty")
	}

	if !cfg.Log.Stdout {
		t.Error("Stdout should be enabled by default")
	}

	t.Logf("Default config: %+v", cfg)
}
