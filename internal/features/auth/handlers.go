package auth

import (
	"reflect"

	exceptions "lfiber/internal/common/exceptions"
	middleware "lfiber/internal/common/middleware"
	userModel "lfiber/internal/features/user"
	helpers "lfiber/internal/support"

	"github.com/gofiber/fiber/v3"
)

// AuthRouteHandler defines the auth handlers used by route registration.
type AuthRouteHandler interface {
	SignUp(ctx fiber.Ctx) error
	VerifySignUp(ctx fiber.Ctx) error
	SignIn(ctx fiber.Ctx) error
	RefreshSession(ctx fiber.Ctx) error
	SignOut(ctx fiber.Ctx) error
	UpdatePassword(ctx fiber.Ctx) error
	SendPasswordReset(ctx fiber.Ctx) error
	VerifyPasswordReset(ctx fiber.Ctx) error
	ConfirmPasswordReset(ctx fiber.Ctx) error
	Session(ctx fiber.Ctx) error
}

// AuthController 认证控制器
type AuthController struct {
	authService AuthService
}

// NewAuthController 创建认证控制器实例
func NewAuthController(authService AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SignUp creates a new account and sends an email verification code.
//
//	@Summary		用户注册
//	@Description	创建一个新的用户账号并发送邮箱验证码，不直接返回访问令牌。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest															true	"注册参数"
//	@Success		201		{object}	support.APISuccessResponse{data=object{user=user.SafeUser,verification_required=bool}}	"注册成功"
//	@Router			/api/v1/auth/sign-up [post]
func (c *AuthController) SignUp(ctx fiber.Ctx) error {
	var req RegisterRequest

	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	result, err := c.authService.Register(ctx.Context(), req.ToInput())
	if err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleCreated(ctx, "Registered successfully", NewSignUpResource(result).ToResponse())
}

// Register is a compatibility wrapper for SignUp.
func (c *AuthController) Register(ctx fiber.Ctx) error {
	return c.SignUp(ctx)
}

// VerifySignUp verifies the signup OTP and returns a session.
//
//	@Summary		验证注册邮箱
//	@Description	校验邮箱验证码并激活账号，成功后返回访问令牌和刷新令牌。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		VerifySignUpRequest																								true	"验证参数"
//	@Success		200		{object}	support.APISuccessResponse{data=object{user=user.SafeUser,tokens=object{access_token=string,refresh_token=string}}}	"验证成功"
//	@Router			/api/v1/auth/sign-up/verify [post]
func (c *AuthController) VerifySignUp(ctx fiber.Ctx) error {
	var req VerifySignUpRequest

	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	result, err := c.authService.VerifySignUp(ctx.Context(), req.ToInput())
	if err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Email verified successfully", NewAuthResultResource(result).ToResponse())
}

// SignIn authenticates a user and returns the session tokens.
//
//	@Summary		用户登录
//	@Description	验证用户凭据，成功后返回 JWT 访问令牌和刷新令牌；未完成邮箱验证的账号会被拒绝。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest																									true	"登录参数"
//	@Success		200		{object}	support.APISuccessResponse{data=object{user=user.SafeUser,tokens=object{access_token=string,refresh_token=string}}}	"登录成功"
//	@Router			/api/v1/auth/sign-in [post]
func (c *AuthController) SignIn(ctx fiber.Ctx) error {
	var req LoginRequest

	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	result, err := c.authService.Login(ctx.Context(), req.ToInput())
	if err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Logged in successfully", NewAuthResultResource(result).ToResponse())
}

// Login is a compatibility wrapper for SignIn.
func (c *AuthController) Login(ctx fiber.Ctx) error {
	return c.SignIn(ctx)
}

// RefreshSession refreshes the access and refresh tokens.
//
//	@Summary		刷新令牌
//	@Description	使用有效的刷新令牌获取一组新的访问令牌和刷新令牌。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshTokenRequest														true	"刷新参数"
//	@Success		200		{object}	support.APISuccessResponse{data=object{access_token=string,refresh_token=string}}	"刷新成功"
//	@Router			/api/v1/auth/refresh [post]
func (c *AuthController) RefreshSession(ctx fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	tokens, err := c.authService.RefreshToken(ctx.Context(), req.RefreshToken)
	if err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Token refreshed successfully", NewAuthTokensResource(tokens).ToResponse())
}

// RefreshToken is a compatibility wrapper for RefreshSession.
func (c *AuthController) RefreshToken(ctx fiber.Ctx) error {
	return c.RefreshSession(ctx)
}

// SignOut logs the current user out.
//
//	@Summary		用户登出
//	@Description	撤销当前访问令牌，使用户退出登录状态。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	support.APISuccessNoDataResponse	"注销成功"
//	@Router			/api/v1/auth/sign-out [post]
func (c *AuthController) SignOut(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleAppError(ctx, exceptions.NewAuthenticationException("unauthenticated user"))
	}

	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return helpers.HandleAppError(ctx, exceptions.NewBadRequestException("failed to resolve access token"))
	}

	if err := c.authService.Logout(ctx.Context(), token); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Logged out successfully", nil)
}

