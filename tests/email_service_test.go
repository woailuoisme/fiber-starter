package tests

import (
	"strings"
	"testing"

	"fiber-starter/internal/config"
	"fiber-starter/internal/services"
)

func TestEmailService_RequiresResendAPIKey(t *testing.T) {
	cfg := &config.Config{}
	cfg.Mail.FromAddress = "noreply@example.com"
	cfg.Mail.FromName = "Fiber Starter"

	email := services.NewEmailService(cfg)
	err := email.SendEmail("user@example.com", "Subject", "Body", false)
	if err == nil {
		t.Fatal("SendEmail expected error when RESEND_API_KEY is not configured")
	}
	if !strings.Contains(err.Error(), "resend api key") {
		t.Fatalf("SendEmail error mismatch: got=%v", err)
	}
}
