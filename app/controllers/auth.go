package controllers

import (
	"fiber-starter/app/services"
	"fiber-starter/app/models"
	"fiber-starter/app/middleware"
	"fiber-starter/app/helpers"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
)

// AuthController 认证控制器
type AuthController struct {
	authService services.AuthService
	validate    *validator.Validate
}

// NewAuthController 创建认证控制器实例
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
		validate:    validator.New(),
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
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

// Register 用户注册
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var req RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 调用认证服务注册用户
	user, err := c.authService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	response := fiber.Map{
		"user": user.ToSafeResponse(),
	}

	return ctx.Status(fiber.StatusCreated).JSON(helpers.SuccessResponse("注册成功", response))
}

// Login 用户登录
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 调用认证服务登录
	tokens, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("登录成功", tokens))
}

// RefreshToken 刷新访问令牌
func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 调用认证服务刷新令牌
	tokens, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("令牌刷新成功", tokens))
}

// Logout 用户登出
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse("未认证用户", nil))
	}

	// 获取令牌
	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("无法获取访问令牌", nil))
	}

	// 调用认证服务登出
	err := c.authService.Logout(token)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("登出成功", nil))
}

// ChangePassword 修改密码
func (c *AuthController) ChangePassword(ctx *fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse("未认证用户", nil))
	}

	var req ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 调用认证服务修改密码
	err := c.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("密码修改成功", nil))
}

// ResetPassword 重置密码
func (c *AuthController) ResetPassword(ctx *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 调用认证服务重置密码
	err := c.authService.ResetPassword(req.Email)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("密码重置邮件已发送", nil))
}

// ConfirmResetPassword 确认重置密码
func (c *AuthController) ConfirmResetPassword(ctx *fiber.Ctx) error {
	var req ConfirmResetPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 调用认证服务确认重置密码
	err := c.authService.ConfirmResetPassword(req.Token, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("密码重置成功", nil))
}

// GetProfile 获取用户资料
func (c *AuthController) GetProfile(ctx *fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse("未认证用户", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("获取用户资料成功", fiber.Map{
		"user": user.ToSafeResponse(),
	}))
}