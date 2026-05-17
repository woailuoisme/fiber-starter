package storage

import (
	storageContracts "fiber-starter/app/Providers/Storage/Contracts"
	"fiber-starter/configs"
)

// Register initializes the storage manager and returns it
func Register(cfg *configs.Config) (storageContracts.StorageManager, error) {
	return NewManager(cfg), nil
}
