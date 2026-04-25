// Package routers 定义应用程序的HTTP路由
package routers

import (
	"os"
	"path/filepath"

	"fiber-starter/internal/transport/http/controllers"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes 配置API路由
func SetupRoutes(
	app *fiber.App,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	healthController *controllers.HealthController,
) {
	// 根路径处理
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

	// 处理 Vite 相关请求（如果有的话）
	app.Get("/@vite/:path*", func(c fiber.Ctx) error {
		// 这是一个纯后端 API，不提供 Vite 服务
		return c.Status(404).JSON(fiber.Map{
			"error":    "Vite dev server is not available. This is an API-only backend.",
			"message":  "Use the API endpoints to interact with the service.",
			"api_docs": "/docs",
		})
	})

	app.Get("/openapi.json", func(c fiber.Ctx) error {
		return c.SendFile(openAPISpecPath())
	})
	app.Get("/docs", scalarDocs)

	// 健康检查
	app.Get("/health", healthController.Health)
	app.Get("/ready", healthController.Ready)

	// API 路由组
	api := app.Group("/api/v1")

	// 认证路由
	auth := api.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.RefreshToken)
	auth.Post("/logout", jwtProtected, authController.Logout)
	auth.Post("/change-password", jwtProtected, authController.ChangePassword)
	auth.Post("/reset-password", authController.ResetPassword)

	// 用户路由
	users := api.Group("/users")
	users.Get("/", jwtProtected, userController.GetUsers)
	users.Get("/me", jwtProtected, userController.GetCurrentUser)
	users.Get("/search", jwtProtected, userController.SearchUsers)
	users.Put("/:id", jwtProtected, userController.UpdateUser)
	users.Delete("/:id", jwtProtected, userController.DeleteUser)
	users.Put("/profile", jwtProtected, userController.UpdateProfile)
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
