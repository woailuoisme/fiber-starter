package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"fiber-starter/app/controllers"
	"fiber-starter/app/helpers"
	"fiber-starter/app/middleware"
	"fiber-starter/config"
)

func main() {
	// 初始化配置
	config.Init()

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(helpers.ErrorResponse(err.Error(), nil))
		},
	})

	// 中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(helpers.SuccessResponse("服务运行正常", fiber.Map{
			"status": "ok",
		}))
	})

	// API 路由组
	api := app.Group("/api/v1")

	// 认证路由
	auth := api.Group("/auth")
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)
	auth.Post("/refresh", controllers.RefreshToken)
	auth.Post("/logout", middleware.JWTProtected(), controllers.Logout)
	auth.Post("/change-password", middleware.JWTProtected(), controllers.ChangePassword)
	auth.Post("/reset-password", controllers.ResetPassword)

	// 用户路由
	users := api.Group("/users")
	users.Get("/", middleware.JWTProtected(), controllers.GetUsers)
	users.Get("/me", middleware.JWTProtected(), controllers.GetCurrentUser)
	users.Get("/search", middleware.JWTProtected(), controllers.SearchUsers)
	users.Put("/:id", middleware.JWTProtected(), controllers.UpdateUser)
	users.Delete("/:id", middleware.JWTProtected(), controllers.DeleteUser)
	users.Put("/profile", middleware.JWTProtected(), controllers.UpdateProfile)

	// 启动服务器
	port := config.GetString("app.port", ":3000")
	log.Printf("服务器启动在端口 %s", port)
	log.Fatal(app.Listen(port))
}