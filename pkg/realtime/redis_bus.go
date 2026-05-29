package realtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisSubscription struct {
	pubsub *redis.PubSub
	ch     chan []byte
	closed chan struct{}
	once   sync.Once
}

func (s *redisSubscription) Messages() <-chan []byte {
	return s.ch
}

func (s *redisSubscription) Close() error {
	var err error
	s.once.Do(func() {
		close(s.closed)
		if s.pubsub != nil {
			err = s.pubsub.Close()
		}
	})
	return err
}

type redisBus struct {
	client *redis.Client
	prefix string
	logger Logger
}

func newRedisBus(client *redis.Client, prefix string, logger Logger) EventBus {
	if logger == nil {
		logger = NewNoopLogger()
	}
	return &redisBus{client: client, prefix: prefix, logger: logger}
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
	ch := make(chan []byte, 128)
	closed := make(chan struct{})

	sub := &redisSubscription{
		ch:     ch,
		closed: closed,
	}

	go b.subscribeWithRetry(ctx, channel, sub)

	return sub, nil
}

func (b *redisBus) subscribeWithRetry(ctx context.Context, channel string, sub *redisSubscription) {
	defer close(sub.ch)

	backoff := 100 * time.Millisecond
	maxBackoff := 8 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-sub.closed:
			return
		default:
		}

		subCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pubsub := b.client.Subscribe(subCtx, b.key(channel))
		cancel()

		// 验证连接是否确实成功
		_, err := pubsub.Receive(ctx)
		if err != nil {
			b.logger.Warn("realtime_redis_subscribe_failed_retrying", "channel", channel, "error", err.Error(), "backoff", backoff.String())
			_ = pubsub.Close()

			select {
			case <-ctx.Done():
				return
			case <-sub.closed:
				return
			case <-time.After(backoff):
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
		}

		// 订阅连接成功后重置退避时间
		backoff = 100 * time.Millisecond
		sub.pubsub = pubsub

		msgCh := pubsub.Channel()
		b.logger.Info("realtime_redis_subscribed", "channel", channel)

		// 内部消费循环
		shouldRetry := false
		for {
			select {
			case <-ctx.Done():
				return
			case <-sub.closed:
				return
			case msg, ok := <-msgCh:
				if !ok {
					// 消息管道断开，说明 Redis 连接可能挂了，需要重试订阅
					shouldRetry = true
					break
				}
				if msg == nil {
					continue
				}
				select {
				case sub.ch <- []byte(msg.Payload):
				case <-ctx.Done():
					return
				case <-sub.closed:
					return
				}
			}
			if shouldRetry {
				break
			}
		}

		if shouldRetry {
			b.logger.Warn("realtime_redis_connection_lost_re_subscribing", "channel", channel)
			_ = pubsub.Close()
		}
	}
}

func (b *redisBus) Close() error {
	return nil
}
