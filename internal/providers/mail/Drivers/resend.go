package Drivers

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/mail/Contracts"
	helpers "fiber-starter/internal/support"

	"github.com/resend/resend-go/v3"
	"go.uber.org/zap"
)

type ResendDriver struct {
	client *resend.Client
	config *configs.Config
}

func NewResendDriver(cfg *configs.Config) *ResendDriver {
	if strings.TrimSpace(cfg.Mail.APIKey) == "" {
		helpers.Warn("Resend API key is not configured, emails will not be sent")
		return &ResendDriver{config: cfg}
	}

	client := resend.NewClient(cfg.Mail.APIKey)
	return &ResendDriver{
		client: client,
		config: cfg,
	}
}

func (d *ResendDriver) To(to ...string) Contracts.Message {
	return &BaseMessage{ToArr: to}
}

func (d *ResendDriver) Send(m Contracts.Message) error {
	if d.client == nil {
		return errors.New("resend client not initialized")
	}

	params := &resend.SendEmailRequest{
		From:    d.formatSender(d.config.Mail.FromName, d.config.Mail.FromAddress),
		To:      m.GetTo(),
		Cc:      m.GetCc(),
		Bcc:     m.GetBcc(),
		Subject: m.GetSubject(),
	}

	if m.IsHtml() {
		params.Html = m.GetBody()
	} else {
		params.Text = m.GetBody()
	}

	if strings.TrimSpace(d.config.Mail.ReplyTo) != "" {
		params.ReplyTo = d.config.Mail.ReplyTo
	}

	if _, err := d.client.Emails.SendWithContext(context.Background(), params); err != nil {
		helpers.LogError("Failed to send email via Resend",
			zap.Error(err),
			zap.Strings("to", m.GetTo()),
			zap.String("subject", m.GetSubject()))
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (d *ResendDriver) Raw(to, subject, body string) error {
	m := d.To(to).Subject(subject).Plain(body)
	return d.Send(m)
}

func (d *ResendDriver) formatSender(name, address string) string {
	if strings.TrimSpace(name) == "" {
		return address
	}
	return (&mail.Address{Name: name, Address: address}).String()
}

func (d *ResendDriver) HealthCheck() error {
	if d.client == nil {
		return errors.New("resend client not initialized")
	}
	// Note: We don't perform a live API check to avoid rate limits
	return nil
}

func (d *ResendDriver) Close() error {
	return nil
}
