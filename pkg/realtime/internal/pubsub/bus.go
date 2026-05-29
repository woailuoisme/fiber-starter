package pubsub

import "context"

// Subscription 接口表示一个活动通道在 EventBus 上的网络订阅
type Subscription interface {
	Messages() <-chan []byte
	Close() error
}

// EventBus 接口负责在不同的应用节点之间传输广播消息信封 (Envelope)
type EventBus interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channel string) (Subscription, error)
	Close() error
}
