// Package database handles database connections and management
package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/database/contracts"
	drivers "fiber-starter/internal/providers/database/drivers"
	helpers "fiber-starter/internal/support"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"go.uber.org/zap"
)

const (
	driverPostgres   = "postgres"
	driverPostgreSQL = "postgresql"
	driverPgSQL      = "pgsql"
	driverSQLite     = "sqlite"
	driverSQLite3    = "sqlite3"

	dialectPostgres = "postgres"
	dialectSQLite   = "sqlite"
)

type Connection struct {
	name     string
	db       *sql.DB
	bunDB    *bun.DB
	config   *configs.Config
	dbConfig configs.DBConnection // Cached connection-specific config
	manager  *Manager
	mu       sync.RWMutex
	openMu   sync.Mutex
}

var _ contracts.Connection = (*Connection)(nil)

// NewConnection creates a new connection instance (but does not open it immediately)
func NewConnection(cfg *configs.Config, name ...string) *Connection {
	connName := cfg.Database.Default
	if len(name) > 0 && name[0] != "" {
		connName = name[0]
	}

	conn := &Connection{
		name:   connName,
		config: cfg,
	}

	// Pre-cache the specific connection config
	if dbCfg, ok := cfg.Database.Connections[connName]; ok {
		conn.dbConfig = dbCfg
	}

	return conn
}

// GetDB returns the underlying sql.DB instance, opening it if necessary
func (c *Connection) GetDB() (*sql.DB, error) {
	c.mu.RLock()
	if c.db != nil {
		db := c.db
		c.mu.RUnlock()
		return db, nil
	}
	c.mu.RUnlock()

	c.openMu.Lock()
	defer c.openMu.Unlock()

	// Double-check after acquiring lock
	c.mu.RLock()
	if c.db != nil {
		db := c.db
		c.mu.RUnlock()
		return db, nil
	}
	c.mu.RUnlock()

	db, err := c.openWithRetry()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.db = db
	c.mu.Unlock()
	return db, nil
}

// BunDB returns the Bun database wrapper for this connection, ensuring it's connected
func (c *Connection) BunDB() (*bun.DB, error) {
	c.mu.RLock()
	if c.bunDB != nil {
		db := c.bunDB
		c.mu.RUnlock()
		return db, nil
	}
	c.mu.RUnlock()

	// Ensure underlying connection is established
	db, err := c.GetDB()
	if err != nil {
		return nil, fmt.Errorf("open bun database connection: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.bunDB == nil {
		c.bunDB = newBunDB(db, c.GetDriverName())
	}
	if c.bunDB == nil {
		return nil, fmt.Errorf("unsupported database driver: %s", c.GetDriverName())
	}
	return c.bunDB, nil
}

// Dialect returns the database dialect name
func (c *Connection) Dialect() (string, error) {
	switch strings.ToLower(strings.TrimSpace(c.GetDriverName())) {
	case driverPostgres, driverPostgreSQL, driverPgSQL:
		return dialectPostgres, nil
	case driverSQLite, driverSQLite3:
		return dialectSQLite, nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", c.dbConfig.Driver)
	}
}

func (c *Connection) GetName() string {
	return c.name
}

func (c *Connection) GetDriverName() string {
	return c.dbConfig.Driver
}

// Close closes the database connection
func (c *Connection) Close() error {
	c.openMu.Lock()
	defer c.openMu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		return nil
	}

	if err := c.db.Close(); err != nil {
		helpers.LogError("Failed to close database connection", zap.Error(err))
		return err
	}

	c.db = nil
	c.bunDB = nil
	helpers.Info("Database connection closed", zap.String("connection", c.name))
	return nil
}

// HealthCheck verifies the database connection is alive
func (c *Connection) HealthCheck() error {
	db, err := c.GetDB()
	if err != nil {
		return err
	}
	return db.Ping()
}

// GetStats returns database connection statistics
func (c *Connection) GetStats() (map[string]interface{}, error) {
	db, err := c.GetDB()
	if err != nil {
		return nil, err
	}

	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}, nil
}

func newBunDB(db *sql.DB, driver string) *bun.DB {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case driverPostgres, driverPostgreSQL, driverPgSQL:
		return bun.NewDB(db, pgdialect.New())
	case driverSQLite, driverSQLite3:
		return bun.NewDB(db, sqlitedialect.New())
	default:
		return nil
	}
}

func (c *Connection) openWithRetry() (*sql.DB, error) {
	if c.dbConfig.Driver == "" {
		return nil, fmt.Errorf("database connection config '%s' does not exist", c.name)
	}

	connector, err := createConnector(c.dbConfig.Driver)
	if err != nil {
		return nil, err
	}

	attempts := 1
	if c.config.Database.RetryAttempts > 0 {
		attempts = c.config.Database.RetryAttempts
	}
	if c.dbConfig.RetryAttempts != nil {
		attempts = *c.dbConfig.RetryAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		db, connectErr := connector.Connect(c.dbConfig)
		if connectErr == nil {
			configureConnectionPool(db, c.config.Database.Pool)
			if err := db.Ping(); err == nil {
				helpers.Info("Database connection established", zap.String("connection", c.name), zap.Int("attempt", attempt))
				return db, nil
			} else {
				lastErr = err
				_ = db.Close()
			}
		} else {
			lastErr = connectErr
		}

		if attempt < attempts {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil, fmt.Errorf("failed to connect to database '%s' after %d attempts: %w", c.name, attempts, lastErr)
}

func configureConnectionPool(db *sql.DB, poolConfig configs.DBPoolConfig) {
	db.SetMaxIdleConns(poolConfig.MaxIdleConns)
	db.SetMaxOpenConns(poolConfig.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(poolConfig.ConnMaxLifetime) * time.Second)
	if poolConfig.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(poolConfig.ConnMaxIdleTime) * time.Second)
	}
}

func createConnector(driver string) (drivers.Connector, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case driverPostgres, driverPostgreSQL, driverPgSQL:
		return &drivers.PostgresConnector{}, nil
	case driverSQLite, driverSQLite3:
		return &drivers.SQLiteConnector{}, nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}
