// Package bootstrap 处理应用程序的初始化和启动流程
package bootstrap

import (
	"fiber-starter/internal/app/providers"
	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

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

	// 配置中间件和路由
	if err := setupAppRoutes(app, container); err != nil {
		helpers.Fatal("failed_to_setup_routes", zap.Error(err))
	}

	// 启动并管理服务器生命周期
	runHTTPServer(app, container, cfg)
}
