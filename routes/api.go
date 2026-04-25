package routes

import (
	"os"
	"path/filepath"

	"fiber-starter/app/Http/Controllers"
	"fiber-starter/app/Http/Middleware"
	"fiber-starter/app/Providers"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupRoutes registers all HTTP routes.
func SetupRoutes(
	app *fiber.App,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	healthController *controllers.HealthController,
) {
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Fiber Starter API",
			"version": "1.0.0",
			"docs":    "/docs",
			"openapi": "/openapi.json",
			"health":  "/health",
			"ready":   "/ready",
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

	app.Get("/health", healthController.Health)
	app.Get("/ready", healthController.Ready)

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.RefreshToken)
	auth.Post("/logout", jwtProtected, authController.Logout)
	auth.Post("/change-password", jwtProtected, authController.ChangePassword)
	auth.Post("/reset-password", authController.ResetPassword)

	users := api.Group("/users")
	users.Get("/", jwtProtected, userController.GetUsers)
	users.Get("/me", jwtProtected, userController.GetCurrentUser)
	users.Get("/search", jwtProtected, userController.SearchUsers)
	users.Put("/:id", jwtProtected, userController.UpdateUser)
	users.Delete("/:id", jwtProtected, userController.DeleteUser)
	users.Put("/profile", jwtProtected, userController.UpdateProfile)
}

// SetupApplicationRoutes binds middleware and routes in one place for Laravel-style bootstrapping.
func SetupApplicationRoutes(app *fiber.App, container *providers.Container) error {
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
		SetupRoutes(app, jwtProtected, authController, userController, healthController)
		if cfg.App.Debug {
			helpers.Info("registered_routes", zap.Int("total", len(app.GetRoutes())))
		}
	})
}

func openAPISpecPath() string {
	path := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	parentPath := filepath.Join("..", "docs", "swagger.json")
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
