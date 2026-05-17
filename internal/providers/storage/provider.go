package storage

import (
	"fiber-starter/configs"
	storageContracts "fiber-starter/internal/providers/storage/Contracts"
)

// Register initializes the storage manager and returns it
func Register(cfg *configs.Config) (storageContracts.StorageManager, error) {
	return NewManager(cfg), nil
}
