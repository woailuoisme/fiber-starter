package auth

import (
	userPkg "lfiber/internal/features/user"

	"github.com/gofiber/fiber/v3"
)

type AuthTokensResource struct {
	Tokens TokenPair
}

func NewAuthTokensResource(tokens TokenPair) AuthTokensResource {
	return AuthTokensResource{Tokens: tokens}
}

func (r AuthTokensResource) ToResponse() fiber.Map {
	return fiber.Map{
		"access_token":  r.Tokens.AccessToken,
		"refresh_token": r.Tokens.RefreshToken,
	}
}

type AuthResultResource struct {
	Result AuthResult
}

func NewAuthResultResource(result AuthResult) AuthResultResource {
	return AuthResultResource{Result: result}
}

func (r AuthResultResource) ToResponse() fiber.Map {
	return fiber.Map{
		"user":   userPkg.NewUserResource(r.Result.User).ToResponse(),
		"tokens": NewAuthTokensResource(r.Result.Tokens).ToResponse(),
	}
}

type SignUpResource struct {
	User *SignUpResult
}

func NewSignUpResource(result SignUpResult) SignUpResource {
	return SignUpResource{User: &result}
}

func (r SignUpResource) ToResponse() fiber.Map {
	if r.User == nil {
		return fiber.Map{
			"verification_required": true,
		}
	}
	return fiber.Map{
		"user":                  userPkg.NewUserResource(r.User.User).ToResponse(),
		"verification_required": r.User.VerificationRequired,
	}
}
