package lock

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrLockAcquireTimeout = errors.New("lock acquisition timed out")
	ErrLockReleaseFailed  = errors.New("failed to release lock: owner mismatch or expired")
)

// luaReleaseScript 是 Redis 官方标准原子的分布式锁释放脚本
var luaReleaseScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

type redisLocker struct {
	client *redis.Client
}

// NewRedisLocker 创建一个基于 Redis 连接的 Locker 管理器
func NewRedisLocker(client *redis.Client) Locker {
	return &redisLocker{client: client}
}

func (l *redisLocker) Acquire(ctx context.Context, key string, ttl time.Duration, opts ...Option) (Lock, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	value := uuid.NewString()
	startTime := time.Now()

	for {
		// SET key value NX PX ttl
		ok, err := l.client.SetNX(ctx, key, value, ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			return &redisLock{
				client: l.client,
				key:    key,
				value:  value,
			}, nil
		}

		// 若未设定超时，抢占不到锁时立即退出报错
		if o.acquireTimeout <= 0 {
			return nil, ErrLockAcquireTimeout
		}

		// 判定是否超过设定的抢占最大超时时间
		if time.Since(startTime) >= o.acquireTimeout {
			return nil, ErrLockAcquireTimeout
		}

		// 自旋退避休眠，等待下一次重试
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(o.retryInterval):
		}
	}
}

type redisLock struct {
	client *redis.Client
	key    string
	value  string
}

func (r *redisLock) Key() string {
	return r.key
}

func (r *redisLock) Release(ctx context.Context) error {
	res, err := luaReleaseScript.Run(ctx, r.client, []string{r.key}, r.value).Result()
	if err != nil {
		return err
	}

	status, ok := res.(int64)
	if !ok || status != 1 {
		return ErrLockReleaseFailed
	}

	return nil
}
