// Package bootstrap 处理应用程序的初始化和启动流程
package bootstrap

import (
	"fiber-starter/app/Providers"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"go.uber.org/zap"
)

func App() {
	container := providers.NewContainer()

	if err := container.RegisterProviders(); err != nil {
		helpers.Fatal("failed_to_register_providers", zap.Error(err))
	}

	var cfg *config.Config
	if err := container.Invoke(func(c *config.Config) {
		cfg = c
	}); err != nil {
		helpers.Fatal("failed_to_load_config", zap.Error(err))
	}

	if err := helpers.Init(); err != nil {
		helpers.Fatal("failed_to_init_logger", zap.Error(err))
	}

	app := createFiberApp(cfg)

	if err := setupAppRoutes(app, container); err != nil {
		helpers.Fatal("failed_to_setup_routes", zap.Error(err))
	}

	runHTTPServer(app, container, cfg)
}
