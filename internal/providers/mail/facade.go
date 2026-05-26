package mail

import (
	"lfiber/internal/providers/mail/contracts"
	"lfiber/internal/support/appctx"
)

// mailer returns the default mailer instance from the container.
func mailer() contracts.Mailer {
	if app := appctx.App(); app != nil {
		return app.EmailServiceValue()
	}
	return nil
}

// manager returns the mail manager instance from the container.
func manager() contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.MailManagerValue()
	}
	return nil
}

// Drive returns a specific mailer instance
func Drive(name ...string) contracts.Mailer {
	if m := manager(); m != nil {
		return m.Drive(name...)
	}
	return nil
}

// To creates a new message with the given recipient
func To(to ...string) contracts.Message {
	if m := mailer(); m != nil {
		return m.To(to...)
	}
	return nil
}

// Send sends a message immediately
func Send(m contracts.Message) error {
	if mail := mailer(); mail != nil {
		return mail.Send(m)
	}
	return nil
}

// Raw sends a raw email
func Raw(to, subject, body string) error {
	if m := mailer(); m != nil {
		return m.Raw(to, subject, body)
	}
	return nil
}

// Close closes the default mailer connection
func Close() error {
	if m := manager(); m != nil {
		return m.Close()
	}
	return nil
}
