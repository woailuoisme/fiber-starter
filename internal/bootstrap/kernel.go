// Package bootstrap 处理应用程序的初始化和启动流程
package bootstrap

import (
	"context"
	"fmt"

	"lfiber/configs"
	requests "lfiber/internal/common/requests"
	user "lfiber/internal/features/user"
	providers "lfiber/internal/providers"
	helpers "lfiber/internal/support"
	"lfiber/internal/support/otel"
)

// Kernel represents the core of the application
type Kernel struct {
	Config  *configs.Config
	Runtime *providers.Runtime
}

// Boot initializes the core components via Service Providers
func Boot() (*Kernel, func(), error) {
	// Build the runtime (Phase 1: Register)
	rt, err := providers.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build runtime: %w", err)
	}

	// Inject model creator to auth manager to avoid circular dependency
	rt.Auth.SetModelCreator(func() any {
		return &user.User{}
	})

	// Phase 2: Boot (IO connection establishment)
	if err := rt.Boot(); err != nil {
		return nil, nil, fmt.Errorf("failed to boot runtime: %w", err)
	}

	// Inject providers into legacy support packages
	helpers.Init(rt.Log)
	helpers.InitStorage(rt.Storage)
	requests.InitValidator(rt.Validation)

	cleanup := func() {
		_ = rt.Log.Sync()
		_ = rt.Close()
	}

	return &Kernel{
		Config:  rt.Config,
		Runtime: rt,
	}, cleanup, nil
}

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
	if err := SetupApplicationRoutes(app); err != nil {
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
