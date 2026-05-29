package lock

import (
	"fmt"
	"strings"

	"lfiber/configs"

	"github.com/redis/go-redis/v9"
)

// Register 初始化并返回通用的分布式锁管理器
func Register(cfg *configs.Config) (Locker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	addr := strings.TrimSpace(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port))
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return NewRedisLocker(client), nil
}
