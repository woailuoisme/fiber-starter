package auth

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"Alice"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

func (r RegisterRequest) ToInput() RegisterInput {
	return RegisterInput(r)
}

// VerifySignUpRequest 注册邮箱验证码请求
type VerifySignUpRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
	Code  string `json:"code" validate:"required,len=6" example:"123456"`
}

func (r VerifySignUpRequest) ToInput() VerifyCodeInput {
	return VerifyCodeInput(r)
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

func (r LoginRequest) ToInput() LoginInput {
	return LoginInput(r)
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` //nolint:lll
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

func (r ChangePasswordRequest) ToInput(userID int64) ChangePasswordInput {
	return ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: r.CurrentPassword,
		NewPassword:     r.NewPassword,
	}
}

// ResetPasswordRequest 请求密码重置验证码
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (r ResetPasswordRequest) ToInput() PasswordResetRequestInput {
	return PasswordResetRequestInput(r)
}

// VerifyResetPasswordRequest 验证重置密码验证码请求
type VerifyResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

func (r VerifyResetPasswordRequest) ToInput() VerifyCodeInput {
	return VerifyCodeInput(r)
}

// ConfirmResetPasswordRequest 确认重置密码请求
type ConfirmResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r ConfirmResetPasswordRequest) ToInput() ConfirmPasswordResetInput {
	return ConfirmPasswordResetInput{
		Token:       r.Token,
		NewPassword: r.Password,
	}
}
