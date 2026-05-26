package cache

import (
	"lfiber/configs"
	cacheContracts "lfiber/internal/providers/cache/contracts"
)

// RegisterCache initializes and returns the cache manager and the default store.
func RegisterCache(cfg *configs.Config) (cacheContracts.Manager, cacheContracts.Store, error) {
	manager := NewManager(cfg)
	return manager, manager.Store(), nil
}
