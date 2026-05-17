package Drivers

import (
	models "fiber-starter/app/Models"
	"fiber-starter/app/Providers/Auth/Contracts"

	"github.com/gofiber/fiber/v3"
)

// JWTGuard implements the Guard interface for JWT-based authentication
type JWTGuard struct {
	provider Contracts.UserProvider
}

// NewJWTGuard creates a new JWT guard instance
func NewJWTGuard(provider Contracts.UserProvider) *JWTGuard {
	return &JWTGuard{provider: provider}
}

// Check determines if the current user is authenticated
func (g *JWTGuard) Check(c fiber.Ctx) bool {
	return g.User(c) != nil
}

// Guest determines if the current user is a guest
func (g *JWTGuard) Guest(c fiber.Ctx) bool {
	return !g.Check(c)
}

// User retrieves the currently authenticated user from context
func (g *JWTGuard) User(c fiber.Ctx) *models.User {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

// Id retrieves the ID for the currently authenticated user
func (g *JWTGuard) Id(c fiber.Ctx) int64 {
	if user := g.User(c); user != nil {
		return user.ID
	}
	return 0
}

// SetUser sets the current user for the given context
func (g *JWTGuard) SetUser(c fiber.Ctx, user *models.User) {
	c.Locals("user", user)
	c.Locals("user_id", user.ID)
	c.Locals("user_email", user.Email)
	c.Locals("user_name", user.Name)
}

// Validate validates a user's credentials without logging them in
func (g *JWTGuard) Validate(credentials map[string]string) bool {
	user, err := g.provider.RetrieveByCredentials(credentials)
	if err != nil || user == nil {
		return false
	}
	return g.provider.ValidateCredentials(user, credentials)
}

// Attempt attempts to authenticate a user using the given credentials
func (g *JWTGuard) Attempt(c fiber.Ctx, credentials map[string]string) bool {
	user, err := g.provider.RetrieveByCredentials(credentials)
	if err != nil || user == nil {
		return false
	}
	if g.provider.ValidateCredentials(user, credentials) {
		g.SetUser(c, user)
		return true
	}
	return false
}

// Login logs the given user instance into the application
func (g *JWTGuard) Login(c fiber.Ctx, user *models.User) error {
	g.SetUser(c, user)
	return nil
}

// LoginUsingId logs the user with the given ID into the application
func (g *JWTGuard) LoginUsingId(c fiber.Ctx, id int64) error {
	user, err := g.provider.RetrieveById(id)
	if err != nil || user == nil {
		return err
	}
	g.SetUser(c, user)
	return nil
}

// Logout logs the user out of the application
func (g *JWTGuard) Logout(c fiber.Ctx) error {
	// For JWT, the actual token invalidation is typically handled
	// by the AuthService's Logout method which manages the blacklist.
	// This method can be used to clear the context user.
	c.Locals("user", nil)
	c.Locals("user_id", nil)
	return nil
}
