package database

import (
	"sync"

	"lfiber/configs"
	"lfiber/internal/providers/database/contracts"
)

// Manager handles multiple database connections (similar to Laravel's DatabaseManager)
type Manager struct {
	config      *configs.Config
	connections map[string]contracts.Connection
	mu          sync.RWMutex
}

// NewManager creates a new database manager
func NewManager(cfg *configs.Config) *Manager {
	return &Manager{
		config:      cfg,
		connections: make(map[string]contracts.Connection),
	}
}

// Connection returns a database connection by name
func (m *Manager) Connection(name ...string) contracts.Connection {
	m.mu.Lock()
	defer m.mu.Unlock()

	target := m.GetDefaultConnection()
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}

	if conn, ok := m.connections[target]; ok {
		return conn
	}

	// Create new connection if it doesn't exist
	// NewConnection doesn't connect immediately
	conn := NewConnection(m.config, target)
	conn.manager = m

	m.connections[target] = conn
	return conn
}

// Reconnect closes and re-opens the given connection
func (m *Manager) Reconnect(name ...string) (contracts.Connection, error) {
	target := m.GetDefaultConnection()
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}

	if err := m.Disconnect(target); err != nil {
		return nil, err
	}

	return m.Connection(target), nil
}

// Disconnect closes the given connection
func (m *Manager) Disconnect(name ...string) error {
	m.mu.Lock()
	target := m.GetDefaultConnection()
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}
	conn, ok := m.connections[target]
	m.mu.Unlock()

	if ok {
		return conn.Close()
	}
	return nil
}

// Purge closes and removes the given connection from the manager
func (m *Manager) Purge(name ...string) {
	target := m.GetDefaultConnection()
	if len(name) > 0 && name[0] != "" {
		target = name[0]
	}

	_ = m.Disconnect(target)

	m.mu.Lock()
	delete(m.connections, target)
	m.mu.Unlock()
}

// GetDefaultConnection returns the default connection name
func (m *Manager) GetDefaultConnection() string {
	if m.config.Database.Default != "" {
		return m.config.Database.Default
	}
	return "default"
}

// SetDefaultConnection sets the default connection name
func (m *Manager) SetDefaultConnection(name string) {
	m.config.Database.Default = name
}

// CloseAll closes all open database connections
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		_ = conn.Close()
	}
	m.connections = make(map[string]contracts.Connection)
	return nil
}
