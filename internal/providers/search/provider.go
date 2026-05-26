package search

import (
	"lfiber/configs"
	searchContracts "lfiber/internal/providers/search/contracts"
)

// Register initializes and returns the search manager and the default engine.
func Register(cfg *configs.Config) (searchContracts.Manager, searchContracts.Engine, error) {
	manager := NewManager(cfg)
	if !cfg.Search.Enabled {
		return manager, manager.Drive("null"), nil
	}
	return manager, manager.Drive(), nil
}
