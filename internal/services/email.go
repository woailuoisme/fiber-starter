package services

import (
	"crypto/tls"
	"fmt"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendWelcomeEmail(to, name string) error
	SendPasswordResetEmail(to, resetToken string) error
	SendVerificationEmail(to, verificationToken string) error
	SendEmail(to, subject, body string, isHTML bool) error
}

// emailService 邮件服务实现
type emailService struct {
	config *config.Config
}

// NewEmailService 创建邮件服务实例
func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{
		config: cfg,
	}
}

// SendEmail 发送邮件
func (s *emailService) SendEmail(to, subject, body string, isHTML bool) error {
	m := mail.NewMsg()

	// Set sender
	if err := m.FromFormat(s.config.Mail.FromName, s.config.Mail.FromAddress); err != nil {
		helpers.LogError("Failed to set sender", zap.Error(err), zap.String("to", to))
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := m.To(to); err != nil {
		helpers.LogError("Failed to set recipient", zap.Error(err), zap.String("to", to))
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Set subject
	m.Subject(subject)

	// Set email content
	contentType := mail.TypeTextPlain
	if isHTML {
		contentType = mail.TypeTextHTML
	}
	m.SetBodyString(contentType, body)

	// Configure client options
	options := []mail.Option{
		mail.WithPort(s.config.Mail.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.config.Mail.Username),
		mail.WithPassword(s.config.Mail.Password),
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: s.config.Mail.TLSInsecure, //nolint:gosec // Configuration driven TLS skip verify
	}
	options = append(options, mail.WithTLSConfig(tlsConfig))

	// Set based on encryption type
	switch s.config.Mail.Encryption {
	case "ssl":
		// SSL usually means implicit TLS (port 465)
		options = append(options, mail.WithSSL())
	case "tls":
		// TLS usually means STARTTLS (port 587)
		options = append(options, mail.WithTLSPolicy(mail.TLSMandatory))
	default:
		options = append(options, mail.WithTLSPolicy(mail.NoTLS))
	}

	// Create SMTP client
	c, err := mail.NewClient(s.config.Mail.Host, options...)
	if err != nil {
		helpers.LogError("Failed to create email client", zap.Error(err), zap.String("to", to))
		return fmt.Errorf("failed to create email client: %w", err)
	}

	// Send email
	if err := c.DialAndSend(m); err != nil {
		helpers.LogError("Failed to send email", zap.Error(err), zap.String("to", to), zap.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendWelcomeEmail Send welcome email
func (s *emailService) SendWelcomeEmail(to, name string) error {
	subject := "Welcome to our platform"
	body := fmt.Sprintf(`
		<h2>Welcome, %s!</h2>
		<p>Thank you for registering on our platform. We are excited to have you join us!</p>
		<p>If you have any questions, please feel free to contact our customer service team.</p>
		<p>Enjoy using our platform!</p>
		<p>Team Name</p>
	`, name)

	return s.SendEmail(to, subject, body, true)
}

// SendPasswordResetEmail Send password reset email
func (s *emailService) SendPasswordResetEmail(to, resetToken string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>You have requested to reset your password. Please click the link below to reset your password:</p>
		<p><a href="%s/reset-password?token=%s">Reset Password</a></p>
		<p>If you did not request a password reset, please ignore this email.</p>
		<p>This link will expire in 24 hours.</p>
		<p>Team Name</p>
	`, fmt.Sprintf("http://%s:%s", s.config.App.Host, s.config.App.Port), resetToken)

	return s.SendEmail(to, subject, body, true)
}

// SendVerificationEmail Send email verification email
func (s *emailService) SendVerificationEmail(to, verificationToken string) error {
	subject := "Email Verification"
	body := fmt.Sprintf(`
		<h2>Email Verification</h2>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="%s/verify-email?token=%s">Verify Email</a></p>
		<p>If you did not register an account, please ignore this email.</p>
		<p>This link will expire in 1 hour.</p>
		<p>Team Name</p>
	`, fmt.Sprintf("http://%s:%s", s.config.App.Host, s.config.App.Port), verificationToken)

	return s.SendEmail(to, subject, body, true)
}
