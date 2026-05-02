package v1

import (
	controllers "fiber-starter/app/Http/Controllers"
	middleware "fiber-starter/app/Http/Middleware"
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
)

// registerAuthRoutes registers auth routes for v1.
func registerAuthRoutes(
	router fiber.Router,
	cfg *config.Config,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
) {
	authRouter := middleware.NewTimeoutRouter(
		router.Group("/auth", middleware.AuthLimiter(cfg), middleware.IdempotencyMiddleware()),
		routeTimeout,
	)

	// Auth 组：公开写接口，通常会叠加限流和幂等保护。
	// 使用方式：注册、登录、刷新令牌等接口走该组；需要鉴权的接口再额外挂 jwtProtected。
	authRouter.Post("/register", authController.Register)
	authRouter.Post("/login", authController.Login)
	authRouter.Post("/refresh", authController.RefreshToken)
	authRouter.Post("/logout", jwtProtected, authController.Logout)
	authRouter.Post("/change-password", jwtProtected, authController.ChangePassword)
	authRouter.Post("/reset-password", authController.ResetPassword)
}
