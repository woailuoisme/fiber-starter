package user

import (
	"time"

	middleware "lfiber/internal/common/middleware"
	authorization "lfiber/internal/providers/authorization"
	"lfiber/internal/support/appctx"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// RegisterRoutes registers user routes under the provided router group.
func RegisterRoutes(router fiber.Router) {
	rt := appctx.App()

	userService := NewUserService(rt.ConnectionValue())
	userDataExchange := NewUserDataExchange(rt.ConnectionValue())
	userController := NewUserController(userService, userDataExchange)
	jwtProtected := middleware.JWTProtected(rt.AppConfig(), rt.CacheStore())
	requirePermission := authorization.RequirePermissions

	usersRouter := middleware.NewTimeoutRouter(
		router.Group("/users", middleware.IdempotencyMiddleware()),
		routeTimeout,
	)

	usersRouter.Get("/", jwtProtected, requirePermission("users:list"), userController.GetUsers)
	usersRouter.Get("/me", jwtProtected, userController.GetCurrentUser)
	usersRouter.Get("/search", jwtProtected, requirePermission("users:read"), userController.SearchUsers)
	usersRouter.Put("/profile", jwtProtected, userController.UpdateProfile)
	usersRouter.Get("/export", jwtProtected, requirePermission("users:export"), userController.ExportUsers)
	usersRouter.Post("/import", jwtProtected, requirePermission("users:import"), userController.ImportUsers)
	usersRouter.Get("/:id", jwtProtected, requirePermission("users:read"), userController.GetUser)
	usersRouter.Put("/:id", jwtProtected, requirePermission("users:update"), userController.UpdateUser)
	usersRouter.Delete("/:id", jwtProtected, requirePermission("users:delete"), userController.DeleteUser)
}
