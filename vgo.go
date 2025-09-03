package vgokit

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/vera-byte/vgo-kit/cache"
	"github.com/vera-byte/vgo-kit/config"
	"github.com/vera-byte/vgo-kit/logger"
	"github.com/vera-byte/vgo-kit/metrics"
	"github.com/vera-byte/vgo-kit/ratelimit"
	"github.com/vera-byte/vgo-kit/sentry"
)

var (
	Log         logger.Logger
	SentryClient sentry.Client
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

	// 初始化日志系统
	var logConfig *logger.BaseConfig
	if v.IsSet("log") {
		logConfig = &logger.BaseConfig{}
		if unmarshalErr := v.UnmarshalKey("log", logConfig); unmarshalErr != nil {
			panic(unmarshalErr)
		}
	}

	// 创建日志实例
	logFactory := logger.NewZapFactory()
	loggerInstance, err := logFactory.CreateLogger(logConfig)
	if err != nil {
		panic(err)
	}
	Log = loggerInstance

	// 初始化Sentry客户端
	var sentryConfig *sentry.SentryConfig
	if v.IsSet("sentry") {
		sentryConfig = &sentry.SentryConfig{}
		if unmarshalErr := v.UnmarshalKey("sentry", sentryConfig); unmarshalErr != nil {
			panic(unmarshalErr)
		}
	}

	// 创建Sentry客户端
	sentryClient, err := sentry.NewClient(sentryConfig)
	if err != nil {
		panic(err)
	}
	SentryClient = sentryClient
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
