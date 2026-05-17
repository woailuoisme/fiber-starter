package api

import (
	controllers "fiber-starter/app/Http/Controllers"
	admincontrollers "fiber-starter/app/Http/Controllers/Admin"
	v1 "fiber-starter/routes/api/v1"

	"github.com/gofiber/fiber/v3"
)

// Setup registers API routes.
func Setup(
	router fiber.Router,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *admincontrollers.UserController,
) {
	v1.Setup(router.Group("/v1"), jwtProtected, authController)
	v1.SetupRoutes(router.Group("/v1"), userController, jwtProtected)
}
