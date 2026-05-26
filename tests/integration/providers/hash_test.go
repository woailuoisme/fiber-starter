package providers

import (
	"testing"

	"lfiber/configs"
	hash "lfiber/internal/providers/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashProvider(t *testing.T) {
	cfg := &configs.Config{
		Hash: configs.HashConfig{
			Driver: "bcrypt",
			Bcrypt: configs.BcryptHashConfig{
				Rounds: 10,
			},
		},
	}

	manager := hash.NewHashManager(cfg)
	hasher := manager

	password := "secret-password"

	// Test Bcrypt
	hashed, err := hasher.Make(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.True(t, hasher.Check(password, hashed))
	assert.False(t, hasher.Check("wrong-password", hashed))
	assert.False(t, hasher.NeedsRehash(hashed))

	// Test Argon2
	cfg.Hash.Driver = "argon2"
	cfg.Hash.Argon2 = configs.Argon2HashConfig{
		Memory:      65536,
		Iterations:  4,
		Parallelism: 2,
	}

	argonHasher, err := manager.Driver("argon2")
	require.NoError(t, err)

	hashed2, err := argonHasher.Make(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hashed2)
	assert.True(t, argonHasher.Check(password, hashed2))
	assert.False(t, argonHasher.Check("wrong-password", hashed2))
	assert.False(t, argonHasher.NeedsRehash(hashed2))

	// Test Info
	info := argonHasher.Info(hashed2)
	assert.Equal(t, "argon2", info["algo"])
	assert.Equal(t, uint32(65536), info["options"].(map[string]interface{})["memory"])

	bcryptHasher, _ := manager.Driver("bcrypt")
	infoBcrypt := bcryptHasher.Info(hashed)
	assert.Equal(t, "bcrypt", infoBcrypt["algo"])
	assert.Equal(t, 10, infoBcrypt["options"].(map[string]interface{})["rounds"])

	// Restore default to bcrypt for final tests
	cfg.Hash.Driver = "bcrypt"

	directBcrypt, err := manager.Driver("")
	require.NoError(t, err)
	directHash, err := manager.Make(password)
	require.NoError(t, err)
	assert.True(t, manager.Check(password, directHash))
	assert.False(t, manager.Check("wrong-password", directHash))
	assert.False(t, manager.NeedsRehash(directHash))
	assert.NotEmpty(t, manager.Info(directHash))
	assert.NotNil(t, directBcrypt)
}
