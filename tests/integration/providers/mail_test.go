package providers_test

import (
	"testing"

	"lfiber/configs"
	providers "lfiber/internal/providers"
	mail "lfiber/internal/providers/mail"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailProvider_ResendErrorWithoutAPIKey(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = true
	cfg.Mail.FromAddress = "noreply@example.com"
	cfg.Mail.FromName = "lfiber"
	cfg.Mail.APIKey = "" // Explicitly empty

	_, mailer, err := mail.Register(cfg)
	require.NoError(t, err)

	err = mailer.Send(mailer.To("user@example.com").Subject("Subject").Plain("Body"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend client not initialized")
}

func TestMailProvider_LogDriver(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = true

	manager := mail.NewManager(cfg)
	runtime := &providers.Runtime{
		MailManager:  manager,
		EmailService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	logMailer := mail.Drive("log")

	msg := logMailer.To("user@example.com").
		Cc("cc@example.com").
		Bcc("bcc@example.com").
		Subject("Complex Subject").
		Plain("Body").
		Data(map[string]interface{}{"key": "value"})

	err := logMailer.Send(msg)
	require.NoError(t, err)
}

func TestMailProvider_Facade(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = true
	manager := mail.NewManager(cfg)
	runtime := &providers.Runtime{
		MailManager:  manager,
		EmailService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	// Test default (resend) through facade - should return error not panic
	err := mail.Send(mail.To("user@example.com").Subject("Subject").Plain("Body"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend client not initialized")
}

func TestMailProvider_SMTPDriver(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = true
	cfg.Mail.Host = "localhost"
	cfg.Mail.Port = 12345 // Random port that likely has no SMTP server

	manager := mail.NewManager(cfg)
	runtime := &providers.Runtime{
		MailManager:  manager,
		EmailService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	smtpMailer := mail.Drive("smtp")

	// Sending should fail since no server is running on port 12345
	err := smtpMailer.Send(smtpMailer.To("user@example.com").Subject("Test").Plain("Body"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send smtp email")
}

func TestMailProvider_Templates(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = true
	manager := mail.NewManager(cfg)
	runtime := &providers.Runtime{
		MailManager:  manager,
		EmailService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	// Use log driver for template testing to avoid network calls
	mailer := mail.Drive("log")

	assert.NoError(t, mailer.Send(mailer.To("test@example.com").Subject("Welcome").Html("Welcome John")))
	assert.NoError(t, mailer.Send(mailer.To("test@example.com").Subject("Password Reset").Plain("token123")))
	assert.NoError(t, mailer.Send(mailer.To("test@example.com").Subject("Verify Email").Plain("token456")))
}

func TestMailProvider_Raw(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = true
	cfg.Mail.Host = "127.0.0.1"
	cfg.Mail.Port = 1

	manager := mail.NewManager(cfg)
	runtime := &providers.Runtime{
		MailManager:  manager,
		EmailService: manager.Drive(),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	logMailer := mail.Drive("log")
	require.NoError(t, logMailer.Raw("user@example.com", "Subject", "Body"))

	resendMailer := mail.Drive("resend")
	err := resendMailer.Raw("user@example.com", "Subject", "Body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend client not initialized")

	smtpMailer := mail.Drive("smtp")
	err = smtpMailer.Raw("user@example.com", "Subject", "Body")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send smtp email")
}
