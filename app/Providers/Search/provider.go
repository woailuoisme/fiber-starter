package search

import (
	searchContracts "fiber-starter/app/Providers/Search/Contracts"
	"fiber-starter/configs"
)

// Register initializes and returns the search manager and the default engine.
func Register(cfg *configs.Config) (searchContracts.Manager, searchContracts.Engine, error) {
	manager := NewManager(cfg)
	if !cfg.Search.Enabled {
		return manager, manager.Drive("null"), nil
	}
	return manager, manager.Drive(), nil
}
