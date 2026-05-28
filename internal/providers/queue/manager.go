package queue

import (
	"errors"
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/queue/contracts"
	"lfiber/internal/providers/queue/drivers"
)

// Manager handles the lifecycle and selection of queue drivers
type Manager struct {
	config  *configs.Config
	drivers map[string]contracts.Queue
	mu      sync.Mutex
}

// NewManager creates a new queue manager instance
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config:  cfg,
		drivers: make(map[string]contracts.Queue),
	}
}

// Drive returns a queue driver instance by name, or the default driver if none specified
func (m *Manager) Drive(name ...string) contracts.Queue {
	m.mu.Lock()
	defer m.mu.Unlock()

	driverName := "asynq" // Default driver
	if len(name) > 0 && name[0] != "" {
		driverName = name[0]
	}

	if driver, ok := m.drivers[driverName]; ok {
		return driver
	}

	var driver contracts.Queue
	switch driverName {
	case "asynq":
		driver = drivers.NewAsynqDriver(m.config)
	case "noop", "null":
		driver = drivers.NewNoopQueue()
	default:
		driver = drivers.NewAsynqDriver(m.config)
	}

	m.drivers[driverName] = driver
	return driver
}

// Close closes all cached queue drivers.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, driver := range m.drivers {
		if err := driver.Close(); err != nil {
			errs = append(errs, errors.New(name+": "+err.Error()))
		}
	}

	return errors.Join(errs...)
}
