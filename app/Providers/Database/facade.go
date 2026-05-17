package database

import (
	"errors"

	"fiber-starter/app/Providers/Database/Contracts"
	"fiber-starter/app/Support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// manager returns the database manager instance from the container.
func manager() Contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.DatabaseManager()
	}
	return nil
}

// GetConnection returns a database connection by name
func GetConnection(name ...string) Contracts.Connection {
	if m := manager(); m != nil {
		return m.Connection(name...)
	}
	return nil
}

// Reconnect closes and re-opens the given connection
func Reconnect(name ...string) (Contracts.Connection, error) {
	if m := manager(); m != nil {
		return m.Reconnect(name...)
	}
	return nil, ErrContainerNotInitialized
}

// Disconnect closes the given connection
func Disconnect(name ...string) error {
	if m := manager(); m != nil {
		return m.Disconnect(name...)
	}
	return ErrContainerNotInitialized
}

// Purge closes and removes the given connection from the manager
func Purge(name ...string) {
	if m := manager(); m != nil {
		m.Purge(name...)
	}
}

// GetDefaultConnection returns the default connection name
func GetDefaultConnection() string {
	if m := manager(); m != nil {
		return m.GetDefaultConnection()
	}
	return ""
}

// SetDefaultConnection sets the default connection name
func SetDefaultConnection(name string) {
	if m := manager(); m != nil {
		m.SetDefaultConnection(name)
	}
}

// --- Global Management ---

// CloseAll closes all open database connections
func CloseAll() error {
	if m := manager(); m != nil {
		return m.CloseAll()
	}
	return ErrContainerNotInitialized
}
