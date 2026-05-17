package v1

import (
	admincontrollers "fiber-starter/app/Http/Controllers/Admin"
	middleware "fiber-starter/app/Http/Middleware"

	"github.com/gofiber/fiber/v3"
)

// registerUserRoutes registers user routes for v1.
func registerUserRoutes(
	router fiber.Router,
	userController *admincontrollers.UserController,
	jwtProtected fiber.Handler,
) {
	usersRouter := middleware.NewTimeoutRouter(
		router.Group("/users", middleware.IdempotencyMiddleware()),
		routeTimeout,
	)

	usersRouter.Get("/", jwtProtected, userController.GetUsers)
	usersRouter.Get("/me", jwtProtected, userController.GetCurrentUser)
	usersRouter.Get("/search", jwtProtected, userController.SearchUsers)
	usersRouter.Put("/:id", jwtProtected, userController.UpdateUser)
	usersRouter.Delete("/:id", jwtProtected, userController.DeleteUser)
	usersRouter.Put("/profile", jwtProtected, userController.UpdateProfile)
	usersRouter.Get("/export", jwtProtected, userController.ExportUsers)
	usersRouter.Post("/import", jwtProtected, userController.ImportUsers)
}
