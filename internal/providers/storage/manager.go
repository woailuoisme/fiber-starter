package storage

import (
	"fmt"
	"sync"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/storage/Contracts"
)

// Manager handles multiple storage disks (similar to Laravel's StorageManager)
type Manager struct {
	config *configs.Config
	disks  map[string]Contracts.Disk
	mu     sync.RWMutex
}

// NewManager creates a new storage manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config: cfg,
		disks:  make(map[string]Contracts.Disk),
	}
}

// Disk returns a storage disk by name
func (m *Manager) Disk(name ...string) Contracts.Disk {
	m.mu.Lock()
	defer m.mu.Unlock()

	target := m.config.Storage.Driver
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}

	if disk, ok := m.disks[target]; ok {
		return disk
	}

	// Create new disk if it doesn't exist
	disk := m.createDisk(target)
	m.disks[target] = disk
	return disk
}

// createDisk is a factory method to create the appropriate disk driver
func (m *Manager) createDisk(driver string) Contracts.Disk {
	disk, err := createDisk(driver, m.config)
	if err != nil {
		// In a production app, we might return a NullDriver or error out gracefully
		// For now, we'll follow the pattern of failing fast if a critical disk is missing
		panic(fmt.Errorf("failed to create storage disk [%s]: %w", driver, err))
	}
	return disk
}

// Close all disks
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, disk := range m.disks {
		if err := disk.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close disk %s: %w", name, err)
		}
	}
	return lastErr
}
