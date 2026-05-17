package cache

import (
	cacheContracts "fiber-starter/app/Providers/Cache/Contracts"
	"fiber-starter/configs"
)

// RegisterCache initializes and returns the cache manager and the default store.
func RegisterCache(cfg *configs.Config) (cacheContracts.Manager, cacheContracts.Store, error) {
	manager := NewManager(cfg)
	return manager, manager.Store(), nil
}
