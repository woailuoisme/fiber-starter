package bootstrap

import (
	"fiber-starter/internal/app/providers"
	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/transport/http/controllers"
	"fiber-starter/internal/transport/http/middleware"
	"fiber-starter/internal/transport/http/routers"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// setupAppRoutes 配置中间件和路由
func setupAppRoutes(app *fiber.App, container *providers.Container) error {
	middleware.SetupMiddleware(app)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupAuthMiddleware(app)

	return container.Invoke(func(
		cfg *config.Config,
		cache helpers.CacheService,
		authController *controllers.AuthController,
		userController *controllers.UserController,
		healthController *controllers.HealthController,
	) {
		jwtProtected := middleware.JWTProtected(cfg, cache)
		routers.SetupRoutes(app, jwtProtected, authController, userController, healthController)

		if cfg.App.Debug {
			printRoutes(app)
		}
	})
}

// printRoutes prints all registered routes
func printRoutes(app *fiber.App) {
	routes := app.GetRoutes()
	helpers.Info("registered_routes", zap.Int("total", len(routes)))
}
