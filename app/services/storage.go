package services

import (
	"fmt"
	"log"
	"time"

	"fiber-starter/config"

	"github.com/gofiber/storage"
	"github.com/gofiber/storage/bbolt"
	"github.com/gofiber/storage/memory"
	redisStorage "github.com/gofiber/storage/redis/v3"
)

// StorageService 存储服务
type StorageService struct {
	storage storage.Storage
	config  *config.StorageConfig
}

// NewStorageService 创建新的存储服务实例
func NewStorageService(cfg *config.StorageConfig, redisCfg *config.RedisConfig) (*StorageService, error) {
	var store storage.Storage

	switch cfg.Driver {
	case "redis":
		// 使用Redis存储
		// 如果密码为空，则不需要密码部分
		var url string
		if redisCfg.Password == "" {
			url = fmt.Sprintf("redis://%s:%s/%d",
				redisCfg.Host, redisCfg.Port, redisCfg.DB)
		} else {
			url = fmt.Sprintf("redis://:%s@%s:%s/%d",
				redisCfg.Password, redisCfg.Host, redisCfg.Port, redisCfg.DB)
		}
		store = redisStorage.New(redisStorage.Config{
			URL:   url,
			Reset: cfg.Reset,
		})

	case "bbolt":
		// 使用BBolt存储
		store = bbolt.New(bbolt.Config{
			Database: cfg.Database,
			Bucket:   "fiber",
			Reset:    false, // 不自动重置
		})

	case "memory":
		// 使用内存存储（默认）
		store = memory.New(memory.Config{})

	default:
		// 默认使用内存存储
		log.Printf("未知的存储驱动 '%s'，使用内存存储作为默认", cfg.Driver)
		store = memory.New(memory.Config{})
	}

	log.Printf("存储服务已初始化，驱动类型: %s", cfg.Driver)

	return &StorageService{
		storage: store,
		config:  cfg,
	}, nil
}

// Get 获取存储值
func (s *StorageService) Get(key string) ([]byte, error) {
	return s.storage.Get(key)
}

// Set 设置存储值
func (s *StorageService) Set(key string, value []byte, ttl time.Duration) error {
	return s.storage.Set(key, value, ttl)
}

// Delete 删除存储值
func (s *StorageService) Delete(key string) error {
	return s.storage.Delete(key)
}

// Reset 重置存储
func (s *StorageService) Reset() error {
	return s.storage.Reset()
}

// Close 关闭存储连接
func (s *StorageService) Close() error {
	if closer, ok := s.storage.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// GetStorage 获取底层存储实例
func (s *StorageService) GetStorage() storage.Storage {
	return s.storage
}

// SetWithDefaultTTL 使用默认TTL设置存储值
func (s *StorageService) SetWithDefaultTTL(key string, value []byte) error {
	// 默认TTL为1小时
	return s.Set(key, value, time.Hour)
}

// GetString 获取字符串值
func (s *StorageService) GetString(key string) (string, error) {
	val, err := s.Get(key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// SetString 设置字符串值
func (s *StorageService) SetString(key, value string, ttl time.Duration) error {
	return s.Set(key, []byte(value), ttl)
}

// SetStringWithDefaultTTL 使用默认TTL设置字符串值
func (s *StorageService) SetStringWithDefaultTTL(key, value string) error {
	return s.SetWithDefaultTTL(key, []byte(value))
}

// Exists 检查键是否存在
func (s *StorageService) Exists(key string) (bool, error) {
	val, err := s.Get(key)
	if err != nil {
		// 检查是否是"未找到"错误
		if err.Error() == "key not found" || err.Error() == "not found" {
			return false, nil
		}
		return false, err
	}
	return val != nil, nil
}

// SetExpire 设置键的过期时间
func (s *StorageService) SetExpire(key string, ttl time.Duration) error {
	// 获取当前值
	val, err := s.Get(key)
	if err != nil {
		return err
	}
	// 重新设置带过期时间的值
	return s.Set(key, val, ttl)
}
