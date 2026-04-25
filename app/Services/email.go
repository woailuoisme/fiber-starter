package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/resend/resend-go/v3"
	"go.uber.org/zap"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendWelcomeEmail(to, name string) error
	SendPasswordResetEmail(to, resetToken string) error
	SendVerificationEmail(to, verificationToken string) error
	SendEmail(to, subject, body string, isHTML bool) error
}

type emailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{config: cfg}
}

func (s *emailService) SendEmail(to, subject, body string, isHTML bool) error {
	if strings.TrimSpace(s.config.Mail.APIKey) == "" {
		return errors.New("resend api key is not configured")
	}

	params := newEmailRequest(s.config.Mail.FromName, s.config.Mail.FromAddress, to, subject)
	if isHTML {
		params.Html = body
	} else {
		params.Text = body
	}
	if strings.TrimSpace(s.config.Mail.ReplyTo) != "" {
		params.ReplyTo = s.config.Mail.ReplyTo
	}

	client := resend.NewClient(s.config.Mail.APIKey)
	if _, err := client.Emails.SendWithContext(context.Background(), params); err != nil {
		helpers.LogError("Failed to send email", zap.Error(err), zap.String("to", to), zap.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *emailService) SendWelcomeEmail(to, name string) error {
	return s.sendTemplateEmail(to, "Welcome to our platform", fmt.Sprintf(`
		<h2>Welcome, %s!</h2>
		<p>Thank you for registering on our platform. We are excited to have you join us!</p>
		<p>If you have any questions, please feel free to contact our customer service team.</p>
		<p>Enjoy using our platform!</p>
		<p>Team Name</p>
	`, name))
}

func (s *emailService) SendPasswordResetEmail(to, resetToken string) error {
	return s.sendTemplateEmail(to, "Password Reset Request", fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>You have requested to reset your password. Please click the link below to reset your password:</p>
		<p><a href="%s/reset-password?token=%s">Reset Password</a></p>
		<p>If you did not request a password reset, please ignore this email.</p>
		<p>This link will expire in 24 hours.</p>
		<p>Team Name</p>
	`, s.config.App.URL, resetToken))
}

func (s *emailService) SendVerificationEmail(to, verificationToken string) error {
	return s.sendTemplateEmail(to, "Email Verification", fmt.Sprintf(`
		<h2>Email Verification</h2>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="%s/verify-email?token=%s">Verify Email</a></p>
		<p>If you did not register an account, please ignore this email.</p>
		<p>This link will expire in 1 hour.</p>
		<p>Team Name</p>
	`, s.config.App.URL, verificationToken))
}

func formatSender(name, address string) string {
	if strings.TrimSpace(name) == "" {
		return address
	}
	return (&mail.Address{Name: name, Address: address}).String()
}

func newEmailRequest(fromName, fromAddress, to, subject string) *resend.SendEmailRequest {
	return &resend.SendEmailRequest{
		From:    formatSender(fromName, fromAddress),
		To:      []string{to},
		Subject: subject,
	}
}

func (s *emailService) sendTemplateEmail(to, subject, body string) error {
	return s.SendEmail(to, subject, body, true)
}
