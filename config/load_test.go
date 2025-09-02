package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig 测试配置加载功能
// 验证配置文件加载是否正常工作
func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// 写入测试配置内容
	configContent := `
server:
  host: localhost
  port: 8080

log:
  level: info
  format: json
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 测试加载配置
	v, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if v == nil {
		t.Fatal("LoadConfig returned nil viper instance")
	}

	// 验证配置内容
	if v.GetString("server.host") != "localhost" {
		t.Errorf("Expected server.host to be 'localhost', got '%s'", v.GetString("server.host"))
	}

	if v.GetInt("server.port") != 8080 {
		t.Errorf("Expected server.port to be 8080, got %d", v.GetInt("server.port"))
	}

	if v.GetString("log.level") != "info" {
		t.Errorf("Expected log.level to be 'info', got '%s'", v.GetString("log.level"))
	}
}

// TestLoadConfigNonExistentFile 测试加载不存在的配置文件
// 验证错误处理是否正确
func TestLoadConfigNonExistentFile(t *testing.T) {
	// 尝试加载不存在的配置文件
	_, err := LoadConfig("/non/existent/path/config.yaml")
	if err == nil {
		t.Error("Expected error when loading non-existent config file, got nil")
	}
}

// TestLoadConfigWithEnvironmentVariables 测试环境变量配置
// 验证环境变量是否能正确覆盖配置文件
func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	// 设置环境变量 (viper使用下划线分隔的大写格式)
	os.Setenv("VGO_SERVER_HOST", "example.com")
	defer os.Unsetenv("VGO_SERVER_HOST")

	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
server:
  host: localhost
  port: 8080
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// 加载配置
	v, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 验证环境变量是否覆盖了配置文件
	if v.GetString("server.host") != "example.com" {
		t.Errorf("Expected server.host to be 'example.com' (from env var), got '%s'", v.GetString("server.host"))
	}
}
