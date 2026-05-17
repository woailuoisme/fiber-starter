package user

import (
	"time"

	"fiber-starter/configs"
	middleware "fiber-starter/internal/common/middleware"
	cache "fiber-starter/internal/providers/cache/contracts"
	database "fiber-starter/internal/providers/database/contracts"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// RegisterRoutes registers user routes under the provided router group.
func RegisterRoutes(router fiber.Router, db database.Connection, cfg *configs.Config, cache cache.Store) {
	userService := NewUserService(db)
	userController := NewUserController(userService)
	jwtProtected := middleware.JWTProtected(cfg, cache)

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
