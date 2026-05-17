package database

import (
	"fiber-starter/configs"
	databaseContracts "fiber-starter/internal/providers/database/Contracts"
)

// RegisterDatabase handles the database initialization and wiring.
// Similar to Laravel's DatabaseServiceProvider.
func RegisterDatabase(cfg *configs.Config) (databaseContracts.Manager, databaseContracts.Connection, error) {
	manager := NewManager(cfg)
	return manager, manager.Connection(), nil
}
