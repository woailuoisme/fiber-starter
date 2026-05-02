package routes

import (
	"os"
	"path/filepath"

	"fiber-starter/app/Http/Controllers"
	"fiber-starter/app/Http/Middleware"
	"fiber-starter/app/Providers"
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
	app.Get("/", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "Welcome to Fiber Starter API", fiber.Map{
			"version": "1.0.0",
			"docs":    "/docs",
			"openapi": "/openapi.json",
			"health":  "/health",
			"ready":   "/ready",
			"readyz":  "/readyz",
			"api":     "/api/v1",
		})
	})

	app.Get("/@vite/:path*", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":    "Vite dev server is not available. This is an API-only backend.",
			"message":  "Use the API endpoints to interact with the service.",
			"api_docs": "/docs",
		})
	})

	app.Get("/openapi.json", func(c fiber.Ctx) error {
		return c.SendFile(openAPISpecPath())
	})
	app.Get("/docs", scalarDocs)

	// Health Check：给 Docker / K8s / 负载均衡器使用的存活探针。
	// 使用方式：调用 /health 获取基础存活状态，/ready 和 /readyz 获取就绪状态。
	app.Get("/health", healthController.Health)
	app.Get("/ready", healthController.Ready)
	app.Get("/readyz", healthController.Ready)

	api := app.Group("/api")
	v1routes.SetupRoutes(api.Group("/v1"), cfg, jwtProtected, authController, userController)
}

// SetupApplicationRoutes binds middleware and routes in one place for Laravel-style bootstrapping.
func SetupApplicationRoutes(app *fiber.App, container *providers.Container) error {
	return container.Invoke(func(
		cfg *config.Config,
		cache helpers.CacheService,
		authController *controllers.AuthController,
		userController *controllers.UserController,
		healthController *controllers.HealthController,
	) {
		jwtProtected := middleware.JWTProtected(cfg, cache)
		middleware.SetupMiddleware(app, cfg)
		middleware.SetupTimeoutRedirect(app)
		middleware.SetupAuthMiddleware(app)
		SetupRoutes(app, cfg, jwtProtected, authController, userController, healthController)
		if cfg.App.Debug {
			helpers.Info("registered_routes", zap.Int("total", len(app.GetRoutes())))
		}
	})
}

func openAPISpecPath() string {
	path := filepath.Join("docs", "openapi.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	parentPath := filepath.Join("..", "docs", "openapi.json")
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath
	}

	return path
}

func scalarDocs(c fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(`<!doctype html>
<html>
  <head>
    <title>Fiber Starter API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>body{margin:0}</style>
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.json',
        layout: 'modern',
        theme: 'default'
      })
    </script>
  </body>
</html>`)
}
