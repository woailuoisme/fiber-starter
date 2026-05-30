package providers_test

import (
	"testing"

	"lfiber/configs"
	"lfiber/internal/providers/asynqredis"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsynqRedis_ClientOptSetsDialerRetries(t *testing.T) {
	cfg := &configs.Config{
		Redis: configs.RedisConfig{
			Host:     "127.0.0.1",
			Port:     "6379",
			Password: "secret",
			DB:       3,
		},
	}

	client, ok := asynqredis.NewClientOpt(cfg, 2).MakeRedisClient().(*redis.Client)
	require.True(t, ok)
	t.Cleanup(func() {
		assert.NoError(t, client.Close())
	})

	opts := client.Options()
	assert.Equal(t, "127.0.0.1:6379", opts.Addr)
	assert.Equal(t, "secret", opts.Password)
	assert.Equal(t, 5, opts.DB)
	assert.Equal(t, 2, opts.DialerRetries)
}
