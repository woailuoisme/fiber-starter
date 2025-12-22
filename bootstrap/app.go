package bootstrap

import (
	"errors"
	"os/exec"

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

	// 获取 Fiber 配置
	cfg := config.GlobalConfig.App.Fiber

	// 设置并发数，如果配置为 0 则使用默认值
	concurrency := cfg.Concurrency
	if concurrency == 0 {
		concurrency = 256 * 1024 // 默认 256K 并发连接
	}

	// 设置请求体大小限制，如果配置为 0 则使用默认值
	bodyLimit := cfg.BodyLimit
	if bodyLimit == 0 {
		bodyLimit = 4 * 1024 * 1024 // 默认 4MB
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		Prefork:      cfg.Prefork,
		ServerHeader: cfg.ServerHeader,
		BodyLimit:    bodyLimit,
		Concurrency:  concurrency,
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

	// 打印所有路由（开发环境）
	if config.GlobalConfig.App.Debug {
		printRoutes(app)
	}

	// 启动服务器
	port := ":" + config.GetString("app.port")

	// 尝试清理占用的端口
	if err := killPortProcess(config.GetString("app.port")); err != nil {
		helpers.Warn("清理端口进程时出现警告", zap.Error(err))
	}

	helpers.Info("服务器启动在端口 ", zap.String("port", config.GetString("app.port")))
	if err := app.Listen(port); err != nil {
		helpers.Fatal("服务器启动失败", zap.Error(err))
	}
}

// printRoutes 打印所有注册的路由
func printRoutes(app *fiber.App) {
	routes := app.GetRoutes()
	helpers.Info("注册的路由", zap.Int("total", len(routes)))
}

// killPortProcess 清理占用指定端口的进程
func killPortProcess(port string) error {
	// 在 macOS 上使用 lsof 查找占用端口的进程
	cmd := "lsof -ti:" + port + " | xargs kill -9 2>/dev/null || true"

	// 使用 sh -c 执行命令
	execCmd := exec.Command("sh", "-c", cmd)
	if err := execCmd.Run(); err != nil {
		// 忽略错误，因为可能端口没有被占用
		return nil
	}

	return nil
}
