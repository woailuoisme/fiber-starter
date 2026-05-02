package requests

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"Alice"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"password123"`
}

func (r *RegisterRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

func (r *LoginRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` //nolint:lll
}

func (r *RefreshTokenRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

func (r *ChangePasswordRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}

// ResetPasswordRequest 密码重置请求
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (r *ResetPasswordRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}

// ConfirmResetPasswordRequest 确认重置密码请求
type ConfirmResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

func (r *ConfirmResetPasswordRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}
