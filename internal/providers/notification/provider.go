package notification

import (
	"fiber-starter/configs"
	mailContracts "fiber-starter/internal/providers/mail/contracts"
	channels "fiber-starter/internal/providers/notification/channels"
	notificationContracts "fiber-starter/internal/providers/notification/contracts"
)

// RegisterNotification initializes the Notification manager.
func RegisterNotification(mailer mailContracts.Mailer) (*Manager, notificationContracts.Dispatcher, error) {
	manager := NewNotificationManager(mailer)
	return manager, manager, nil
}

// RegisterConfiguredChannels registers optional notification channels from config.
func RegisterConfiguredChannels(cfg *configs.Config, manager *Manager) error {
	if cfg == nil || manager == nil {
		return nil
	}

	if channel, err := channels.NewGotifyChannel(cfg.Notification.Gotify); err != nil {
		return err
	} else if channel != nil {
		manager.Extend("gotify", channel)
	}

	if channel, err := channels.NewTelegramChannel(cfg.Notification.Telegram); err != nil {
		return err
	} else if channel != nil {
		manager.Extend("telegram", channel)
	}

	return nil
}
