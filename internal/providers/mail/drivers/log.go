package drivers

import (
	"fiber-starter/internal/providers/mail/contracts"
	helpers "fiber-starter/internal/support"

	"go.uber.org/zap"
)

type LogDriver struct{}

func NewLogDriver() *LogDriver {
	return &LogDriver{}
}

func (d *LogDriver) To(to ...string) contracts.Message {
	return &BaseMessage{ToArr: to}
}

func (d *LogDriver) Send(m contracts.Message) error {
	helpers.Info("Email sent (logged)",
		zap.Strings("to", m.GetTo()),
		zap.Strings("cc", m.GetCc()),
		zap.Strings("bcc", m.GetBcc()),
		zap.String("subject", m.GetSubject()),
		zap.Bool("isHTML", m.IsHtml()),
		zap.Int("attachments", len(m.GetAttachments())))
	return nil
}

func (d *LogDriver) Raw(to, subject, body string) error {
	helpers.Info("Raw Email sent (logged)",
		zap.String("to", to),
		zap.String("subject", subject))
	return nil
}

func (d *LogDriver) HealthCheck() error {
	return nil
}

func (d *LogDriver) Close() error {
	return nil
}
