// Package models 定义应用程序的数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Email           string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password        string         `gorm:"size:255;not null" json:"-"`
	Avatar          *string        `gorm:"size:500" json:"avatar,omitempty"`
	Phone           *string        `gorm:"size:20" json:"phone,omitempty"`
	Status          UserStatus     `gorm:"type:enum('active','inactive','banned');default:'active'" json:"status"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserStatus 用户状态枚举
type UserStatus string

const (
	// UserStatusActive active user status
	UserStatusActive UserStatus = "active"
	// UserStatusInactive inactive user status
	UserStatusInactive UserStatus = "inactive"
	// UserStatusBanned banned user status
	UserStatusBanned UserStatus = "banned"
)

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前钩子
func (u *User) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

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

// ToSafeResponse 转换为安全的用户响应（用于API输出，与ToSafeUser功能相同）
func (u *User) ToSafeResponse() SafeUser {
	return u.ToSafeUser()
}

// SafeUser 安全的用户信息（用于API响应）
type SafeUser struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Avatar          *string    `json:"avatar,omitempty"`
	Phone           *string    `json:"phone,omitempty"`
	Status          UserStatus `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
