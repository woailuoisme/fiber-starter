// Package routers 定义应用程序的HTTP路由
package routers

import (
	_ "fiber-starter/docs" // swagger docs
	"fiber-starter/internal/transport/http/controllers"

	fiberSwagger "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
)

// SetupRoutes 配置API路由
func SetupRoutes(
	app *fiber.App,
	jwtProtected fiber.Handler,
	authController *controllers.AuthController,
	userController *controllers.UserController,
	storageController *controllers.StorageController,
	healthController *controllers.HealthController,
) {
	// 根路径处理
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Fiber Starter API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
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
			"api_docs": "/swagger/index.html",
		})
	})

	// Swagger 文档路由
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)

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

	// 存储路由
	SetupStorageRoutes(api, storageController)
}
