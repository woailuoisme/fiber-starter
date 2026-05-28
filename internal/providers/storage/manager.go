package storage

import (
	"fmt"
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/storage/contracts"
	drivers "lfiber/internal/providers/storage/drivers"
	support "lfiber/internal/support"

	"go.uber.org/zap"
)

// Manager handles multiple storage disks (similar to Laravel's StorageManager)
type Manager struct {
	config *configs.Config
	disks  map[string]contracts.Disk
	mu     sync.RWMutex
}

// NewManager creates a new storage manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config: cfg,
		disks:  make(map[string]contracts.Disk),
	}
}

// Disk returns a storage disk by name
func (m *Manager) Disk(name ...string) contracts.Disk {
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
func (m *Manager) createDisk(driver string) contracts.Disk {
	disk, err := createDisk(driver, m.config)
	if err != nil || disk == nil {
		support.Warn("Failed to create storage disk, falling back to NoopDisk", zap.String("driver", driver), zap.Error(err))
		return drivers.NewNoopDisk()
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