// Logout is a compatibility wrapper for SignOut.
func (c *AuthController) Logout(ctx fiber.Ctx) error {
	return c.SignOut(ctx)
}

// UpdatePassword changes the current user's password.
//
//	@Summary		修改密码
//	@Description	在用户已登录的情况下，验证旧密码并更新为新密码。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		ChangePasswordRequest	true	"修改参数"
//	@Success		200		{object}	support.APISuccessNoDataResponse		"修改成功"
//	@Router			/api/v1/auth/change-password [post]
func (c *AuthController) UpdatePassword(ctx fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == 0 {
		return helpers.HandleAppError(ctx, exceptions.NewAuthenticationException("unauthenticated user"))
	}

	var req ChangePasswordRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	if err := c.authService.ChangePassword(ctx.Context(), req.ToInput(userID)); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password changed successfully", nil)
}

// ChangePassword is a compatibility wrapper for UpdatePassword.
func (c *AuthController) ChangePassword(ctx fiber.Ctx) error {
	return c.UpdatePassword(ctx)
}

// SendPasswordReset sends a reset email to the user.
//
//	@Summary		重置密码
//	@Description	发送密码重置验证码。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ResetPasswordRequest	true	"重置密码参数"
//	@Success		200		{object}	support.APISuccessNoDataResponse		"成功"
//	@Router			/api/v1/auth/reset-password [post]
func (c *AuthController) SendPasswordReset(ctx fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	if err := c.authService.RequestPasswordReset(ctx.Context(), req.ToInput()); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset code sent", nil)
}

// ResetPassword is a compatibility wrapper for SendPasswordReset.
func (c *AuthController) ResetPassword(ctx fiber.Ctx) error {
	return c.SendPasswordReset(ctx)
}

// VerifyPasswordReset verifies the password reset OTP and returns a short-lived reset token.
//
//	@Summary		验证重置密码验证码
//	@Description	使用邮箱验证码换取一个短期重置令牌。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		VerifyResetPasswordRequest								true	"验证重置密码参数"
//	@Success		200		{object}	support.APISuccessResponse{data=object{reset_token=string}}	"验证成功"
//	@Router			/api/v1/auth/reset-password/verify [post]
func (c *AuthController) VerifyPasswordReset(ctx fiber.Ctx) error {
	var req VerifyResetPasswordRequest

	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	resetToken, err := c.authService.VerifyPasswordReset(ctx.Context(), req.ToInput())
	if err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset code verified successfully", fiber.Map{
		"reset_token": resetToken.Token,
	})
}

// ConfirmPasswordReset confirms a password reset token.
//
//	@Summary		确认重置密码
//	@Description	使用重置令牌设置新密码。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ConfirmResetPasswordRequest	true	"确认重置密码参数"
//	@Success		200		{object}	support.APISuccessNoDataResponse			"成功"
//	@Router			/api/v1/auth/reset-password/confirm [post]
func (c *AuthController) ConfirmPasswordReset(ctx fiber.Ctx) error {
	var req ConfirmResetPasswordRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	if err := c.authService.ResetPassword(ctx.Context(), req.ToInput()); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Password reset successfully", nil)
}

// ConfirmResetPassword is a compatibility wrapper for ConfirmPasswordReset.
func (c *AuthController) ConfirmResetPassword(ctx fiber.Ctx) error {
	return c.ConfirmPasswordReset(ctx)
}

// Session returns the current authenticated user's profile.
//
//	@Summary		获取当前用户资料
//	@Description	根据访问令牌返回当前登录用户的详细个人信息。
//	@Tags			认证中心
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	support.APISuccessResponse{data=object{user=user.SafeUser}}	"获取成功"
//	@Router			/api/v1/auth/session [get]
func (c *AuthController) Session(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleAppError(ctx, exceptions.NewAuthenticationException("unauthenticated user"))
	}

	if usr, ok := user.(*userModel.User); ok {
		return helpers.HandleSuccess(ctx, "Profile fetched successfully", fiber.Map{
			"user": usr.ToSafeUser(),
		})
	}

	// Dynamic fallback for other types (JWT AuthUser, etc.)
	userID := middleware.GetCurrentUserID(ctx)
	name := ""
	email := ""
	if val := reflect.ValueOf(user); val.Kind() == reflect.Pointer {
		val = val.Elem()
		if val.Kind() == reflect.Struct {
			if f := val.FieldByName("Name"); f.IsValid() && f.Kind() == reflect.String {
				name = f.String()
			}
			if f := val.FieldByName("Email"); f.IsValid() && f.Kind() == reflect.String {
				email = f.String()
			}
		}
	}

	return helpers.HandleSuccess(ctx, "Profile fetched successfully", fiber.Map{
		"user": userModel.SafeUser{
			ID:    userID,
			Name:  name,
			Email: email,
		},
	})
}

// GetProfile is a compatibility wrapper for Session.
func (c *AuthController) GetProfile(ctx fiber.Ctx) error {
	return c.Session(ctx)
}
