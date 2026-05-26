package realtime

import (
	"context"
	"fmt"
	"strings"

	"lfiber/configs"

	"github.com/redis/go-redis/v9"
)

type redisSubscription struct {
	pubsub *redis.PubSub
	ch     chan []byte
}

func (s *redisSubscription) Messages() <-chan []byte {
	return s.ch
}

func (s *redisSubscription) Close() error {
	if s == nil || s.pubsub == nil {
		return nil
	}
	return s.pubsub.Close()
}

type redisBus struct {
	client *redis.Client
	prefix string
}

func newRedisBus(client *redis.Client, prefix string) EventBus {
	return &redisBus{client: client, prefix: prefix}
}

func (b *redisBus) key(channel string) string {
	if b.prefix == "" {
		return channel
	}
	return fmt.Sprintf("%s:%s", b.prefix, channel)
}

func (b *redisBus) Publish(ctx context.Context, channel string, payload []byte) error {
	return b.client.Publish(ctx, b.key(channel), payload).Err()
}

func (b *redisBus) Subscribe(ctx context.Context, channel string) (Subscription, error) {
	pubsub := b.client.Subscribe(ctx, b.key(channel))
	ch := make(chan []byte, 128)

	go func() {
		defer close(ch)
		msgCh := pubsub.Channel()
		for msg := range msgCh {
			if msg == nil {
				continue
			}
			ch <- []byte(msg.Payload)
		}
	}()

	return &redisSubscription{pubsub: pubsub, ch: ch}, nil
}

func (b *redisBus) Close() error {
	return nil
}

func newRedisClient(cfg *configs.Config) *redis.Client {
	addr := strings.TrimSpace(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port))
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}
