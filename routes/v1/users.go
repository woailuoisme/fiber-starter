package v1

import (
	controllers "fiber-starter/app/Http/Controllers"
	middleware "fiber-starter/app/Http/Middleware"

	"github.com/gofiber/fiber/v3"
)

// registerUserRoutes registers user routes for v1.
func registerUserRoutes(
	router fiber.Router,
	userController *controllers.UserController,
	jwtProtected fiber.Handler,
) {
	usersRouter := middleware.NewTimeoutRouter(
		router.Group("/users", middleware.IdempotencyMiddleware()),
		routeTimeout,
	)

	// Users 组：写操作默认带幂等保护，避免重复提交导致多次更新。
	// 使用方式：列表和查询接口直接使用 jwtProtected；修改类接口共用幂等中间件。
	usersRouter.Get("/", jwtProtected, userController.GetUsers)
	usersRouter.Get("/me", jwtProtected, userController.GetCurrentUser)
	usersRouter.Get("/search", jwtProtected, userController.SearchUsers)
	usersRouter.Put("/:id", jwtProtected, userController.UpdateUser)
	usersRouter.Delete("/:id", jwtProtected, userController.DeleteUser)
	usersRouter.Put("/profile", jwtProtected, userController.UpdateProfile)
}
