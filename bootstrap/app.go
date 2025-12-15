package bootstrap

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"fiber-starter/app/helpers"
	"fiber-starter/app/http/controllers"
	"fiber-starter/app/http/middleware"
	"fiber-starter/app/http/resources"
	"fiber-starter/app/providers"
	"fiber-starter/app/routers"
	"fiber-starter/config"

	_ "fiber-starter/docs"
)

// App 启动应用程序
func App() {
	// 初始化配置
	err := config.Init()
	if err != nil {
		panic(err)
	}

	// 初始化日志
	if err := helpers.Init(); err != nil {
		panic(err)
	}

	// 创建依赖注入容器
	container := providers.NewContainer()

	// 注册所有依赖
	if err := container.RegisterProviders(); err != nil {
		helpers.Fatal("注册依赖失败", zap.Error(err))
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(resources.ErrorResponse(err.Error(), nil))
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
		helpers.Fatal("设置路由失败", zap.Error(err))
	}

	// 启动服务器
	port := ":" + config.GetString("app.port")
	helpers.Info("服务器启动在端口 ", zap.String("port", config.GetString("app.port")))
	if err := app.Listen(port); err != nil {
		helpers.Fatal("服务器启动失败", zap.Error(err))
	}
}
