package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 速率限制器接口
type RateLimiter interface {
	// Allow 检查是否允许请求
	Allow(ctx context.Context, key string) (bool, error)
	// AllowN 检查是否允许N个请求
	AllowN(ctx context.Context, key string, n int) (bool, error)
	// Reset 重置指定key的限制
	Reset(ctx context.Context, key string) error
	// GetRemaining 获取剩余请求数
	GetRemaining(ctx context.Context, key string) (int, error)
}

// RedisRateLimiter Redis实现的速率限制器
type RedisRateLimiter struct {
	client   *redis.Client
	limit    int           // 限制数量
	window   time.Duration // 时间窗口
	prefix   string        // key前缀
}

// NewRedisRateLimiter 创建Redis速率限制器
func NewRedisRateLimiter(client *redis.Client, limit int, window time.Duration, prefix string) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		limit:  limit,
		window: window,
		prefix: prefix,
	}
}

// Allow 检查是否允许请求
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN 检查是否允许N个请求
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	fullKey := r.getKey(key)
	now := time.Now().Unix()
	windowStart := now - int64(r.window.Seconds())

	// 使用Lua脚本确保原子性
	luaScript := `
		local key = KEYS[1]
		local window_start = ARGV[1]
		local now = ARGV[2]
		local limit = tonumber(ARGV[3])
		local increment = tonumber(ARGV[4])
		local ttl = tonumber(ARGV[5])

		-- 清理过期的记录
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- 获取当前窗口内的请求数
		local current = redis.call('ZCARD', key)

		-- 检查是否超过限制
		if current + increment > limit then
			return {0, current}
		end

		-- 添加新的请求记录
		for i = 1, increment do
			redis.call('ZADD', key, now, now .. ':' .. i)
		end

		-- 设置过期时间
		redis.call('EXPIRE', key, ttl)

		return {1, current + increment}
	`

	result, err := r.client.Eval(ctx, luaScript, []string{fullKey},
		windowStart, now, r.limit, n, int(r.window.Seconds())+1).Result()
	if err != nil {
		return false, err
	}

	results := result.([]interface{})
	allowed := results[0].(int64) == 1

	return allowed, nil
}

// Reset 重置指定key的限制
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := r.getKey(key)
	return r.client.Del(ctx, fullKey).Err()
}

// GetRemaining 获取剩余请求数
func (r *RedisRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	fullKey := r.getKey(key)
	now := time.Now().Unix()
	windowStart := now - int64(r.window.Seconds())

	// 清理过期记录并获取当前计数
	luaScript := `
		local key = KEYS[1]
		local window_start = ARGV[1]
		local limit = tonumber(ARGV[2])

		-- 清理过期的记录
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- 获取当前窗口内的请求数
		local current = redis.call('ZCARD', key)

		return limit - current
	`

	result, err := r.client.Eval(ctx, luaScript, []string{fullKey}, windowStart, r.limit).Result()
	if err != nil {
		return 0, err
	}

	remaining := int(result.(int64))
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// getKey 获取完整的key
func (r *RedisRateLimiter) getKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// MemoryRateLimiter 内存实现的速率限制器（用于测试或单机部署）
type MemoryRateLimiter struct {
	limit    int
	window   time.Duration
	requests map[string][]time.Time
}

// NewMemoryRateLimiter 创建内存速率限制器
func NewMemoryRateLimiter(limit int, window time.Duration) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
	}
}

// Allow 检查是否允许请求
func (m *MemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return m.AllowN(ctx, key, 1)
}

// AllowN 检查是否允许N个请求
func (m *MemoryRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-m.window)

	// 清理过期请求
	if requests, exists := m.requests[key]; exists {
		var validRequests []time.Time
		for _, req := range requests {
			if req.After(windowStart) {
				validRequests = append(validRequests, req)
			}
		}
		m.requests[key] = validRequests
	}

	// 检查是否超过限制
	currentCount := len(m.requests[key])
	if currentCount+n > m.limit {
		return false, nil
	}

	// 添加新请求
	for i := 0; i < n; i++ {
		m.requests[key] = append(m.requests[key], now)
	}

	return true, nil
}

// Reset 重置指定key的限制
func (m *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	delete(m.requests, key)
	return nil
}

// GetRemaining 获取剩余请求数
func (m *MemoryRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	now := time.Now()
	windowStart := now.Add(-m.window)

	// 清理过期请求
	if requests, exists := m.requests[key]; exists {
		var validRequests []time.Time
		for _, req := range requests {
			if req.After(windowStart) {
				validRequests = append(validRequests, req)
			}
		}
		m.requests[key] = validRequests
		return m.limit - len(validRequests), nil
	}

	return m.limit, nil
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	Type       string        `yaml:"type" json:"type"` // "redis" or "memory"
	Limit      int           `yaml:"limit" json:"limit"`
	Window     time.Duration `yaml:"window" json:"window"`
	Prefix     string        `yaml:"prefix" json:"prefix"`
	RedisAddr  string        `yaml:"redis_addr" json:"redis_addr"`
	RedisDB    int           `yaml:"redis_db" json:"redis_db"`
	RedisPass  string        `yaml:"redis_pass" json:"redis_pass"`
}

// DefaultRateLimitConfig 默认速率限制配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled: true,
		Type:    "memory",
		Limit:   100,
		Window:  time.Minute,
		Prefix:  "ratelimit",
	}
}

// NewRateLimiter 根据配置创建速率限制器
func NewRateLimiter(config *RateLimitConfig) (RateLimiter, error) {
	if !config.Enabled {
		return &NoOpRateLimiter{}, nil
	}

	switch config.Type {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			DB:       config.RedisDB,
			Password: config.RedisPass,
		})
		return NewRedisRateLimiter(client, config.Limit, config.Window, config.Prefix), nil
	case "memory":
		return NewMemoryRateLimiter(config.Limit, config.Window), nil
	default:
		return nil, fmt.Errorf("unsupported rate limiter type: %s", config.Type)
	}
}

// NoOpRateLimiter 无操作速率限制器（禁用时使用）
type NoOpRateLimiter struct{}

func (noop *NoOpRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (noop *NoOpRateLimiter) AllowN(ctx context.Context, key string, count int) (bool, error) {
	return true, nil
}

func (noop *NoOpRateLimiter) Reset(ctx context.Context, key string) error {
	return nil
}

func (noop *NoOpRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	return 999999, nil
}