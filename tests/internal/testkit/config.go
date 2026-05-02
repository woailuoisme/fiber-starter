package testkit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"
)

// NewSQLiteConfig returns a minimal SQLite-backed application config for tests.
func NewSQLiteConfig(t *testing.T) *config.Config {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	return &config.Config{
		Database: config.DatabaseConfig{
			Default: "sqlite",
			Connections: map[string]config.DBConnection{
				"sqlite": {
					Driver:   "sqlite",
					Database: dbPath,
				},
			},
			Pool: config.DBPoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    2,
				ConnMaxLifetime: 60,
				ConnMaxIdleTime: 30,
			},
		},
	}
}

// DefaultRedisConfig returns a default Redis config for tests.
func DefaultRedisConfig() *config.RedisConfig {
	return &config.RedisConfig{
		Host: "localhost",
		Port: "6379",
		DB:   0,
	}
}

// SetLoggerConfig swaps the global logger config for a test and restores it afterward.
func SetLoggerConfig(t *testing.T, loggerCfg config.LoggerConfig) {
	t.Helper()

	prevConfig := config.GlobalConfig
	prevLogger := helpers.Logger
	config.GlobalConfig = &config.Config{Logger: loggerCfg}

	t.Cleanup(func() {
		helpers.Logger = prevLogger
		config.GlobalConfig = prevConfig
	})
}

// UseTempWorkDir changes the current working directory to a temporary directory.
func UseTempWorkDir(t *testing.T) string {
	t.Helper()

	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(old)
	})

	return dir
}

// RepoRoot returns the repository root directory.
func RepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
