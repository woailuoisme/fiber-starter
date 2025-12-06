package main

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"fiber-starter/app/controllers"
	"fiber-starter/app/helpers"
	"fiber-starter/app/logger"
	"fiber-starter/app/middleware"
	"fiber-starter/app/providers"
	"fiber-starter/app/routers"
	"fiber-starter/config"

	_ "fiber-starter/docs" // swagger docs
)

func main() {
	// 初始化配置
	err := config.Init()
	if err != nil {
		panic(err)
	}

	// 初始化日志
	if err := logger.Init(); err != nil {
		panic(err)
	}

	// 创建依赖注入容器
	container := providers.NewContainer()

	// 注册所有依赖
	if err := container.RegisterProviders(); err != nil {
		logger.Fatal("注册依赖失败", zap.Error(err))
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		//Prefork: true,
		//Prefork:                 true,                  // 启用多进程模式
		//EnableTrustedProxyCheck: true,                  // 启用信任代理检查
		//TrustedProxies:          []string{"10.0.0.0/8"}, // 信任的代理 IP 段
		//// 性能优化配置
		//Immutable:       true,                         // 上下文不可变，提升性能
		//BodyLimit:       10 * 1024 * 1024,             // 请求体限制 10MB
		//Concurrency:     1024 * 1024,                  // 最大并发连接数
		//ReadBufferSize:  8192,                         // 读取缓冲区 8KB
		//WriteBufferSize: 8192,                         // 写入缓冲区 8KB
		//// 超时配置
		//ReadTimeout:  5 * time.Second,                 // 读取超时 5s
		//WriteTimeout: 10 * time.Second,                // 写入超时 10s
		//IdleTimeout:  60 * time.Second,                // 空闲超时 60s
		//// 路由配置
		//StrictRouting:  true,                         // 严格路由匹配
		//CaseSensitive:  false,                        // 大小写不敏感
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
		logger.Fatal("设置路由失败", zap.Error(err))
	}

	// 启动服务器
	port := ":" + config.GetString("app.port")
	logger.Info("服务器启动在端口 " + port)
	if err := app.Listen(port); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
