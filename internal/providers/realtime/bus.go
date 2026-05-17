package realtime

import "context"

// Subscription represents a live channel subscription on the event bus.
type Subscription interface {
	Messages() <-chan []byte
	Close() error
}

// EventBus transports broadcast envelopes across application instances.
type EventBus interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channel string) (Subscription, error)
	Close() error
}
