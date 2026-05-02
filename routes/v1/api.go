package v1

import (
	"time"

	controllers "fiber-starter/app/Http/Controllers"
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// SetupRoutes registers version 1 API routes.
func SetupRoutes(
	router fiber.Router,
	cfg *config.Config,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *controllers.UserController,
) {
	registerAuthRoutes(router, cfg, jwtProtected, authController)
	registerUserRoutes(router, userController, jwtProtected)
}
