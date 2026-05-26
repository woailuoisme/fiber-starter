package cache

import (
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/cache/contracts"
	drivers "lfiber/internal/providers/cache/drivers"
)

// Manager manages multiple cache stores (similar to Laravel's CacheManager)
type Manager struct {
	config *configs.Config
	stores map[string]contracts.Store
	mu     sync.RWMutex
}

// NewManager creates a new cache manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config: cfg,
		stores: make(map[string]contracts.Store),
	}
}

// Store returns a cache store by name
func (m *Manager) Store(name ...string) contracts.Store {
	m.mu.Lock()
	defer m.mu.Unlock()

	target := m.config.Cache.Driver
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}

	if store, ok := m.stores[target]; ok {
		return store
	}

	// Create new store if it doesn't exist
	var store contracts.Store
	switch target {
	case "redis":
		store = drivers.NewRedisStore(m.config)
	case "memory":
		store = drivers.NewMemoryStore(m.config.Cache.Prefix)
	default:
		// Fallback to memory if unknown
		store = drivers.NewMemoryStore(m.config.Cache.Prefix)
	}

	m.stores[target] = store
	return store
}

// Close all stores
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, store := range m.stores {
		_ = store.Close()
	}
	return nil
}
