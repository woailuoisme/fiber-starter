// Package bootstrap 处理应用程序的初始化和启动流程
package bootstrap

import (
	"context"
	"fmt"

	"fiber-starter/app/Support/otel"
)

// ServeHTTP starts the HTTP server using the provided kernel.
// It uses the pre-initialized Runtime from the kernel to setup routes and middleware.
func ServeHTTP(kernel *Kernel) error {
	// Initialize OTEL
	shutdown, err := otel.InitOTEL(kernel.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize otel: %w", err)
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

	// Create Fiber App
	app := NewHTTPApp(kernel.Runtime.Config)

	// Setup Routes using the global application container
	if err := setupAppRoutes(app); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	return Serve(app, kernel.Runtime.Config)
}

// App is a standalone convenience wrapper that boots the kernel and starts the HTTP server.
func App() error {
	kernel, cleanup, err := Boot()
	if err != nil {
		return err
	}
	defer cleanup()

	return ServeHTTP(kernel)
}
