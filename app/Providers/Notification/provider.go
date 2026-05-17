package notification

import (
	mailContracts "fiber-starter/app/Providers/Mail/Contracts"
	channels "fiber-starter/app/Providers/Notification/Channels"
	notificationContracts "fiber-starter/app/Providers/Notification/Contracts"
	"fiber-starter/configs"
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
