package tests

import (
	"testing"

	database "fiber-starter/database"
	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseConnection Test database connection
func TestDatabaseConnection(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Test if connection is successful
	db, err := conn.GetDB()
	require.NoError(t, err)
	assert.NotNil(t, db)
}

// TestDatabaseHealthCheck Test database health check
func TestDatabaseHealthCheck(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Perform health check
	err = conn.HealthCheck()
	require.NoError(t, err)
}

// TestDatabaseConnectionStats Test getting database connection pool stats
func TestDatabaseConnectionStats(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Get connection pool stats
	stats, err := conn.GetStats()
	require.NoError(t, err)

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
		_, exists := stats[field]
		assert.Truef(t, exists, "stats missing field: %s", field)
	}
}

// TestDatabaseConnectionPool Test database connection pool config
func TestDatabaseConnectionPool(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	// Create database connection
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Get underlying sql.DB
	db, err := conn.GetDB()
	require.NoError(t, err)

	// Verify connection pool config
	stats := db.Stats()

	assert.Equal(t, cfg.Database.Pool.MaxOpenConns, stats.MaxOpenConnections)
}

// TestGlobalHealthCheck Test global health check function
func TestGlobalHealthCheck(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	// Create database connection and set global instance
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	_, _ = conn.GetDB()

	// Test global health check function
	err = database.HealthCheck()
	require.NoError(t, err)
}

// TestGetConnectionStats Test global get connection stats function
func TestGetConnectionStats(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	// Create database connection and set global instance
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	_, _ = conn.GetDB()

	// Test global get connection stats function
	stats, err := database.GetConnectionStats()
	require.NoError(t, err)
	require.NotEmpty(t, stats)
}
