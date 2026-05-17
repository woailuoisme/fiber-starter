package notification

import (
	"fiber-starter/internal/providers/notification/Contracts"
	"fiber-starter/internal/support/appctx"
)

// dispatcher returns the notification dispatcher instance from the container.
func dispatcher() Contracts.Dispatcher {
	if app := appctx.App(); app != nil {
		return app.NotificationService()
	}
	return nil
}

// Send sends the given notification to the given notifiable entities.
func Send(notifiables interface{}, n Contracts.Notification) error {
	if d := dispatcher(); d != nil {
		return d.Send(notifiables, n)
	}
	return nil
}
