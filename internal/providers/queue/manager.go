package queue

import (
	"errors"
	"sync"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/queue/Contracts"
	"fiber-starter/internal/providers/queue/Drivers"
)

// Manager handles the lifecycle and selection of queue drivers
type Manager struct {
	config  *configs.Config
	drivers map[string]Contracts.Queue
	mu      sync.Mutex
}

// NewManager creates a new queue manager instance
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config:  cfg,
		drivers: make(map[string]Contracts.Queue),
	}
}

// Drive returns a queue driver instance by name, or the default driver if none specified
func (m *Manager) Drive(name ...string) Contracts.Queue {
	m.mu.Lock()
	defer m.mu.Unlock()

	driverName := "asynq" // Default driver
	if len(name) > 0 && name[0] != "" {
		driverName = name[0]
	}

	if driver, ok := m.drivers[driverName]; ok {
		return driver
	}

	var driver Contracts.Queue
	switch driverName {
	case "asynq":
		driver = Drivers.NewAsynqDriver(m.config)
	default:
		driver = Drivers.NewAsynqDriver(m.config)
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
