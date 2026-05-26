package hash

import (
	"fmt"
	"sync"

	"lfiber/configs"
	contracts "lfiber/internal/providers/hash/contracts"
	drivers "lfiber/internal/providers/hash/drivers"
)

// Manager handles multiple hashing drivers.
type Manager struct {
	cfg     *configs.Config
	drivers map[string]contracts.Hasher
	mu      sync.RWMutex
}

// NewHashManager creates a new Manager instance.
func NewHashManager(cfg *configs.Config) *Manager {
	return &Manager{
		cfg:     cfg,
		drivers: make(map[string]contracts.Hasher),
	}
}

// Driver returns a hasher instance by driver name.
func (m *Manager) Driver(name string) (contracts.Hasher, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		name = m.cfg.Hash.Driver
	}

	if driver, ok := m.drivers[name]; ok {
		return driver, nil
	}

	var driver contracts.Hasher
	switch name {
	case "bcrypt":
		driver = drivers.NewBcryptHasher(m.cfg.Hash.Bcrypt.Rounds)
	case "argon2":
		driver = drivers.NewArgon2Hasher(
			m.cfg.Hash.Argon2.Memory,
			m.cfg.Hash.Argon2.Iterations,
			m.cfg.Hash.Argon2.Parallelism,
		)
	default:
		return nil, fmt.Errorf("unsupported hash driver: %s", name)
	}

	m.drivers[name] = driver
	return driver, nil
}

// Make creates a hash for the given value using the default driver.
func (m *Manager) Make(value string) (string, error) {
	driver, err := m.Driver("")
	if err != nil {
		return "", err
	}
	return driver.Make(value)
}

// Check verifies the given value against a hash using the default driver.
func (m *Manager) Check(value, hashedValue string) bool {
	driver, err := m.Driver("")
	if err != nil {
		return false
	}
	return driver.Check(value, hashedValue)
}

// NeedsRehash checks if the given hash has been hashed using the default driver's options.
func (m *Manager) NeedsRehash(hashedValue string) bool {
	driver, err := m.Driver("")
	if err != nil {
		return true
	}
	return driver.NeedsRehash(hashedValue)
}

// Info returns information about the given hashed value using the default driver.
func (m *Manager) Info(hashedValue string) map[string]interface{} {
	driver, err := m.Driver("")
	if err != nil {
		return map[string]interface{}{}
	}
	return driver.Info(hashedValue)
}
