package v1

import (
	"fiber-starter/app/Http/Controllers"
	"fiber-starter/app/Http/Middleware"
	"fiber-starter/config"
	"time"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers version 1 API routes.
func SetupRoutes(
	router fiber.Router,
	cfg *config.Config,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *controllers.UserController,
) {
	authRouter := middleware.NewTimeoutRouter(router.Group("/auth", middleware.AuthLimiter(cfg), middleware.IdempotencyMiddleware()), 30*time.Second)
	// Auth 组：公开写接口，通常会叠加限流和幂等保护。
	// 使用方式：注册、登录、刷新令牌等接口走该组；需要鉴权的接口再额外挂 jwtProtected。
	authRouter.Post("/register", authController.Register)
	authRouter.Post("/login", authController.Login)
	authRouter.Post("/refresh", authController.RefreshToken)
	authRouter.Post("/logout", jwtProtected, authController.Logout)
	authRouter.Post("/change-password", jwtProtected, authController.ChangePassword)
	authRouter.Post("/reset-password", authController.ResetPassword)

	usersRouter := middleware.NewTimeoutRouter(router.Group("/users", middleware.IdempotencyMiddleware()), 30*time.Second)
	// Users 组：写操作默认带幂等保护，避免重复提交导致多次更新。
	// 使用方式：列表和查询接口直接使用 jwtProtected；修改类接口共用幂等中间件。
	usersRouter.Get("/", jwtProtected, userController.GetUsers)
	usersRouter.Get("/me", jwtProtected, userController.GetCurrentUser)
	usersRouter.Get("/search", jwtProtected, userController.SearchUsers)
	usersRouter.Put("/:id", jwtProtected, userController.UpdateUser)
	usersRouter.Delete("/:id", jwtProtected, userController.DeleteUser)
	usersRouter.Put("/profile", jwtProtected, userController.UpdateProfile)
}
