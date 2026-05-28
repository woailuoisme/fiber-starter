package contracts

import "github.com/gofiber/fiber/v3"

// Authorizer exposes permission-based authorization middleware.
type Authorizer interface {
	RequirePermissions(permissions ...string) fiber.Handler
	Subject(c fiber.Ctx) string
}
