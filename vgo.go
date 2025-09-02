package vgokit

import (
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
	// 初始化日志服务
	v, err := config.LoadConfig("config/config.yaml")
	Cfg = v
	if err != nil {
		panic(err)
	}

	// 反序列化到结构体，如果配置文件中不存在sentry配置则保持nil
	var cfg *sentry.Config
	if v.IsSet("sentry") || v.IsSet("log") {
		if unmarshalErr := v.UnmarshalKey("sentry", &cfg); unmarshalErr != nil {
			panic(unmarshalErr)
		}
	}

	// 初始化日志器，如果cfg为nil则使用默认配置
	log, err := sentry.InitLogger(cfg)
	if err != nil {
		panic(err)
	}
	Log = log
}
