package controllers

import (
	"fiber-starter/app/Exceptions"
	"fiber-starter/app/Http/Middleware"
	services "fiber-starter/app/Http/Services"
	models "fiber-starter/app/Models"
	helpers "fiber-starter/app/Support"

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

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"Alice" swagger:"required,user_name"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com" swagger:"required,email_address"`
	Password string `json:"password" validate:"required,min=6" example:"password123" swagger:"required,password"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com" swagger:"required,email_address"`
	Password string `json:"password" validate:"required" example:"password123" swagger:"required,password"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." swagger:"required,refresh_token"` //nolint:lll
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// ResetPasswordRequest 密码重置请求
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmResetPasswordRequest 确认重置密码请求
type ConfirmResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建一个新的用户账号。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册参数"
// @Success 201 {object} resources.APIResponse "创建成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 409 {object} resources.APIResponse "邮箱已被注册"
// @Router /api/v1/auth/register [post]
func (c *AuthController) Register(ctx fiber.Ctx) error {
	var req RegisterRequest

	if err := helpers.ParseAndValidate(ctx, &req, c.validate); err != nil {
		return helpers.HandleError(ctx, err)
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	if err := c.authService.Register(user); err != nil {
		return helpers.HandleError(ctx, err)
	}

	loginUser, accessToken, refreshToken, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleCreated(ctx, "Registered successfully", fiber.Map{
		"user":          loginUser.ToSafeUser(),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 验证用户身份并返回令牌。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Router /api/v1/auth/login [post]
func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req LoginRequest

	if err := helpers.ParseAndValidate(ctx, &req, c.validate); err != nil {
		return helpers.HandleError(ctx, err)
	}

	user, accessToken, refreshToken, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Logged in successfully", fiber.Map{
		"user": user.ToSafeUser(),
		"tokens": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// RefreshToken 刷新访问令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌换取新的访问令牌。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌参数"
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Router /api/v1/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := helpers.ParseAndValidate(ctx, &req, c.validate); err != nil {
		return helpers.HandleError(ctx, err)
	}

	accessToken, refreshToken, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Token refreshed successfully", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 退出当前登录用户。
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 500 {object} resources.APIResponse "服务器错误"
// @Router /api/v1/auth/logout [post]
func (c *AuthController) Logout(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, exceptions.Unauthorized("unauthenticated user"))
	}

	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return helpers.HandleError(ctx, exceptions.BadRequest("failed to resolve access token"))
	}

	if err := c.authService.Logout(token); err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Logged out successfully", nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码。
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ChangePasswordRequest true "修改密码参数"
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Router /api/v1/auth/change-password [post]
func (c *AuthController) ChangePassword(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, exceptions.Unauthorized("unauthenticated user"))
	}

	var req ChangePasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, exceptions.BadRequest("failed to parse request body"))
	}

	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, exceptions.ValidationWithDetails("request validation failed", helpers.FormatValidationErrorsToString(err)))
	}

	if err := c.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password changed successfully", nil)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 发送密码重置邮件。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "重置密码参数"
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Router /api/v1/auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, exceptions.BadRequest("failed to parse request body"))
	}

	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, exceptions.ValidationWithDetails("request validation failed", helpers.FormatValidationErrorsToString(err)))
	}

	if err := c.authService.ForgotPassword(req.Email); err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset email sent", nil)
}

// ConfirmResetPassword 确认重置密码
// @Summary 确认重置密码
// @Description 使用重置令牌设置新密码。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ConfirmResetPasswordRequest true "确认重置密码参数"
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Router /api/v1/auth/confirm-reset-password [post]
func (c *AuthController) ConfirmResetPassword(ctx fiber.Ctx) error {
	var req ConfirmResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleError(ctx, exceptions.BadRequest("failed to parse request body"))
	}

	if err := c.validate.Struct(req); err != nil {
		return helpers.HandleError(ctx, exceptions.ValidationWithDetails("request validation failed", helpers.FormatValidationErrorsToString(err)))
	}

	if err := c.authService.ResetPassword(req.Token, "", req.Password); err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset successfully", nil)
}

// GetProfile 获取用户资料
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的资料。
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Router /api/v1/auth/profile [get]
func (c *AuthController) GetProfile(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleError(ctx, exceptions.Unauthorized("unauthenticated user"))
	}

	return helpers.HandleSuccess(ctx, "Profile fetched successfully", fiber.Map{
		"user": user.ToSafeUser(),
	})
}
