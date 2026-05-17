package auth

import (
	authContracts "fiber-starter/app/Providers/Auth/Contracts"
	database "fiber-starter/app/Providers/Database/Contracts"
	hashContracts "fiber-starter/app/Providers/Hash/Contracts"
	"fiber-starter/configs"
)

// Register initializes the Auth manager and returns the manager contract.
func Register(cfg *configs.Config, db database.Connection, hasher hashContracts.Hasher) (authContracts.Manager, error) {
	return NewManager(cfg, db, hasher), nil
}
