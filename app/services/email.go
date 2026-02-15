package services

import (
	"crypto/tls"
	"fmt"

	"fiber-starter/config"
	"github.com/wneessen/go-mail"
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

	// 设置发件人
	if err := m.FromFormat(s.config.Mail.FromName, s.config.Mail.FromAddress); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	if err := m.To(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	// 设置主题
	m.Subject(subject)

	// 设置邮件内容
	contentType := mail.TypeTextPlain
	if isHTML {
		contentType = mail.TypeTextHTML
	}
	m.SetBodyString(contentType, body)

	// 配置客户端选项
	options := []mail.Option{
		mail.WithPort(s.config.Mail.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.config.Mail.Username),
		mail.WithPassword(s.config.Mail.Password),
	}

	// 配置TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: s.config.Mail.TLSInsecure, //nolint:gosec // Configuration driven TLS skip verify
	}
	options = append(options, mail.WithTLSConfig(tlsConfig))

	// 根据加密类型设置
	switch s.config.Mail.Encryption {
	case "ssl":
		// SSL通常意味着隐式TLS (端口465)
		options = append(options, mail.WithSSL())
	case "tls":
		// TLS通常意味着STARTTLS (端口587)
		options = append(options, mail.WithTLSPolicy(mail.TLSMandatory))
	default:
		options = append(options, mail.WithTLSPolicy(mail.NoTLS))
	}

	// 创建SMTP客户端
	c, err := mail.NewClient(s.config.Mail.Host, options...)
	if err != nil {
		return fmt.Errorf("创建邮件客户端失败: %w", err)
	}

	// 发送邮件
	if err := c.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// SendWelcomeEmail 发送欢迎邮件
func (s *emailService) SendWelcomeEmail(to, name string) error {
	subject := "欢迎注册我们的平台"
	body := fmt.Sprintf(`
		<h2>欢迎，%s！</h2>
		<p>感谢您注册我们的平台。我们很高兴您加入我们！</p>
		<p>如果您有任何问题，请随时联系我们的客服团队。</p>
		<p>祝您使用愉快！</p>
		<p>团队名称</p>
	`, name)

	return s.SendEmail(to, subject, body, true)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *emailService) SendPasswordResetEmail(to, resetToken string) error {
	subject := "密码重置请求"
	body := fmt.Sprintf(`
		<h2>密码重置请求</h2>
		<p>您请求重置密码。请点击下面的链接重置您的密码：</p>
		<p><a href="%s/reset-password?token=%s">重置密码</a></p>
		<p>如果您没有请求重置密码，请忽略此邮件。</p>
		<p>此链接将在24小时后过期。</p>
		<p>团队名称</p>
	`, fmt.Sprintf("http://%s:%s", s.config.App.Host, s.config.App.Port), resetToken)

	return s.SendEmail(to, subject, body, true)
}

// SendVerificationEmail 发送邮箱验证邮件
func (s *emailService) SendVerificationEmail(to, verificationToken string) error {
	subject := "邮箱验证"
	body := fmt.Sprintf(`
		<h2>邮箱验证</h2>
		<p>请点击下面的链接验证您的邮箱地址：</p>
		<p><a href="%s/verify-email?token=%s">验证邮箱</a></p>
		<p>如果您没有注册账户，请忽略此邮件。</p>
		<p>此链接将在1小时后过期。</p>
		<p>团队名称</p>
	`, fmt.Sprintf("http://%s:%s", s.config.App.Host, s.config.App.Port), verificationToken)

	return s.SendEmail(to, subject, body, true)
}
