package hash

import (
	"lfiber/configs"
	hashContracts "lfiber/internal/providers/hash/contracts"
)

// RegisterHash initializes and returns the Hash manager as a Hasher contract.
func RegisterHash(cfg *configs.Config) (hashContracts.Hasher, error) {
	return NewHashManager(cfg), nil
}
