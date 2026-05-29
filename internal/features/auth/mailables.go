package auth

import (
	"fmt"
	"time"
)

// SignupVerificationMail 注册邮箱验证邮件 Mailable。
// 用于发送注册激活验证码，保证用户的邮箱有效性。
type SignupVerificationMail struct {
	Name      string        // 用户姓名/昵称
	Code      string        // 验证码
	ExpiresIn time.Duration // 过期时间
}

func (m SignupVerificationMail) Subject() string {
	return "Verify your email"
}

func (m SignupVerificationMail) Template() (string, map[string]interface{}) {
	expiresStr := fmt.Sprintf("%d minutes", int(m.ExpiresIn.Minutes()))
	return "verification_code", map[string]interface{}{
		"Name":      m.Name,
		"Code":      m.Code,
		"ExpiresIn": expiresStr,
	}
}

// PasswordResetOTPMail 重置密码验证码邮件 Mailable。
// 用于向用户发送找回密码所需的验证码。
type PasswordResetOTPMail struct {
	Code      string        // 验证码
	ExpiresIn time.Duration // 过期时间
}

func (m PasswordResetOTPMail) Subject() string {
	return "Reset your password"
}

func (m PasswordResetOTPMail) Template() (string, map[string]interface{}) {
	expiresStr := fmt.Sprintf("%d minutes", int(m.ExpiresIn.Minutes()))
	return "verification_code", map[string]interface{}{
		"Code":      m.Code,
		"ExpiresIn": expiresStr,
	}
}
