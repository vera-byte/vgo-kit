package vgokit

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/vera-byte/vgo-kit/cache"
	"github.com/vera-byte/vgo-kit/config"
	"github.com/vera-byte/vgo-kit/metrics"
	"github.com/vera-byte/vgo-kit/ratelimit"
	"github.com/vera-byte/vgo-kit/sentry"
)

var (
	Log         *sentry.Logger
	Cfg         *viper.Viper
	Metrics     metrics.MetricsCollector
	Cache       cache.Cache
	RateLimiter ratelimit.RateLimiter
)

// init 初始化vgokit包的全局变量
// 加载配置文件并初始化日志服务
func init() {
	// 在测试环境中跳过初始化
	if isTestEnvironment() {
		return
	}
	
	v, err := config.LoadConfig("config/config.yaml")
	Cfg = v
	if err != nil {
		panic(err)
	}

	// 构建sentry配置，分别解析log和sentry部分
	var cfg *sentry.Config
	if v.IsSet("log") || v.IsSet("sentry") {
		cfg = &sentry.Config{}
		
		// 解析log配置
		if v.IsSet("log") {
			if unmarshalErr := v.UnmarshalKey("log", &cfg.Log); unmarshalErr != nil {
				panic(unmarshalErr)
			}
		}
		
		// 解析sentry配置
		if v.IsSet("sentry") {
			if unmarshalErr := v.UnmarshalKey("sentry", &cfg.Sentry); unmarshalErr != nil {
				panic(unmarshalErr)
			}
		}
	}

	// 初始化日志器，如果cfg为nil则使用默认配置
	log, err := sentry.InitLogger(cfg)
	if err != nil {
		panic(err)
	}
	Log = log
}

// isTestEnvironment 检查是否在测试环境中
func isTestEnvironment() bool {
	// 检查是否有测试相关的命令行参数
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") {
			return true
		}
	}
	return false
}
