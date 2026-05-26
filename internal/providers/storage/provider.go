package storage

import (
	"lfiber/configs"
	storageContracts "lfiber/internal/providers/storage/contracts"
)

// Register initializes the storage manager and returns it
func Register(cfg *configs.Config) (storageContracts.StorageManager, error) {
	return NewManager(cfg), nil
}
