package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	Services "fiber-starter/app/Services"
	"fiber-starter/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStorageService Test storage service functionality
func TestStorageService(t *testing.T) {
	if os.Getenv("RUN_STORAGE_TESTS") != "1" {
		t.Skip("Skip storage service tests (requires RUN_STORAGE_TESTS=1 with available storage backend)")
	}

	t.Run("RedisStorage", func(t *testing.T) {
		testStorageDriver(t, "redis")
	})
}

// testStorageDriver Test storage functionality for specified driver
func testStorageDriver(t *testing.T, driver string) {
	// Create storage service
	storageService, err := createTestStorageService(t, driver)
	require.NoError(t, err)

	testKey := "test_key_" + driver
	testValue := []byte("test_value_" + driver)

	performStorageOperations(t, storageService, testKey, testValue)
}

// createTestStorageService Create test storage service
func createTestStorageService(t *testing.T, driver string) (*Services.StorageService, error) {
	t.Helper()

	storageCfg := &config.StorageConfig{
		Driver:   driver,
		Database: filepath.Join(t.TempDir(), "test_storage.db"),
		Reset:    true,
	}
	redisCfg := &config.RedisConfig{
		Host:     envOr("REDIS_HOST", "localhost"),
		Port:     envOr("REDIS_PORT", "6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}

	return Services.NewStorageService(storageCfg, redisCfg)
}

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// performStorageOperations Perform storage operation tests
func performStorageOperations(t *testing.T, storageService *Services.StorageService, testKey string, testValue []byte) {
	// Set value
	err := storageService.Set(testKey, testValue, time.Minute)
	require.NoError(t, err)

	// Get value
	value, err := storageService.Get(testKey)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(value, testValue))

	// Check if key exists
	exists, err := storageService.Exists(testKey)
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete key
	err = storageService.Delete(testKey)
	require.NoError(t, err)

	// Check key existence again
	exists, err = storageService.Exists(testKey)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test getting non-existent key
	_, err = storageService.Get("non_existent_key")
	assert.Error(t, err)
}

// TestStorageServiceExpire Test storage expiration functionality
func TestStorageServiceExpire(t *testing.T) {
	if os.Getenv("RUN_STORAGE_TESTS") != "1" {
		t.Skip("Skip storage service expiration tests (requires RUN_STORAGE_TESTS=1)")
	}

	// Create temporary config
	storageCfg := &config.StorageConfig{
		Driver:   "redis",
		Database: filepath.Join(t.TempDir(), "test_storage.db"),
		Reset:    true,
	}
	redisCfg := &config.RedisConfig{
		Host:     envOr("REDIS_HOST", "localhost"),
		Port:     envOr("REDIS_PORT", "6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}

	// Create storage service
	storageService, err := Services.NewStorageService(storageCfg, redisCfg)
	require.NoError(t, err)

	testKey := "expire_test_key"
	testValue := []byte("expire_test_value")

	// Set a short-lived value (1 second)
	err = storageService.Set(testKey, testValue, time.Second)
	require.NoError(t, err)

	// Should exist when getting immediately
	_, err = storageService.Get(testKey)
	require.NoError(t, err)

	// Wait 2 seconds, should expire
	time.Sleep(2 * time.Second)

	_, err = storageService.Get(testKey)
	assert.Error(t, err)
}
