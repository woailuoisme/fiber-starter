package contracts

import (
	"database/sql"

	"github.com/uptrace/bun"
)

// Connection defines the contract for a database connection
type Connection interface {
	// --- Core Instances ---

	// GetDB returns the underlying sql.DB instance
	GetDB() (*sql.DB, error)
	// BunDB returns the Bun database wrapper, opening the connection on first use.
	BunDB() (*bun.DB, error)
	// Dialect returns the database dialect (e.g., "sqlite", "psql")
	Dialect() (string, error)

	// --- Metadata & Status ---

	// GetName returns the connection name
	GetName() string
	// GetDriverName returns the driver name
	GetDriverName() string
	// HealthCheck verifies the database connection is alive
	HealthCheck() error
	// GetStats returns database connection statistics
	GetStats() (map[string]interface{}, error)
	// Close closes the database connection
	Close() error
}

// Manager defines the contract for database connection management
type Manager interface {
	// --- Connection Management ---

	// Connection returns a database connection by name
	Connection(name ...string) Connection
	// Reconnect closes and re-opens the given connection
	Reconnect(name ...string) (Connection, error)
	// Disconnect closes the given connection
	Disconnect(name ...string) error
	// Purge closes and removes the given connection from the manager
	Purge(name ...string)

	// --- Default Connection ---

	// GetDefaultConnection returns the default connection name
	GetDefaultConnection() string
	// SetDefaultConnection sets the default connection name
	SetDefaultConnection(name string)

	// --- Global Management ---

	// CloseAll closes all open database connections
	CloseAll() error
}
