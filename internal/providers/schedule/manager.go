package schedule

import (
	"sync"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/schedule/Contracts"
	"fiber-starter/internal/providers/schedule/Drivers"
)

// Manager handles the lifecycle of the scheduler
type Manager struct {
	config    *configs.Config
	scheduler Contracts.Scheduler
	mu        sync.Mutex
}

// NewManager creates a new schedule manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{config: cfg}
}

// Scheduler returns the default scheduler instance
func (m *Manager) Scheduler() Contracts.Scheduler {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scheduler != nil {
		return m.scheduler
	}

	// Defaulting to Asynq driver
	m.scheduler = Drivers.NewAsynqScheduler(m.config)
	return m.scheduler
}

// Close stops the cached scheduler if it has been created.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scheduler == nil {
		return nil
	}

	return m.scheduler.Stop()
}
