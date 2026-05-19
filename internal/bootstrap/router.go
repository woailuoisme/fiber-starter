package bootstrap

import (
	"strings"

	middleware "fiber-starter/internal/common/middleware"
	"fiber-starter/internal/features/auth"
	"fiber-starter/internal/features/monitoring"
	"fiber-starter/internal/features/user"
	providers "fiber-starter/internal/providers"
	helpers "fiber-starter/internal/support"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupApplicationRoutes binds middleware and routes in one place.
func SetupApplicationRoutes(app *fiber.App) error {
	rt := providers.App()

	// 1. Setup middleware
	middleware.SetupMiddleware(app, rt.Config)

	// 2. Setup public/system routes
	monitoring.RegisterRoutes(app, monitoring.NewHealthController(rt.Config))

	// 3. Register domain routes
	apiGroup := app.Group("/api")
	v1Group := apiGroup.Group("/v1")

	// Register user and auth routes
	auth.RegisterRoutes(v1Group)
	user.RegisterRoutes(v1Group)

	registerRealtimeRoutes(app, middleware.JWTProtected(rt.Config, rt.Cache))

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
