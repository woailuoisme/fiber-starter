package Drivers

import (
	"fmt"
	"net/smtp"
	"strings"

	"fiber-starter/app/Providers/Mail/Contracts"
	helpers "fiber-starter/app/Support"
	"fiber-starter/configs"

	"go.uber.org/zap"
)

type SMTPDriver struct {
	config *configs.Config
}

func NewSMTPDriver(cfg *configs.Config) *SMTPDriver {
	return &SMTPDriver{config: cfg}
}

func (d *SMTPDriver) To(to ...string) Contracts.Message {
	return &BaseMessage{ToArr: to}
}

func (d *SMTPDriver) Send(m Contracts.Message) error {
	addr := fmt.Sprintf("%s:%d", d.config.Mail.Host, d.config.Mail.Port)
	from := d.config.Mail.FromAddress

	header := make(map[string]string)
	header["From"] = d.formatSender(d.config.Mail.FromName, from)
	header["To"] = strings.Join(m.GetTo(), ", ")
	if len(m.GetCc()) > 0 {
		header["Cc"] = strings.Join(m.GetCc(), ", ")
	}
	header["Subject"] = m.GetSubject()

	if m.IsHtml() {
		header["MIME-Version"] = "1.0"
		header["Content-Type"] = "text/html; charset=\"utf-8\""
	} else {
		header["Content-Type"] = "text/plain; charset=\"utf-8\""
	}

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + m.GetBody()

	var auth smtp.Auth
	if d.config.Mail.Username != "" {
		auth = smtp.PlainAuth("", d.config.Mail.Username, d.config.Mail.Password, d.config.Mail.Host)
	}

	// Combine all recipients
	recipients := append(m.GetTo(), m.GetCc()...)
	recipients = append(recipients, m.GetBcc()...)

	err := smtp.SendMail(addr, auth, from, recipients, []byte(message))
	if err != nil {
		helpers.LogError("Failed to send email via SMTP",
			zap.Error(err),
			zap.String("addr", addr),
			zap.Strings("to", m.GetTo()))
		return fmt.Errorf("failed to send smtp email: %w", err)
	}

	helpers.Info("Email sent via SMTP", zap.Strings("to", m.GetTo()), zap.String("subject", m.GetSubject()))
	return nil
}

func (d *SMTPDriver) Raw(to, subject, body string) error {
	m := d.To(to).Subject(subject).Plain(body)
	return d.Send(m)
}

func (d *SMTPDriver) formatSender(name, address string) string {
	if strings.TrimSpace(name) == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", name, address)
}

func (d *SMTPDriver) HealthCheck() error {
	addr := fmt.Sprintf("%s:%d", d.config.Mail.Host, d.config.Mail.Port)
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	return nil
}

func (d *SMTPDriver) Close() error {
	return nil
}
