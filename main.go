package main

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"fiber-starter/app/controllers"
	"fiber-starter/app/helpers"
	"fiber-starter/app/middleware"
	"fiber-starter/app/providers"
	"fiber-starter/app/routers"
	"fiber-starter/config"

	_ "fiber-starter/docs" // swagger docs
)

// @title Fiber Starter API
// @version 1.0
// @description 这是一个基于 Fiber 框架的 Go 项目启动模板 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 初始化配置
	err := config.Init()
	if err != nil {
		return
	}

	// 创建依赖注入容器
	container := providers.NewContainer()

	// 注册所有依赖
	if err := container.RegisterProviders(); err != nil {
		log.Fatalf("注册依赖失败: %v", err)
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(helpers.ErrorResponse(err.Error(), nil))
		},
	})

	// 配置中间件
	middleware.SetupMiddleware(app)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupErrorHandling(app)
	middleware.SetupAuthMiddleware(app)

	// 从容器中获取控制器
	err = container.Invoke(func(authController *controllers.AuthController,
		userController *controllers.UserController,
		storageController *controllers.StorageController) {

		// 配置路由
		routers.SetupRoutes(app, authController, userController, storageController)
	})

	if err != nil {
		log.Fatalf("设置路由失败: %v", err)
	}

	// 启动服务器
	port := ":" + config.GetString("app.port")
	log.Printf("服务器启动在端口 %s", port)
	log.Fatal(app.Listen(port))
}
