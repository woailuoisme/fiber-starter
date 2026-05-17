package notification

import (
	"fmt"
	"sync"

	mailContracts "fiber-starter/app/Providers/Mail/Contracts"
	channels "fiber-starter/app/Providers/Notification/Channels"
	contracts "fiber-starter/app/Providers/Notification/Contracts"
)

// Manager handles sending notifications through various channels.
type Manager struct {
	channels map[string]contracts.Channel
	mu       sync.RWMutex
}

// NewNotificationManager creates a new Manager instance.
func NewNotificationManager(mailer mailContracts.Mailer) *Manager {
	m := &Manager{
		channels: make(map[string]contracts.Channel),
	}

	// Register default channels
	m.channels["mail"] = channels.NewMailChannel(mailer)

	return m
}

// Send sends the given notification to the given notifiables.
func (m *Manager) Send(notifiables interface{}, notification contracts.Notification) error {
	ns, ok := notifiables.([]interface{})
	if !ok {
		ns = []interface{}{notifiables}
	}

	for _, n := range ns {
		for _, channelName := range notification.Via(n) {
			channel, ok := m.channels[channelName]
			if !ok {
				return fmt.Errorf("unsupported notification channel: %s", channelName)
			}

			if err := channel.Send(n, notification); err != nil {
				return err
			}
		}
	}

	return nil
}

// Extend allows registering custom notification channels.
func (m *Manager) Extend(name string, channel contracts.Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[name] = channel
}
