package auth

import (
	"fiber-starter/internal/providers/auth/Contracts"
	"fiber-starter/internal/support/appctx"

	"github.com/gofiber/fiber/v3"
)

// manager returns the auth manager instance from the container.
func manager() Contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.AuthManager()
	}
	return nil
}

// Guard returns an authentication guard by name
func Guard(name ...string) Contracts.Guard {
	if m := manager(); m != nil {
		return m.Guard(name...)
	}
	return nil
}

// User retrieves the currently authenticated user from the default guard
func User(c fiber.Ctx) any {
	if g := Guard(); g != nil {
		return g.User(c)
	}
	return nil
}

// Id retrieves the ID for the currently authenticated user from the default guard
func Id(c fiber.Ctx) int64 {
	if g := Guard(); g != nil {
		return g.Id(c)
	}
	return 0
}

// Check determines if the current user is authenticated from the default guard
func Check(c fiber.Ctx) bool {
	if g := Guard(); g != nil {
		return g.Check(c)
	}
	return false
}

// Guest determines if the current user is a guest
func Guest(c fiber.Ctx) bool {
	if g := Guard(); g != nil {
		return g.Guest(c)
	}
	return true
}

// SetUser sets the current user for the given context using the default guard
func SetUser(c fiber.Ctx, user any) {
	if g := Guard(); g != nil {
		g.SetUser(c, user)
	}
}

// Attempt attempts to authenticate a user using the given credentials with the default guard
func Attempt(c fiber.Ctx, credentials map[string]string) bool {
	if g := Guard(); g != nil {
		return g.Attempt(c, credentials)
	}
	return false
}

// Validate validates a user's credentials with the default guard
func Validate(credentials map[string]string) bool {
	if g := Guard(); g != nil {
		return g.Validate(credentials)
	}
	return false
}

// Login logs the given user instance into the application
func Login(c fiber.Ctx, user any) error {
	if g := Guard(); g != nil {
		return g.Login(c, user)
	}
	return nil
}

// LoginUsingId logs the user with the given ID into the application
func LoginUsingId(c fiber.Ctx, id int64) error {
	if g := Guard(); g != nil {
		return g.LoginUsingId(c, id)
	}
	return nil
}

// Logout logs the user out using the default guard
func Logout(c fiber.Ctx) error {
	if g := Guard(); g != nil {
		return g.Logout(c)
	}
	return nil
}
