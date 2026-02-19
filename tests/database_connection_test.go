package tests

import (
	"path/filepath"
	"testing"

	"fiber-starter/internal/config"
	database "fiber-starter/internal/db"
)

func newTestConfigSQLite(t *testing.T) *config.Config {
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

// TestDatabaseConnection Test database connection
func TestDatabaseConnection(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// Test if connection is successful
	db, err := conn.GetDB()
	if err != nil {
		t.Fatalf("Failed to get database connection: %v", err)
	}
	if db == nil {
		t.Fatal("Database connection is nil")
	}

	t.Log("Database connection successful")
}

// TestDatabaseHealthCheck Test database health check
func TestDatabaseHealthCheck(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// Perform health check
	err = conn.HealthCheck()
	if err != nil {
		t.Fatalf("Database health check failed: %v", err)
	}

	t.Log("Database health check passed")
}

// TestDatabaseConnectionStats Test getting database connection pool stats
func TestDatabaseConnectionStats(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// Get connection pool stats
	stats, err := conn.GetStats()
	if err != nil {
		t.Fatalf("Failed to get connection pool stats: %v", err)
	}

	// Verify stats contain required fields
	requiredFields := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"wait_duration",
	}

	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Stats missing field: %s", field)
		}
	}

	t.Logf("Connection pool stats: %+v", stats)
}

// TestDatabaseConnectionPool Test database connection pool config
func TestDatabaseConnectionPool(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// Get underlying sql.DB
	db, err := conn.GetDB()
	if err != nil {
		t.Fatalf("Failed to get database connection: %v", err)
	}

	// Verify connection pool config
	stats := db.Stats()

	if stats.MaxOpenConnections != cfg.Database.Pool.MaxOpenConns {
		t.Errorf("Max open connections config mismatch: expected %d, got %d",
			cfg.Database.Pool.MaxOpenConns, stats.MaxOpenConnections)
	}

	t.Logf("Connection pool config verified - MaxOpenConns: %d, MaxIdleConns: %d",
		cfg.Database.Pool.MaxOpenConns, cfg.Database.Pool.MaxIdleConns)
}

// TestGlobalHealthCheck Test global health check function
func TestGlobalHealthCheck(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	// Create database connection and set global instance
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	_, _ = conn.GetDB()

	// Test global health check function
	err = database.HealthCheck()
	if err != nil {
		t.Fatalf("Global health check failed: %v", err)
	}

	t.Log("Global health check passed")
}

// TestGetConnectionStats Test global get connection stats function
func TestGetConnectionStats(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	// Create database connection and set global instance
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	_, _ = conn.GetDB()

	// Test global get connection stats function
	stats, err := database.GetConnectionStats()
	if err != nil {
		t.Fatalf("Failed to get connection stats: %v", err)
	}

	if len(stats) == 0 {
		t.Fatal("Connection stats is empty")
	}

	t.Logf("Global connection stats: %+v", stats)
}
