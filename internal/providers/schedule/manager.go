package schedule

import (
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/schedule/contracts"
	"lfiber/internal/providers/schedule/drivers"
)

// Manager handles the lifecycle of the scheduler
type Manager struct {
	config    *configs.Config
	scheduler contracts.Scheduler
	mu        sync.Mutex
}

// NewManager creates a new schedule manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{config: cfg}
}

// Scheduler returns the default scheduler instance.
func (m *Manager) Scheduler() (contracts.Scheduler, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scheduler != nil {
		return m.scheduler, nil
	}

	// Defaulting to Asynq driver
	scheduler, err := drivers.NewAsynqScheduler(m.config)
	if err != nil {
		return nil, err
	}

	m.scheduler = scheduler
	return m.scheduler, nil
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
