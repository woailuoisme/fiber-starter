package database

import (
	databaseContracts "fiber-starter/app/Providers/Database/Contracts"
	"fiber-starter/configs"
)

// RegisterDatabase handles the database initialization and wiring.
// Similar to Laravel's DatabaseServiceProvider.
func RegisterDatabase(cfg *configs.Config) (databaseContracts.Manager, databaseContracts.Connection, error) {
	manager := NewManager(cfg)
	return manager, manager.Connection(), nil
}
