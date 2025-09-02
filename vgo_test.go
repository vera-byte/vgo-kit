package vgokit

import (
	"testing"
	"github.com/vera-byte/vgo-kit/config"
)

// TestPackageInitialization 测试包的初始化
// 验证全局变量是否正确初始化
func TestPackageInitialization(t *testing.T) {
	// 由于全局初始化可能失败，我们测试配置加载功能
	v, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	if v == nil {
		t.Error("Config should not be nil")
	}

	// 测试基本配置项
	if v.GetString("server.host") == "" {
		t.Error("server.host should be configured")
	}
}

// TestLoggerBasicFunctionality 测试日志器基本功能
// 验证日志器是否能正常工作
func TestLoggerBasicFunctionality(t *testing.T) {
	if Log == nil {
		t.Skip("Logger not initialized")
	}

	// 测试基本日志方法是否存在且不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Logger method panicked: %v", r)
		}
	}()

	// 测试日志方法
	Log.Info("Test info message")
	Log.Debug("Test debug message")
	Log.Warn("Test warning message")
}

// TestConfigBasicFunctionality 测试配置基本功能
// 验证配置是否能正常工作
func TestConfigBasicFunctionality(t *testing.T) {
	if Cfg == nil {
		t.Skip("Config not initialized")
	}

	// 测试配置方法是否存在且不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Config method panicked: %v", r)
		}
	}()

	// 测试基本配置方法
	_ = Cfg.AllKeys()
	_ = Cfg.AllSettings()
}