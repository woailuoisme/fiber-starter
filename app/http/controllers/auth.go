// Package controllers implements the HTTP controllers.
package controllers

import (
	apierrors "fiber-starter/app/errors"
	"fiber-starter/app/helpers"
	"fiber-starter/app/http/middleware"
	"fiber-starter/app/models"
	"fiber-starter/app/services"

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
	Name     string `json:"name" validate:"required,min=2,max=100" example:"张三" swagger:"required,用户姓名"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com" swagger:"required,邮箱地址"`
	Password string `json:"password" validate:"required,min=6" example:"password123" swagger:"required,密码"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com" swagger:"required,邮箱地址"`
	Password string `json:"password" validate:"required" example:"password123" swagger:"required,密码"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." swagger:"required,刷新令牌"` //nolint:lll
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
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} helpers.APIResponse "注册成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 409 {object} helpers.APIResponse "邮箱已被注册"
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

	return helpers.HandleCreated(ctx, "注册成功", fiber.Map{
		"user":          loginUser.ToSafeResponse(),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户身份验证
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} helpers.APIResponse "登录成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 401 {object} helpers.APIResponse "认证失败"
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

	return helpers.HandleSuccess(ctx, "登录成功", response)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌信息"
// @Success 200 {object} helpers.APIResponse "刷新成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 401 {object} helpers.APIResponse "认证失败"
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
	return helpers.HandleSuccess(ctx, "令牌刷新成功", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户退出登录
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "登出成功"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 500 {object} helpers.APIResponse "服务器错误"
// @Router /api/v1/auth/logout [post]
func (c *AuthController) Logout(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, apierrors.Unauthorized("未认证用户"))
	}

	// 获取令牌
	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return helpers.HandleError(ctx, apierrors.BadRequest("无法获取访问令牌"))
	}

	// 调用认证服务登出
	err := c.authService.Logout(token)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "登出成功", nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} helpers.APIResponse "修改成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Router /api/v1/auth/change-password [post]
func (c *AuthController) ChangePassword(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, apierrors.Unauthorized("未认证用户"))
	}

	var req ChangePasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, apierrors.BadRequest("请求参数解析失败"))
	}

	// 验证请求参数
	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, apierrors.ValidationWithDetails("请求参数验证失败", helpers.FormatValidationErrorsToString(err)))
	}

	// 调用认证服务修改密码
	err := c.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "密码修改成功", nil)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 发送密码重置邮件
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} helpers.APIResponse "邮件发送成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Router /api/v1/auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, apierrors.BadRequest("请求参数解析失败"))
	}

	// 验证请求参数
	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, apierrors.ValidationWithDetails("请求参数验证失败", helpers.FormatValidationErrorsToString(err)))
	}

	// 调用认证服务发送重置邮件
	err := c.authService.ForgotPassword(req.Email)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "密码重置邮件已发送", nil)
}

// ConfirmResetPassword 确认重置密码
// @Summary 确认重置密码
// @Description 使用重置令牌设置新密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ConfirmResetPasswordRequest true "确认重置密码请求"
// @Success 200 {object} helpers.APIResponse "重置成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Router /api/v1/auth/confirm-reset-password [post]
func (c *AuthController) ConfirmResetPassword(ctx fiber.Ctx) error {
	var req ConfirmResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, apierrors.BadRequest("请求参数解析失败"))
	}

	// 验证请求参数
	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, apierrors.ValidationWithDetails("请求参数验证失败", helpers.FormatValidationErrorsToString(err)))
	}

	// 调用认证服务重置密码
	err := c.authService.ResetPassword(req.Token, "", req.Password)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "密码重置成功", nil)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前用户的个人资料
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "获取成功"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Router /api/v1/auth/profile [get]
func (c *AuthController) GetProfile(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, apierrors.Unauthorized("未认证用户"))
	}

	return helpers.HandleSuccess(ctx, "获取用户资料成功", fiber.Map{
		"user": user.ToSafeResponse(),
	})
}
