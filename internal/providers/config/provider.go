package config

import (
	configContracts "fiber-starter/internal/providers/config/Contracts"

	"github.com/knadh/koanf/v2"
)

// RegisterConfig initializes the Config repository.
// RegisterConfig initializes and returns the configuration repository as a contract.
func RegisterConfig(k *koanf.Koanf) (configContracts.Repository, error) {
	return NewRepository(k), nil
}
