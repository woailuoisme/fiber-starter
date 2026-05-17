package models

import (
	"time"

	enums "fiber-starter/app/Enums"

	"github.com/uptrace/bun"
)

// User 用户模型
type User struct {
	bun.BaseModel   `bun:"table:users,alias:u"`
	ID              int64      `bun:"id,pk,autoincrement" json:"id"`
	Name            string     `bun:"name,notnull" json:"name"`
	Email           string     `bun:"email,notnull" json:"email"`
	Password        string     `bun:"password,notnull" json:"-"`
	Avatar          *string    `bun:"avatar" json:"avatar,omitempty"`
	Phone           *string    `bun:"phone" json:"phone,omitempty"`
	Status          UserStatus `bun:"status,notnull" json:"status"`
	EmailVerifiedAt *time.Time `bun:"email_verified_at" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull" json:"updated_at"`
	DeletedAt       *time.Time `bun:"deleted_at" json:"-"`
}

// UserStatus 用户状态枚举
type UserStatus = enums.UserStatus

const (
	// UserStatusActive active user status
	UserStatusActive = enums.UserStatusActive
	// UserStatusInactive inactive user status
	UserStatusInactive = enums.UserStatusInactive
	// UserStatusPending pending user status
	UserStatusPending = enums.UserStatusPending
	// UserStatusSuspended suspended user status
	UserStatusSuspended = enums.UserStatusSuspended
	// UserStatusBanned banned user status
	UserStatusBanned = enums.UserStatusBanned
)

// IsEmailVerified 检查邮箱是否已验证
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsActive 检查用户是否处于活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// ToSafeUser 转换为安全的用户信息（不包含敏感信息）
func (u *User) ToSafeUser() SafeUser {
	return SafeUser{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Avatar:          u.Avatar,
		Phone:           u.Phone,
		Status:          u.Status,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// SafeUser 安全的用户信息（用于API响应）
type SafeUser struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Avatar          *string    `json:"avatar,omitempty"`
	Phone           *string    `json:"phone,omitempty"`
	Status          UserStatus `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
