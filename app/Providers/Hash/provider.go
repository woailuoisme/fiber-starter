package hash

import (
	hashContracts "fiber-starter/app/Providers/Hash/Contracts"
	"fiber-starter/configs"
)

// RegisterHash initializes and returns the Hash manager as a Hasher contract.
func RegisterHash(cfg *configs.Config) (hashContracts.Hasher, error) {
	return NewHashManager(cfg), nil
}
