// Package models 定义应用程序的数据模型
package models

import (
	"time"

	enums "fiber-starter/internal/domain/enum"
)

// User 用户模型
type User struct {
	ID              int64      `db:"id" gorm:"column:id;primaryKey" json:"id"`
	Name            string     `db:"name" gorm:"column:name" json:"name"`
	Email           string     `db:"email" gorm:"column:email" json:"email"`
	Password        string     `db:"password" gorm:"column:password" json:"-"`
	Avatar          *string    `db:"avatar" gorm:"column:avatar" json:"avatar,omitempty"`
	Phone           *string    `db:"phone" gorm:"column:phone" json:"phone,omitempty"`
	Status          UserStatus `db:"status" gorm:"column:status" json:"status"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" gorm:"column:email_verified_at" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" gorm:"column:updated_at" json:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at" gorm:"column:deleted_at" json:"-"`
}

// TableName returns the Atlas-managed table name for GORM.
func (User) TableName() string {
	return "users"
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
