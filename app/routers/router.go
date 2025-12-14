package routers

import (
	"fiber-starter/app/http/controllers"
	"fiber-starter/app/http/middleware"

	_ "fiber-starter/docs" // swagger docs

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// SetupRoutes 配置API路由
func SetupRoutes(app *fiber.App, authController *controllers.AuthController, userController *controllers.UserController, storageController *controllers.StorageController) {
	// 根路径处理
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "欢迎使用 Fiber Starter API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
			"health":  "/health",
			"api":     "/api/v1",
		})
	})

	// 处理 Vite 相关请求（如果有的话）
	app.Get("/@vite/:path*", func(c *fiber.Ctx) error {
		// 这是一个纯后端 API，不提供 Vite 服务
		return c.Status(404).JSON(fiber.Map{
			"error":    "Vite 开发服务器不可用，这是一个纯后端 API 服务",
			"message":  "请使用 API 端点进行交互",
			"api_docs": "/swagger/index.html",
		})
	})

	// Swagger 文档路由
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// API 路由组
	api := app.Group("/api/v1")

	// 认证路由
	auth := api.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.RefreshToken)
	auth.Post("/logout", middleware.JWTProtected(), authController.Logout)
	auth.Post("/change-password", middleware.JWTProtected(), authController.ChangePassword)
	auth.Post("/reset-password", authController.ResetPassword)

	// 用户路由
	users := api.Group("/users")
	users.Get("/", middleware.JWTProtected(), userController.GetUsers)
	users.Get("/me", middleware.JWTProtected(), userController.GetCurrentUser)
	users.Get("/search", middleware.JWTProtected(), userController.SearchUsers)
	users.Put("/:id", middleware.JWTProtected(), userController.UpdateUser)
	users.Delete("/:id", middleware.JWTProtected(), userController.DeleteUser)
	users.Put("/profile", middleware.JWTProtected(), userController.UpdateProfile)

	// 存储路由
	SetupStorageRoutes(api, storageController)
}
