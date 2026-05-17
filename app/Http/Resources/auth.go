package resources

import (
	services "fiber-starter/app/Http/Services"

	"github.com/gofiber/fiber/v3"
)

type AuthTokensResource struct {
	Tokens services.TokenPair
}

func NewAuthTokensResource(tokens services.TokenPair) AuthTokensResource {
	return AuthTokensResource{Tokens: tokens}
}

func (r AuthTokensResource) ToResponse() fiber.Map {
	return fiber.Map{
		"access_token":  r.Tokens.AccessToken,
		"refresh_token": r.Tokens.RefreshToken,
	}
}

type AuthResultResource struct {
	Result services.AuthResult
}

func NewAuthResultResource(result services.AuthResult) AuthResultResource {
	return AuthResultResource{Result: result}
}

func (r AuthResultResource) ToResponse() fiber.Map {
	return fiber.Map{
		"user":   NewUserResource(r.Result.User).ToResponse(),
		"tokens": NewAuthTokensResource(r.Result.Tokens).ToResponse(),
	}
}

type SignUpResource struct {
	User *services.SignUpResult
}

func NewSignUpResource(result services.SignUpResult) SignUpResource {
	return SignUpResource{User: &result}
}

func (r SignUpResource) ToResponse() fiber.Map {
	if r.User == nil {
		return fiber.Map{
			"verification_required": true,
		}
	}
	return fiber.Map{
		"user":                  NewUserResource(r.User.User).ToResponse(),
		"verification_required": r.User.VerificationRequired,
	}
}
