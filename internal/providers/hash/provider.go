package hash

import (
	"fiber-starter/configs"
	hashContracts "fiber-starter/internal/providers/hash/Contracts"
)

// RegisterHash initializes and returns the Hash manager as a Hasher contract.
func RegisterHash(cfg *configs.Config) (hashContracts.Hasher, error) {
	return NewHashManager(cfg), nil
}
