package auth

import (
	"lfiber/configs"
	authContracts "lfiber/internal/providers/auth/contracts"
	database "lfiber/internal/providers/database/contracts"
	hashContracts "lfiber/internal/providers/hash/contracts"
)

// Register initializes the Auth manager and returns the manager contract.
func Register(cfg *configs.Config, db database.Connection, hasher hashContracts.Hasher) (authContracts.Manager, error) {
	return NewManager(cfg, db, hasher), nil
}
