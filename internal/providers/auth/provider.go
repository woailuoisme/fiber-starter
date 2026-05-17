package auth

import (
	"fiber-starter/configs"
	authContracts "fiber-starter/internal/providers/auth/Contracts"
	database "fiber-starter/internal/providers/database/Contracts"
	hashContracts "fiber-starter/internal/providers/hash/Contracts"
)

// Register initializes the Auth manager and returns the manager contract.
func Register(cfg *configs.Config, db database.Connection, hasher hashContracts.Hasher) (authContracts.Manager, error) {
	return NewManager(cfg, db, hasher), nil
}
