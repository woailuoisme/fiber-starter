package tests

import (
	"testing"

	Services "fiber-starter/app/Services"
	"fiber-starter/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailService_RequiresResendAPIKey(t *testing.T) {
	cfg := &config.Config{}
	cfg.Mail.FromAddress = "noreply@example.com"
	cfg.Mail.FromName = "Fiber Starter"

	email := Services.NewEmailService(cfg)
	err := email.SendEmail("user@example.com", "Subject", "Body", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resend api key")
}
