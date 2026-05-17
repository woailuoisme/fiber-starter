package v1

import (
	"time"

	admincontrollers "fiber-starter/app/Http/Controllers/Admin"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// SetupRoutes registers version 1 API routes.
func SetupRoutes(
	router fiber.Router,
	userController *admincontrollers.UserController,
	jwtProtected fiber.Handler,
) {
	registerUserRoutes(router, userController, jwtProtected)
}
