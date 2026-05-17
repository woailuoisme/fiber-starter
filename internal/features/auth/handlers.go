package auth

import (
	"reflect"

	exceptions "fiber-starter/internal/common/exceptions"
	middleware "fiber-starter/internal/common/middleware"
	helpers "fiber-starter/internal/common/support"
	userModel "fiber-starter/internal/features/user"

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
