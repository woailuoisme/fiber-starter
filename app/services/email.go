package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"fiber-starter/config"

	"github.com/jordan-wright/email"
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
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.config.Mail.FromName, s.config.Mail.FromAddress)
	e.To = []string{to}
	e.Subject = subject

	if isHTML {
		e.HTML = []byte(body)
	} else {
		e.Text = []byte(body)
	}

	// 配置SMTP服务器
	smtpHost := fmt.Sprintf("%s:%s", s.config.Mail.Host, s.config.Mail.Port)
	
	// 配置TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: s.config.Mail.TLSInsecure,
		ServerName:         s.config.Mail.Host,
	}

	// 如果使用SSL，则使用TLS连接
	if s.config.Mail.Encryption == "ssl" {
		tlsConfig.InsecureSkipVerify = true
	}

	// 发送邮件
	var err error
	if s.config.Mail.Encryption == "tls" || s.config.Mail.Encryption == "ssl" {
		err = e.SendWithTLS(smtpHost, smtp.PlainAuth("", s.config.Mail.Username, s.config.Mail.Password, s.config.Mail.Host), tlsConfig)
	} else {
		err = e.Send(smtpHost, smtp.PlainAuth("", s.config.Mail.Username, s.config.Mail.Password, s.config.Mail.Host))
	}

	if err != nil {
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
	`, s.config.App.URL, resetToken)

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
	`, s.config.App.URL, verificationToken)

	return s.SendEmail(to, subject, body, true)
}