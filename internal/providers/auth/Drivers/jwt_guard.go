package Drivers

import (
	"reflect"

	"fiber-starter/internal/providers/auth/Contracts"

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
func (g *JWTGuard) User(c fiber.Ctx) any {
	user := c.Locals("user")
	if user == nil {
		return nil
	}
	return user
}

// Id retrieves the ID for the currently authenticated user
func (g *JWTGuard) Id(c fiber.Ctx) int64 {
	if user := g.User(c); user != nil {
		return getInt64Field(user, "ID")
	}
	return 0
}

// SetUser sets the current user for the given context
func (g *JWTGuard) SetUser(c fiber.Ctx, user any) {
	c.Locals("user", user)
	c.Locals("user_id", getInt64Field(user, "ID"))
	c.Locals("user_email", getStringField(user, "Email"))
	c.Locals("user_name", getStringField(user, "Name"))
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
func (g *JWTGuard) Login(c fiber.Ctx, user any) error {
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

func getInt64Field(obj any, name string) int64 {
	if obj == nil {
		return 0
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return 0
	}
	f := val.FieldByName(name)
	if !f.IsValid() {
		return 0
	}
	if f.Kind() == reflect.Int64 {
		return f.Int()
	}
	return 0
}

func getStringField(obj any, name string) string {
	if obj == nil {
		return ""
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}
	f := val.FieldByName(name)
	if !f.IsValid() {
		return ""
	}
	if f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}
