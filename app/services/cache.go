package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fiber-starter/config"

	"github.com/go-redis/redis/v8"
)

// CacheService 缓存服务接口
type CacheService interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	GetBytes(key string) ([]byte, error)
	GetJSON(key string, dest interface{}) error
	Delete(key string) error
	DeletePattern(pattern string) error
	Exists(key string) (bool, error)
	TTL(key string) (time.Duration, error)
	Expire(key string, expiration time.Duration) error
	Increment(key string) (int64, error)
	Decrement(key string) (int64, error)
}

// redisCache Redis缓存实现
type redisCache struct {
	client *redis.Client
	prefix string
}

// NewCacheService 创建缓存服务实例
func NewCacheService(cfg *config.Config) CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("Redis连接失败: %v\n", err)
	} else {
		fmt.Printf("Redis连接成功: %s:%s\n", cfg.Redis.Host, cfg.Redis.Port)
	}

	return &redisCache{
		client: rdb,
		prefix: cfg.Cache.Prefix,
	}
}

// buildKey 构建带前缀的缓存键
func (c *redisCache) buildKey(key string) string {
	return c.prefix + key
}

// Set 设置缓存
func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("序列化缓存值失败: %w", err)
		}
		val = string(jsonBytes)
	}

	return c.client.Set(ctx, cacheKey, val, expiration).Err()
}

// Get 获取缓存（字符串）
func (c *redisCache) Get(key string) (string, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	val, err := c.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("缓存键不存在")
		}
		return "", fmt.Errorf("获取缓存失败: %w", err)
	}

	return val, nil
}

// GetBytes 获取缓存（字节数组）
func (c *redisCache) GetBytes(key string) ([]byte, error) {
	val, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

// GetJSON 获取缓存并反序列化到目标对象
func (c *redisCache) GetJSON(key string, dest interface{}) error {
	val, err := c.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("反序列化缓存值失败: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (c *redisCache) Delete(key string) error {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	return c.client.Del(ctx, cacheKey).Err()
}

// DeletePattern 根据模式删除缓存
func (c *redisCache) DeletePattern(pattern string) error {
	ctx := context.Background()
	cachePattern := c.buildKey(pattern)

	keys, err := c.client.Keys(ctx, cachePattern).Result()
	if err != nil {
		return fmt.Errorf("获取匹配键失败: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists 检查缓存键是否存在
func (c *redisCache) Exists(key string) (bool, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	count, err := c.client.Exists(ctx, cacheKey).Result()
	if err != nil {
		return false, fmt.Errorf("检查缓存键存在性失败: %w", err)
	}

	return count > 0, nil
}

// TTL 获取缓存键的剩余生存时间
func (c *redisCache) TTL(key string) (time.Duration, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	duration, err := c.client.TTL(ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("获取缓存TTL失败: %w", err)
	}

	return duration, nil
}

// Expire 设置缓存键的过期时间
func (c *redisCache) Expire(key string, expiration time.Duration) error {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	return c.client.Expire(ctx, cacheKey, expiration).Err()
}

// Increment 递增缓存值
func (c *redisCache) Increment(key string) (int64, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	return c.client.Incr(ctx, cacheKey).Result()
}

// Decrement 递减缓存值
func (c *redisCache) Decrement(key string) (int64, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	return c.client.Decr(ctx, cacheKey).Result()
}

// Close 关闭Redis连接
func (c *redisCache) Close() error {
	return c.client.Close()
}