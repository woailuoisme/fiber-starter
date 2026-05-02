package routes

import (
	controllers "fiber-starter/app/Http/Controllers"
	middleware "fiber-starter/app/Http/Middleware"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"
	v1routes "fiber-starter/routes/v1"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupRoutes registers all HTTP routes.
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	healthController *controllers.HealthController,
) {
	registerPublicRoutes(app, healthController)
	api := app.Group("/api")
	v1routes.SetupRoutes(api.Group("/v1"), cfg, jwtProtected, authController, userController)
}

// SetupApplicationRoutes binds middleware and routes in one place for Laravel-style bootstrapping.
func SetupApplicationRoutes(
	app *fiber.App,
	cfg *config.Config,
	cache helpers.CacheService,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	healthController *controllers.HealthController,
) error {
	jwtProtected := middleware.JWTProtected(cfg, cache)
	middleware.SetupMiddleware(app, cfg)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupAuthMiddleware(app)
	SetupRoutes(app, cfg, jwtProtected, authController, userController, healthController)
	if cfg.App.Debug {
		helpers.Info("registered_route_entries", zap.Int("total", len(app.GetRoutes())))
	}
	return nil
}
