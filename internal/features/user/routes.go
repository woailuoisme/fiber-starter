package user

import (
	"time"

	middleware "lfiber/internal/common/middleware"
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

	usersRouter := middleware.NewTimeoutRouter(
		router.Group("/users", middleware.IdempotencyMiddleware()),
		routeTimeout,
	)

	usersRouter.Get("/", jwtProtected, userController.GetUsers)
	usersRouter.Get("/me", jwtProtected, userController.GetCurrentUser)
	usersRouter.Get("/search", jwtProtected, userController.SearchUsers)
	usersRouter.Get("/:id", jwtProtected, userController.GetUser)
	usersRouter.Put("/:id", jwtProtected, userController.UpdateUser)
	usersRouter.Delete("/:id", jwtProtected, userController.DeleteUser)
	usersRouter.Put("/profile", jwtProtected, userController.UpdateProfile)
	usersRouter.Get("/export", jwtProtected, userController.ExportUsers)
	usersRouter.Post("/import", jwtProtected, userController.ImportUsers)
}
