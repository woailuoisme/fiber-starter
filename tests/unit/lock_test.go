package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"lfiber/pkg/lock"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skip("Local Redis (127.0.0.1:6379) is not available; skipping lock integration tests")
	}

	return client
}

func TestRedisLocker_AcquireAndRelease(t *testing.T) {
	client := prepareRedis(t)
	defer client.Close()

	locker := lock.NewRedisLocker(client)
	ctx := context.Background()
	lockKey := "test:lock:acquire_release"

	// 清理脏数据
	_ = client.Del(ctx, lockKey).Err()

	// 1. 获取锁成功
	l1, err := locker.Acquire(ctx, lockKey, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, lockKey, l1.Key())

	// 2. 互斥冲突：第二个人在锁未释放且不自旋时，抢占失败
	_, err = locker.Acquire(ctx, lockKey, 2*time.Second)
	require.ErrorIs(t, err, lock.ErrLockAcquireTimeout)

	// 3. 第一个人释放锁
	err = l1.Release(ctx)
	require.NoError(t, err)

	// 4. 再次获取锁成功
	l2, err := locker.Acquire(ctx, lockKey, 2*time.Second)
	require.NoError(t, err)
	require.NoError(t, l2.Release(ctx))
}

func TestRedisLocker_SpinRetryAndTimeout(t *testing.T) {
	client := prepareRedis(t)
	defer client.Close()

	locker := lock.NewRedisLocker(client)
	ctx := context.Background()
	lockKey := "test:lock:spin_retry"

	_ = client.Del(ctx, lockKey).Err()

	// 1. 协程 A 抢占锁，锁定时间 5s
	l1, err := locker.Acquire(ctx, lockKey, 5*time.Second)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	var errB error
	var l2 lock.Lock

	// 2. 协程 B 尝试以 300ms 总超时自旋抢占这把锁，预期超时失败
	go func() {
		defer wg.Done()
		l2, errB = locker.Acquire(
			ctx,
			lockKey,
			2*time.Second,
			lock.WithTimeout(300*time.Millisecond),
			lock.WithRetryInterval(50*time.Millisecond),
		)
	}()

	wg.Wait()
	require.ErrorIs(t, errB, lock.ErrLockAcquireTimeout)
	assert.Nil(t, l2)

	// 3. 释放锁
	require.NoError(t, l1.Release(ctx))
}

func TestRedisLocker_SafetyRelease(t *testing.T) {
	client := prepareRedis(t)
	defer client.Close()

	locker := lock.NewRedisLocker(client)
	ctx := context.Background()
	lockKey := "test:lock:safety_release"

	_ = client.Del(ctx, lockKey).Err()

	// 1. 抢占成功
	l1, err := locker.Acquire(ctx, lockKey, 500*time.Millisecond)
	require.NoError(t, err)

	// 2. 强行模拟锁因为网络抖动/处理超时导致在 Redis 中过期
	_ = client.Del(ctx, lockKey).Err()

	// 此时被其他人写入了相同 key 且不同的 value
	_ = client.Set(ctx, lockKey, "other-owner-uuid", 5*time.Second).Err()

	// 3. A 尝试去释放原来属于他的锁，预期应该报错（释放失败，防止安全泄露）
	err = l1.Release(ctx)
	require.ErrorIs(t, err, lock.ErrLockReleaseFailed)

	// 清理
	_ = client.Del(ctx, lockKey).Err()
}
