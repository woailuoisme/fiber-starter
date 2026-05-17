package auth

import (
	"time"

	"github.com/uptrace/bun"
)

// AuthOTPPurpose 表示 OTP 用途
type AuthOTPPurpose string

const (
	// AuthOTPPurposeSignup 注册验证
	AuthOTPPurposeSignup AuthOTPPurpose = "signup"
	// AuthOTPPurposePasswordReset 密码重置
	AuthOTPPurposePasswordReset AuthOTPPurpose = "password_reset"
)

// AuthOTP 认证 OTP 记录
type AuthOTP struct {
	bun.BaseModel `bun:"table:auth_otps,alias:ao"`
	ID            int64          `bun:"id,pk,autoincrement" json:"id"`
	Email         string         `bun:"email,notnull" json:"email"`
	Purpose       AuthOTPPurpose `bun:"purpose,notnull" json:"purpose"`
	CodeHash      string         `bun:"code_hash,notnull" json:"-"`
	ExpiresAt     time.Time      `bun:"expires_at,notnull" json:"expires_at"`
	SentAt        time.Time      `bun:"sent_at,notnull" json:"sent_at"`
	Attempts      int64          `bun:"attempts,notnull" json:"attempts"`
	MaxAttempts   int64          `bun:"max_attempts,notnull" json:"max_attempts"`
	ConsumedAt    *time.Time     `bun:"consumed_at" json:"consumed_at,omitempty"`
	CreatedAt     time.Time      `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt     time.Time      `bun:"updated_at,notnull" json:"updated_at"`
}

// AuthOTPRecord 保持对旧命名的兼容。
type AuthOTPRecord = AuthOTP
