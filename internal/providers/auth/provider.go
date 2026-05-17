package auth

import (
	"fiber-starter/configs"
	authContracts "fiber-starter/internal/providers/auth/contracts"
	database "fiber-starter/internal/providers/database/contracts"
	hashContracts "fiber-starter/internal/providers/hash/contracts"
)

// Register initializes the Auth manager and returns the manager contract.
func Register(cfg *configs.Config, db database.Connection, hasher hashContracts.Hasher) (authContracts.Manager, error) {
	return NewManager(cfg, db, hasher), nil
}
