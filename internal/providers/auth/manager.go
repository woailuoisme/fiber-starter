package auth

import (
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/auth/contracts"
	"lfiber/internal/providers/auth/drivers"
	database "lfiber/internal/providers/database/contracts"
	hashContracts "lfiber/internal/providers/hash/contracts"
)

// Manager manages authentication guards and user providers
type Manager struct {
	cfg          *configs.Config
	db           database.Connection
	hasher       hashContracts.Hasher
	guards       map[string]contracts.Guard
	modelCreator func() any
	mu           sync.RWMutex
}

// NewManager creates a new auth manager instance
func NewManager(cfg *configs.Config, db database.Connection, hasher hashContracts.Hasher) *Manager {
	return &Manager{
		cfg:    cfg,
		db:     db,
		hasher: hasher,
		guards: make(map[string]contracts.Guard),
	}
}

// SetModelCreator sets the model creator function
func (m *Manager) SetModelCreator(creator func() any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelCreator = creator
}

// Guard returns an authentication guard by name
func (m *Manager) Guard(name ...string) contracts.Guard {
	m.mu.Lock()
	defer m.mu.Unlock()

	target := "jwt" // Default guard
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}

	if guard, ok := m.guards[target]; ok {
		return guard
	}

	guard := m.resolve(target)
	m.guards[target] = guard
	return guard
}

// resolve creates the requested guard instance based on configuration
func (m *Manager) resolve(name string) contracts.Guard {
	guardCfg, ok := m.cfg.Auth.Guards[name]
	if !ok {
		// If not found, use a default configuration or return a basic JWT guard
		provider := drivers.NewDatabaseUserProvider(m.db, "users", m.hasher)
		provider.SetModelCreator(m.modelCreator)
		return drivers.NewJWTGuard(provider)
	}

	// Resolve the user provider for this guard
	providerCfg := m.cfg.Auth.Providers[guardCfg.Provider]
	var provider contracts.UserProvider

	switch providerCfg.Driver {
	case "database":
		dbProvider := drivers.NewDatabaseUserProvider(m.db, providerCfg.Table, m.hasher)
		dbProvider.SetModelCreator(m.modelCreator)
		provider = dbProvider
	default:
		dbProvider := drivers.NewDatabaseUserProvider(m.db, "users", m.hasher)
		dbProvider.SetModelCreator(m.modelCreator)
		provider = dbProvider
	}

	// Resolve the guard driver
	switch guardCfg.Driver {
	case "jwt":
		return drivers.NewJWTGuard(provider)
	default:
		return drivers.NewJWTGuard(provider)
	}
}
