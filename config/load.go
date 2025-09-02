package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()

	// 1. 设置配置类型和文件名
	v.SetConfigType("yaml")
	v.SetConfigName(filepath.Base(configPath)) // 不带扩展名的文件名
	v.AddConfigPath(filepath.Dir(configPath))  // 配置文件所在目录

	// 2. 自动读取环境变量（可选）
	v.AutomaticEnv()
	v.SetEnvPrefix("VGO") // 环境变量前缀 IAM_SERVER_HOST

	// 3. 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return v, nil
}
