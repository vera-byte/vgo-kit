package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheConfig Redis缓存配置
type CacheConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// Cache 缓存接口
type Cache interface {
	// 基础操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 字符串操作
	SetString(ctx context.Context, key, value string, expiration time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)

	// 哈希操作
	HSet(ctx context.Context, key string, values ...interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error

	// 列表操作
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)

	// 集合操作
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	// 有序集合操作
	ZAdd(ctx context.Context, key string, members ...redis.Z) error
	ZRem(ctx context.Context, key string, members ...interface{}) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)

	// 管道操作
	Pipeline() redis.Pipeliner

	// 事务操作
	TxPipeline() redis.Pipeliner

	// 连接管理
	Ping(ctx context.Context) error
	Close() error

	// 获取原始客户端
	GetClient() *redis.Client
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	config *CacheConfig
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(config *CacheConfig) (*RedisCache, error) {
	if config == nil {
		return nil, fmt.Errorf("cache config is required")
	}

	// 设置默认值
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 5
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		config: config,
	}, nil
}

// Set 设置缓存
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Del 删除缓存
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余生存时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// SetString 设置字符串值
func (r *RedisCache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// GetString 获取字符串值
func (r *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Incr 递增
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr 递减
func (r *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// HSet 设置哈希字段
func (r *RedisCache) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段值
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// LPush 从左侧推入列表
func (r *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 从右侧推入列表
func (r *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop 从左侧弹出列表元素
func (r *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出列表元素
func (r *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (r *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// SAdd 添加集合成员
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SRem 移除集合成员
func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// ZAdd 添加有序集合成员
func (r *RedisCache) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRem 移除有序集合成员
func (r *RedisCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (r *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的成员及分数
func (r *RedisCache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// Pipeline 获取管道
func (r *RedisCache) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline 获取事务管道
func (r *RedisCache) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// Ping 测试连接
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GetClient 获取原始Redis客户端
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

// 全局缓存实例
var globalCache Cache

// InitGlobalCache 初始化全局缓存
func InitGlobalCache(config *CacheConfig) error {
	cache, err := NewRedisCache(config)
	if err != nil {
		return err
	}
	globalCache = cache
	return nil
}

// GetGlobalCache 获取全局缓存实例
func GetGlobalCache() Cache {
	return globalCache
}

// 便捷函数

// Set 设置缓存的便捷函数
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if globalCache == nil {
		return fmt.Errorf("global cache not initialized")
	}
	return globalCache.Set(ctx, key, value, expiration)
}

// Get 获取缓存的便捷函数
func Get(ctx context.Context, key string, dest interface{}) error {
	if globalCache == nil {
		return fmt.Errorf("global cache not initialized")
	}
	return globalCache.Get(ctx, key, dest)
}

// Del 删除缓存的便捷函数
func Del(ctx context.Context, keys ...string) error {
	if globalCache == nil {
		return fmt.Errorf("global cache not initialized")
	}
	return globalCache.Del(ctx, keys...)
}

// SetString 设置字符串缓存的便捷函数
func SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	if globalCache == nil {
		return fmt.Errorf("global cache not initialized")
	}
	return globalCache.SetString(ctx, key, value, expiration)
}

// GetString 获取字符串缓存的便捷函数
func GetString(ctx context.Context, key string) (string, error) {
	if globalCache == nil {
		return "", fmt.Errorf("global cache not initialized")
	}
	return globalCache.GetString(ctx, key)
}

// NoOpCache 空操作缓存实现，用于测试或不需要缓存的场景
type NoOpCache struct{}

// NewNoOpCache 创建空操作缓存实例
func NewNoOpCache() Cache {
	return &NoOpCache{}
}

// Set 空操作
func (n *NoOpCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

// Get 空操作，总是返回未找到错误
func (n *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return fmt.Errorf("key not found")
}

// Del 空操作
func (n *NoOpCache) Del(ctx context.Context, keys ...string) error {
	return nil
}

// Exists 空操作，总是返回false
func (n *NoOpCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 0, nil
}

// Expire 空操作
func (n *NoOpCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

// TTL 空操作，总是返回-1
func (n *NoOpCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return -1, nil
}

// SetString 空操作
func (n *NoOpCache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return nil
}

// GetString 空操作，总是返回空字符串和未找到错误
func (n *NoOpCache) GetString(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("key not found")
}

// Incr 空操作，总是返回1
func (n *NoOpCache) Incr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

// Decr 空操作，总是返回-1
func (n *NoOpCache) Decr(ctx context.Context, key string) (int64, error) {
	return -1, nil
}

// HSet 空操作
func (n *NoOpCache) HSet(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

// HGet 空操作，总是返回空字符串和未找到错误
func (n *NoOpCache) HGet(ctx context.Context, key, field string) (string, error) {
	return "", fmt.Errorf("key not found")
}

// HGetAll 空操作，总是返回空map
func (n *NoOpCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return make(map[string]string), nil
}

// HDel 空操作
func (n *NoOpCache) HDel(ctx context.Context, key string, fields ...string) error {
	return nil
}

// LPush 空操作
func (n *NoOpCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

// RPush 空操作
func (n *NoOpCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

// LPop 空操作，总是返回空字符串和未找到错误
func (n *NoOpCache) LPop(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("key not found")
}

// RPop 空操作，总是返回空字符串和未找到错误
func (n *NoOpCache) RPop(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("key not found")
}

// LLen 空操作，总是返回0
func (n *NoOpCache) LLen(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

// SAdd 空操作
func (n *NoOpCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

// SRem 空操作
func (n *NoOpCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

// SMembers 空操作，总是返回空切片
func (n *NoOpCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil
}

// SIsMember 空操作，总是返回false
func (n *NoOpCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return false, nil
}

// ZAdd 空操作
func (n *NoOpCache) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return nil
}

// ZRem 空操作
func (n *NoOpCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

// ZRange 空操作，总是返回空切片
func (n *NoOpCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return []string{}, nil
}

// ZRangeWithScores 空操作，总是返回空切片
func (n *NoOpCache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return []redis.Z{}, nil
}

// Pipeline 空操作，返回nil
func (n *NoOpCache) Pipeline() redis.Pipeliner {
	return nil
}

// TxPipeline 空操作，返回nil
func (n *NoOpCache) TxPipeline() redis.Pipeliner {
	return nil
}

// Ping 空操作
func (n *NoOpCache) Ping(ctx context.Context) error {
	return nil
}

// Close 空操作
func (n *NoOpCache) Close() error {
	return nil
}

// GetClient 空操作，返回nil
func (n *NoOpCache) GetClient() *redis.Client {
	return nil
}
