package storage

import (
	"fiber-starter/configs"
	storageContracts "fiber-starter/internal/providers/storage/contracts"
)

// Register initializes the storage manager and returns it
func Register(cfg *configs.Config) (storageContracts.StorageManager, error) {
	return NewManager(cfg), nil
}
