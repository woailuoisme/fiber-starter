package support

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"fiber-starter/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	Close() error
}

// redisCache Redis缓存实现
type redisCache struct {
	client *redis.Client
	prefix string
}

// NewCacheService 创建缓存服务实例
func NewCacheService(cfg *config.Config) CacheService {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Lazy connection: no Ping check here
	// Connection will be established automatically when commands are actually called
	Logger.Info("Redis service configured (lazy connection)",
		zap.String("host", cfg.Redis.Host),
		zap.String("port", cfg.Redis.Port),
		zap.Int("db", cfg.Redis.DB))

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
			Logger.Error("Failed to serialize cache value", zap.String("key", key), zap.Error(err))
			return err
		}
		val = string(jsonBytes)
	}

	err := c.client.Set(ctx, cacheKey, val, expiration).Err()
	if err != nil {
		Logger.Error("Failed to set cache",
			zap.String("key", cacheKey),
			zap.Duration("expiration", expiration),
			zap.Error(err))
		return err
	}

	Logger.Debug("Cache set successfully",
		zap.String("key", cacheKey),
		zap.Duration("expiration", expiration),
		zap.Int("size", len(val)))
	return nil
}

func (c *redisCache) Close() error {
	return c.client.Close()
}

// Get 获取缓存（字符串）
func (c *redisCache) Get(key string) (string, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	val, err := c.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			Logger.Debug("Cache key does not exist", zap.String("key", cacheKey))
			return "", redis.Nil
		}
		Logger.Error("Failed to get cache", zap.String("key", cacheKey), zap.Error(err))
		return "", err
	}

	Logger.Debug("Cache get successful", zap.String("key", cacheKey), zap.Int("size", len(val)))
	return val, nil
}

// GetBytes Get cache (byte array)
func (c *redisCache) GetBytes(key string) ([]byte, error) {
	val, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

// GetJSON Get cache and deserialize to target object
func (c *redisCache) GetJSON(key string, dest interface{}) error {
	val, err := c.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		Logger.Error("Failed to deserialize cache value", zap.String("key", key), zap.Error(err))
		return err
	}

	Logger.Debug("Cache deserialization successful", zap.String("key", key))
	return nil
}

// Delete cache
func (c *redisCache) Delete(key string) error {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	err := c.client.Del(ctx, cacheKey).Err()
	if err != nil {
		Logger.Error("Failed to delete cache", zap.String("key", cacheKey), zap.Error(err))
		return err
	}

	Logger.Debug("Cache delete successful", zap.String("key", cacheKey))
	return nil
}

// DeletePattern Delete cache by pattern
func (c *redisCache) DeletePattern(pattern string) error {
	ctx := context.Background()
	cachePattern := c.buildKey(pattern)

	keys, err := c.client.Keys(ctx, cachePattern).Result()
	if err != nil {
		Logger.Error("Failed to get matching keys", zap.String("pattern", cachePattern), zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			Logger.Error("Failed to batch delete cache",
				zap.String("pattern", cachePattern),
				zap.Int("count", len(keys)),
				zap.Error(err))
			return err
		}
		Logger.Debug("Batch cache delete successful",
			zap.String("pattern", cachePattern),
			zap.Int("count", len(keys)))
	} else {
		Logger.Debug("Batch cache delete, no matching keys found", zap.String("pattern", cachePattern))
	}

	return nil
}

// Exists Check if cache key exists
func (c *redisCache) Exists(key string) (bool, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	count, err := c.client.Exists(ctx, cacheKey).Result()
	if err != nil {
		Logger.Error("Failed to check cache key existence", zap.String("key", cacheKey), zap.Error(err))
		return false, err
	}

	exists := count > 0
	Logger.Debug("Check cache key existence", zap.String("key", cacheKey), zap.Bool("exists", exists))
	return exists, nil
}

// TTL Get remaining TTL of cache key
func (c *redisCache) TTL(key string) (time.Duration, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	duration, err := c.client.TTL(ctx, cacheKey).Result()
	if err != nil {
		Logger.Error("Failed to get cache TTL", zap.String("key", cacheKey), zap.Error(err))
		return 0, err
	}

	Logger.Debug("Get cache TTL", zap.String("key", cacheKey), zap.Duration("ttl", duration))
	return duration, nil
}

// Expire Set expiration time for cache key
func (c *redisCache) Expire(key string, expiration time.Duration) error {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	err := c.client.Expire(ctx, cacheKey, expiration).Err()
	if err != nil {
		Logger.Error("Failed to set cache expiration time",
			zap.String("key", cacheKey),
			zap.Duration("expiration", expiration),
			zap.Error(err))
		return err
	}

	Logger.Debug("Set cache expiration time successful",
		zap.String("key", cacheKey),
		zap.Duration("expiration", expiration))
	return nil
}

// Increment cache value
func (c *redisCache) Increment(key string) (int64, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	result, err := c.client.Incr(ctx, cacheKey).Result()
	if err != nil {
		Logger.Error("Failed to increment cache value", zap.String("key", cacheKey), zap.Error(err))
		return 0, err
	}

	Logger.Debug("Increment cache value successful", zap.String("key", cacheKey), zap.Int64("value", result))
	return result, nil
}

// Decrement cache value
func (c *redisCache) Decrement(key string) (int64, error) {
	ctx := context.Background()
	cacheKey := c.buildKey(key)

	result, err := c.client.Decr(ctx, cacheKey).Result()
	if err != nil {
		Logger.Error("Failed to decrement cache value", zap.String("key", cacheKey), zap.Error(err))
		return 0, err
	}

	Logger.Debug("Decrement cache value successful", zap.String("key", cacheKey), zap.Int64("value", result))
	return result, nil
}
