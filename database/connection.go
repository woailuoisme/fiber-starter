// Package database 处理数据库连接和管理
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	driverSQLite     = "sqlite"
	driverSQLite3    = "sqlite3"
	driverPostgres   = "postgres"
	driverPostgreSQL = "postgresql"
	dialectSQLite    = "sqlite"
	dialectPostgres  = "psql"
)

var DB *sql.DB

type Connection struct {
	db     *sql.DB
	gormDB *gorm.DB
	config *config.Config
	mu     sync.RWMutex
}

func NewConnection(cfg *config.Config) (*Connection, error) {
	return &Connection{config: cfg}, nil
}

func (c *Connection) GetDB() (*sql.DB, error) {
	c.mu.RLock()
	if c.db != nil {
		db := c.db
		c.mu.RUnlock()
		return db, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		return c.db, nil
	}

	connConfig, err := c.configuredConnection()
	if err != nil {
		return nil, err
	}

	driverName, dsn, err := buildSQLDriverAndDSN(connConfig)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		helpers.LogError("Database connection failed",
			zap.Error(err),
			zap.String("driver", connConfig.Driver),
			zap.String("database", connConfig.Database))
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	configureConnectionPool(db, c.config.Database.Pool)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		helpers.LogError("Database connection test failed",
			zap.Error(err),
			zap.String("driver", connConfig.Driver),
			zap.String("database", connConfig.Database))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	helpers.Info("Database connected",
		zap.String("driver", connConfig.Driver),
		zap.String("database", connConfig.Database))

	c.db = db
	DB = db

	return db, nil
}

func (c *Connection) Dialect() (string, error) {
	connConfig, err := c.configuredConnection()
	if err != nil {
		return "", err
	}

	switch strings.ToLower(strings.TrimSpace(connConfig.Driver)) {
	case driverSQLite, driverSQLite3:
		return dialectSQLite, nil
	case driverPostgres, driverPostgreSQL:
		return dialectPostgres, nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", connConfig.Driver)
	}
}

func (c *Connection) GetGormDB() (*gorm.DB, error) {
	c.mu.RLock()
	if c.gormDB != nil {
		db := c.gormDB
		c.mu.RUnlock()
		return db, nil
	}
	c.mu.RUnlock()

	sqlDB, err := c.GetDB()
	if err != nil {
		return nil, err
	}

	dialect, err := c.Dialect()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.gormDB != nil {
		return c.gormDB, nil
	}

	var gormDB *gorm.DB
	switch dialect {
	case dialectPostgres:
		gormDB, err = gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	case dialectSQLite:
		gormDB, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	default:
		err = fmt.Errorf("unsupported database dialect: %s", dialect)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gorm: %w", err)
	}

	c.gormDB = gormDB
	return gormDB, nil
}

func (c *Connection) Close() error {
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
	c.gormDB = nil
	helpers.Info("Database connection closed")
	return nil
}

func (c *Connection) HealthCheck() error {
	db, err := c.GetDB()
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		helpers.LogError("Database health check failed", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func (c *Connection) GetStats() (map[string]interface{}, error) {
	db, err := c.GetDB()
	if err != nil {
		return nil, err
	}

	return connectionStats(db), nil
}

func GetDB() *sql.DB {
	return DB
}

func HealthCheck() error {
	db, err := globalDB()
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func GetConnectionStats() (map[string]interface{}, error) {
	db, err := globalDB()
	if err != nil {
		return nil, err
	}

	return connectionStats(db), nil
}

func (c *Connection) configuredConnection() (config.DBConnection, error) {
	if c == nil || c.config == nil {
		return config.DBConnection{}, errors.New("database config is nil")
	}

	defaultConn := c.config.Database.Default
	connConfig, exists := c.config.Database.Connections[defaultConn]
	if !exists {
		return config.DBConnection{}, fmt.Errorf("database connection config '%s' does not exist", defaultConn)
	}

	return connConfig, nil
}

func globalDB() (*sql.DB, error) {
	if DB == nil {
		return nil, errors.New("database connection is not initialized")
	}
	return DB, nil
}

func connectionStats(db *sql.DB) map[string]interface{} {
	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

func configureConnectionPool(db *sql.DB, poolConfig config.DBPoolConfig) {
	db.SetMaxIdleConns(poolConfig.MaxIdleConns)
	db.SetMaxOpenConns(poolConfig.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(poolConfig.ConnMaxLifetime) * time.Second)
	if poolConfig.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(poolConfig.ConnMaxIdleTime) * time.Second)
	}
}

func buildSQLDriverAndDSN(cfg config.DBConnection) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case driverSQLite, driverSQLite3:
		path := strings.TrimSpace(cfg.Database)
		if path == "" {
			path = ":memory:"
		}

		if path != ":memory:" && !strings.HasPrefix(path, "file:") {
			if !filepath.IsAbs(path) {
				path = filepath.Clean(path)
			}
		}

		if strings.HasPrefix(path, "file:") {
			return "sqlite3", path, nil
		}

		u := url.URL{Scheme: "file", Path: path}
		q := u.Query()
		q.Set("_foreign_keys", "1")
		q.Set("_busy_timeout", "5000")
		u.RawQuery = q.Encode()
		return "sqlite3", u.String(), nil
	case driverPostgres, driverPostgreSQL:
		host := strings.TrimSpace(cfg.Host)
		port := strings.TrimSpace(cfg.Port)
		user := strings.TrimSpace(cfg.Username)
		pass := cfg.Password
		dbname := strings.TrimSpace(cfg.Database)

		if host == "" {
			host = "127.0.0.1"
		}
		if port == "" {
			port = "5432"
		}
		if user == "" {
			user = "postgres"
		}
		if dbname == "" {
			return "", "", errors.New("postgres database name is empty")
		}

		sslmode := strings.TrimSpace(cfg.SSLMode)
		if sslmode == "" {
			sslmode = "disable"
		}

		u := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(user, pass),
			Host:   fmt.Sprintf("%s:%s", host, port),
			Path:   dbname,
		}
		q := u.Query()
		q.Set("sslmode", sslmode)
		u.RawQuery = q.Encode()
		return "pgx", u.String(), nil
	default:
		return "", "", fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}
