package search

import (
	"errors"
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/search/contracts"
	"lfiber/internal/providers/search/drivers"
)

// Manager handles the lifecycle and selection of search engines
type Manager struct {
	config  *configs.Config
	engines map[string]contracts.Engine
	mu      sync.Mutex
}

// NewManager creates a new search manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config:  cfg,
		engines: make(map[string]contracts.Engine),
	}
}

// Drive returns a specific search engine instance
func (m *Manager) Drive(name ...string) contracts.Engine {
	m.mu.Lock()
	defer m.mu.Unlock()

	driver := m.config.Search.Default
	if len(name) > 0 && name[0] != "" {
		driver = name[0]
	}

	if engine, ok := m.engines[driver]; ok {
		return engine
	}

	var engine contracts.Engine
	switch driver {
	case "meilisearch":
		engine = drivers.NewMeilisearchDriver(m.config)
	case "null":
		engine = drivers.NewNullDriver()
	default:
		engine = drivers.NewMeilisearchDriver(m.config)
	}

	m.engines[driver] = engine
	return engine
}

// Close closes all cached search engines.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, engine := range m.engines {
		if err := engine.Close(); err != nil {
			errs = append(errs, errors.New(name+": "+err.Error()))
		}
	}

	return errors.Join(errs...)
}
