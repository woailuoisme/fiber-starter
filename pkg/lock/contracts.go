package lock

import (
	"context"
	"time"
)

// Locker 定义了分布式锁管理器的接口规范
type Locker interface {
	// Acquire 尝试抢占分布式锁，如果获取失败或超时，返回错误
	Acquire(ctx context.Context, key string, ttl time.Duration, options ...Option) (Lock, error)
}

// Lock 代表一个已成功获取的分布式锁实例
type Lock interface {
	// Release 原子释放该分布式锁
	Release(ctx context.Context) error

	// Key 返回该锁的唯一标识键
	Key() string
}
