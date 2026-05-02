package command

import (
	providers "fiber-starter/app/Providers"
	"fiber-starter/config"
)

func buildRuntime() (*providers.Runtime, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return providers.Build(cfg)
}
