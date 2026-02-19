// Package controllers implements the HTTP controllers.
package controllers

import (
	models "fiber-starter/internal/domain/model"
	apierrors "fiber-starter/internal/platform/apierrors"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/services"
	"fiber-starter/internal/transport/http/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// AuthController 认证控制器
type AuthController struct {
	authService services.AuthService
	validate    *validator.Validate
}

// NewAuthController 创建认证控制器实例
func NewAuthController(authService services.AuthService, validate *validator.Validate) *AuthController {
	return &AuthController{
		authService: authService,
		validate:    validate,
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"Alice" swagger:"required,user_name"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com" swagger:"required,email_address"`
	Password string `json:"password" validate:"required,min=6" example:"password123" swagger:"required,password"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com" swagger:"required,email_address"`
	Password string `json:"password" validate:"required" example:"password123" swagger:"required,password"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." swagger:"required,refresh_token"` //nolint:lll
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// ResetPasswordRequest 重置密码请求结构
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmResetPasswordRequest 确认重置密码请求结构
type ConfirmResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// Register 注册新用户
// @Summary Register
// @Description Create a new user account.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register payload"
// @Success 201 {object} resources.APIResponse "Created"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 409 {object} resources.APIResponse "Email already registered"
// @Router /api/v1/auth/register [post]
func (c *AuthController) Register(ctx fiber.Ctx) error {
	var req RegisterRequest

	// 解析和验证请求参数
	if err := helpers.ParseAndValidate(ctx, &req, c.validate); err != nil {
		return helpers.HandleError(ctx, err)
	}

	// 调用认证服务注册用户
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	err := c.authService.Register(user)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	// 注册成功后自动登录
	loginUser, accessToken, refreshToken, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleCreated(ctx, "Registered successfully", fiber.Map{
		"user":          loginUser.ToSafeResponse(),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Login 用户登录
// @Summary Login
// @Description Authenticate a user.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Router /api/v1/auth/login [post]
func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req LoginRequest

	// 解析和验证请求参数
	if err := helpers.ParseAndValidate(ctx, &req, c.validate); err != nil {
		return helpers.HandleError(ctx, err)
	}

	// 调用认证服务登录
	user, accessToken, refreshToken, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	// 构建响应数据
	response := fiber.Map{
		"user": user.ToSafeUser(),
		"tokens": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}

	return helpers.HandleSuccess(ctx, "Logged in successfully", response)
}

// RefreshToken 刷新访问令牌
// @Summary Refresh token
// @Description Exchange a refresh token for a new access token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Router /api/v1/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx fiber.Ctx) error {
	var req RefreshTokenRequest

	// 解析和验证请求参数
	if err := helpers.ParseAndValidate(ctx, &req, c.validate); err != nil {
		return helpers.HandleError(ctx, err)
	}

	// 调用认证服务刷新令牌
	accessToken, refreshToken, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	// 返回刷新成功响应
	return helpers.HandleSuccess(ctx, "Token refreshed successfully", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout 用户登出
// @Summary Logout
// @Description Log out the current user.
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 500 {object} resources.APIResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (c *AuthController) Logout(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, apierrors.Unauthorized("unauthenticated user"))
	}

	// 获取令牌
	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return helpers.HandleError(ctx, apierrors.BadRequest("failed to resolve access token"))
	}

	// 调用认证服务登出
	err := c.authService.Logout(token)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Logged out successfully", nil)
}

// ChangePassword 修改密码
// @Summary Change password
// @Description Change the current user's password.
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ChangePasswordRequest true "Change password payload"
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Router /api/v1/auth/change-password [post]
func (c *AuthController) ChangePassword(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, apierrors.Unauthorized("unauthenticated user"))
	}

	var req ChangePasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, apierrors.BadRequest("failed to parse request body"))
	}

	// 验证请求参数
	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, apierrors.ValidationWithDetails("request validation failed", helpers.FormatValidationErrorsToString(err)))
	}

	// 调用认证服务修改密码
	err := c.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password changed successfully", nil)
}

// ResetPassword 重置密码
// @Summary Reset password
// @Description Send a password reset email.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password payload"
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Router /api/v1/auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, apierrors.BadRequest("failed to parse request body"))
	}

	// 验证请求参数
	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, apierrors.ValidationWithDetails("request validation failed", helpers.FormatValidationErrorsToString(err)))
	}

	// 调用认证服务发送重置邮件
	err := c.authService.ForgotPassword(req.Email)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset email sent", nil)
}

// ConfirmResetPassword 确认重置密码
// @Summary Confirm reset password
// @Description Set a new password using a reset token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ConfirmResetPasswordRequest true "Confirm reset password payload"
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Router /api/v1/auth/confirm-reset-password [post]
func (c *AuthController) ConfirmResetPassword(ctx fiber.Ctx) error {
	var req ConfirmResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, apierrors.BadRequest("failed to parse request body"))
	}

	// 验证请求参数
	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, apierrors.ValidationWithDetails("request validation failed", helpers.FormatValidationErrorsToString(err)))
	}

	// 调用认证服务重置密码
	err := c.authService.ResetPassword(req.Token, "", req.Password)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset successfully", nil)
}

// GetProfile 获取用户资料
// @Summary Get profile
// @Description Get the current user's profile.
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Router /api/v1/auth/profile [get]
func (c *AuthController) GetProfile(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, apierrors.Unauthorized("unauthenticated user"))
	}

	return helpers.HandleSuccess(ctx, "Profile fetched successfully", fiber.Map{
		"user": user.ToSafeResponse(),
	})
}
