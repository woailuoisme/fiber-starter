package channels

import (
	"errors"

	mailContracts "lfiber/internal/providers/mail/contracts"
	contracts "lfiber/internal/providers/notification/contracts"
)

// MailChannel sends notifications via email.
type MailChannel struct {
	mailer mailContracts.Mailer
}

// NewMailChannel creates a new MailChannel instance.
func NewMailChannel(mailer mailContracts.Mailer) *MailChannel {
	return &MailChannel{mailer: mailer}
}

// Send sends the given notification.
func (c *MailChannel) Send(notifiable interface{}, notification contracts.Notification) error {
	mailNotification, ok := notification.(contracts.MailNotification)
	if !ok {
		return nil // Skip if not a mail notification
	}

	message := mailNotification.ToMail(notifiable)
	if message == nil {
		return nil
	}

	// Here we would extract the email address from the notifiable.
	// For now, we'll assume the notifiable has a method RouteNotificationForMail().
	address := ""
	if router, ok := notifiable.(interface{ RouteNotificationForMail() string }); ok {
		address = router.RouteNotificationForMail()
	}

	if address == "" {
		return errors.New("could not determine email address for notifiable")
	}

	// Send the mail using the mailer.
	msg := c.mailer.To(address)
	// If the notification provided a subject or body, we could set them here.
	if m, ok := message.(mailContracts.Message); ok {
		return c.mailer.Send(m)
	}

	return c.mailer.Send(msg)
}
