// Package bootstrap 处理应用程序的初始化和启动流程
package bootstrap

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"fiber-starter/internal/app/providers"
	"fiber-starter/internal/config"
	database "fiber-starter/internal/db"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/services"
	"fiber-starter/internal/transport/http/controllers"
	"fiber-starter/internal/transport/http/middleware"
	"fiber-starter/internal/transport/http/routers"

	// 引入 swagger 文档
	_ "fiber-starter/docs"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// App 启动应用程序
func App() {
	// 创建依赖注入容器
	container := providers.NewContainer()

	// 注册所有依赖
	if err := container.RegisterProviders(); err != nil {
		helpers.Fatal("failed_to_register_providers", zap.Error(err))
	}

	var cfg *config.Config
	if err := container.Invoke(func(c *config.Config) {
		cfg = c
	}); err != nil {
		helpers.Fatal("failed_to_load_config", zap.Error(err))
	}

	// 初始化日志（依赖 config.GlobalConfig，因此必须在配置加载后）
	if err := helpers.Init(); err != nil {
		helpers.Fatal("failed_to_init_logger", zap.Error(err))
	}

	// 创建 Fiber 应用
	app := createFiberApp(cfg)

	// 配置中间件
	middleware.SetupMiddleware(app)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupAuthMiddleware(app)

	// 从容器中获取控制器
	err := container.Invoke(func(
		cfg *config.Config,
		cache helpers.CacheService,
		authController *controllers.AuthController,
		userController *controllers.UserController,
		storageController *controllers.StorageController,
		healthController *controllers.HealthController,
	) {
		jwtProtected := middleware.JWTProtected(cfg, cache)
		routers.SetupRoutes(app, jwtProtected, authController, userController, storageController, healthController)
	})

	if err != nil {
		helpers.Fatal("failed_to_setup_routes", zap.Error(err))
	}

	// 打印所有路由（开发环境）
	if cfg.App.Debug {
		printRoutes(app)
	}

	// 启动服务器
	port := ":" + cfg.App.Port

	helpers.Info("server_listening", zap.String("port", cfg.App.Port))
	listenErr := make(chan error, 1)
	go func() {
		listenErr <- app.Listen(port, fiber.ListenConfig{
			EnablePrefork:         cfg.App.Fiber.Prefork,
			DisableStartupMessage: true,
		})
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		if err != nil {
			helpers.Fatal("server_failed_to_start", zap.Error(err))
		}
		return
	case <-sigCh:
		helpers.Info("shutdown_signal_received")
	}

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- app.Shutdown()
	}()

	select {
	case err := <-shutdownDone:
		if err != nil {
			helpers.Warn("server_shutdown_failed", zap.Error(err))
		}
	case <-time.After(15 * time.Second):
		helpers.Warn("server_shutdown_timed_out")
	}

	_ = container.Invoke(func(conn *database.Connection, cache helpers.CacheService, queue services.QueueService, storage *services.StorageService) {
		_ = storage.Close()
		_ = queue.Close()
		_ = cache.Close()
		_ = conn.Close()
	})
	_ = helpers.Sync()
}

// createFiberApp 创建并配置 Fiber 应用
func createFiberApp(cfg *config.Config) *fiber.App {
	// 获取 Fiber 配置
	fiberCfg := cfg.App.Fiber

	// 设置并发数，如果配置为 0 则使用默认值
	concurrency := fiberCfg.Concurrency
	if concurrency == 0 {
		concurrency = 256 * 1024 // 默认 256K 并发连接
	}

	// 设置请求体大小限制，如果配置为 0 则使用默认值
	bodyLimit := fiberCfg.BodyLimit
	if bodyLimit == 0 {
		bodyLimit = 4 * 1024 * 1024 // 默认 4MB
	}

	return fiber.New(fiber.Config{
		ServerHeader: fiberCfg.ServerHeader,
		BodyLimit:    bodyLimit,
		Concurrency:  concurrency,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return middleware.HandleError(c, err)
		},
	})
}

// printRoutes prints all registered routes
func printRoutes(app *fiber.App) {
	routes := app.GetRoutes()
	helpers.Info("registered_routes", zap.Int("total", len(routes)))
}
