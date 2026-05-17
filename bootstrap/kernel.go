package bootstrap

import (
	"fmt"

	requests "fiber-starter/app/Http/Requests"
	providers "fiber-starter/app/Providers"
	helpers "fiber-starter/app/Support"
	"fiber-starter/configs"
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
