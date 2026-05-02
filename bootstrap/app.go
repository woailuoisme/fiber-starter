// Package bootstrap 处理应用程序的初始化和启动流程
package bootstrap

import (
	"fmt"

	providers "fiber-starter/app/Providers"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"
)

func App() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	if err := helpers.Init(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		_ = helpers.Sync()
	}()

	runtime, err := providers.Build(cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = runtime.Close()
	}()

	app := createFiberApp(runtime.Config)

	if err := setupAppRoutes(app, runtime); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	return runHTTPServer(app, runtime.Config)
}
