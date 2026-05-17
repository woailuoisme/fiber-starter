package routes

import (
	"strings"

	controllers "fiber-starter/app/Http/Controllers"
	admincontrollers "fiber-starter/app/Http/Controllers/Admin"
	middleware "fiber-starter/app/Http/Middleware"
	httpservices "fiber-starter/app/Http/Services"
	providers "fiber-starter/app/Providers"
	helpers "fiber-starter/app/Support"
	apiRoutes "fiber-starter/routes/api"
	webRoutes "fiber-starter/routes/web"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupRoutes registers all HTTP routes.
func SetupRoutes(
	app *fiber.App,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *admincontrollers.UserController,
	healthController *controllers.HealthController,
) {
	webRoutes.Setup(app, healthController)
	apiRoutes.Setup(app.Group("/api"), jwtProtected, authController, userController)
}

// SetupApplicationRoutes binds middleware and routes in one place for Laravel-style bootstrapping.
func SetupApplicationRoutes(
	app *fiber.App,
) error {
	rt := providers.App()

	authService := httpservices.NewAuthService(rt.Connection, rt.Config, rt.Cache, rt.EmailService, rt.Hash)
	userService := httpservices.NewUserService(rt.Connection)

	authController := controllers.NewAuthController(authService)
	userController := admincontrollers.NewUserController(userService)
	healthController := controllers.NewHealthController(rt.Config)

	jwtProtected := middleware.JWTProtected(rt.Config, rt.Cache)
	middleware.SetupMiddleware(app, rt.Config)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupAuthMiddleware(app)

	SetupRoutes(app, jwtProtected, authController, userController, healthController)
	registerRealtimeRoutes(app, jwtProtected)

	if rt.Config.App.Debug {
		helpers.Info("registered_route_entries", zap.Int("total", len(app.GetRoutes())))
	}
	return nil
}

func registerRealtimeRoutes(app *fiber.App, jwtProtected fiber.Handler) {
	rt := providers.App()
	if rt == nil || rt.Realtime == nil || rt.Config == nil {
		return
	}

	if path := strings.TrimSpace(rt.Config.WebSocket.AuthPath); path != "" {
		app.Post(path, jwtProtected, rt.Realtime.AuthHandler())
	}

	if path := strings.TrimSpace(rt.Config.WebSocket.Path); path != "" {
		app.Get(path, rt.Realtime.Handler())
	}
}
