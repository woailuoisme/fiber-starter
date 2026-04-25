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

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	// 注册 PostgreSQL 驱动
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

// DB 全局数据库实例（仅用于兼容旧代码，推荐通过 Connection 注入使用）
var DB *sql.DB

// Connection 数据库连接结构体
type Connection struct {
	db     *sql.DB
	gormDB *gorm.DB
	config *config.Config
	mu     sync.RWMutex
}

// NewConnection 创建新的数据库连接管理器
func NewConnection(cfg *config.Config) (*Connection, error) {
	return &Connection{config: cfg}, nil
}

// GetDB 获取数据库实例（懒加载）
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

	if c.config == nil {
		return nil, errors.New("database config is nil")
	}

	defaultConn := c.config.Database.Default
	connConfig, exists := c.config.Database.Connections[defaultConn]
	if !exists {
		return nil, fmt.Errorf("database connection config '%s' does not exist", defaultConn)
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

// Dialect 返回当前默认连接对应的 SQL 方言标识（sqlite / psql）。
func (c *Connection) Dialect() (string, error) {
	if c == nil || c.config == nil {
		return "", errors.New("database config is nil")
	}

	defaultConn := c.config.Database.Default
	connConfig, ok := c.config.Database.Connections[defaultConn]
	if !ok {
		return "", fmt.Errorf("database connection config '%s' does not exist", defaultConn)
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

// GetGormDB 获取 GORM 实例（懒加载，复用 GetDB 的连接池）。
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

// Close 关闭数据库连接
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

// HealthCheck Check database connection health status
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

// GetStats 获取数据库连接池统计信息
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
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	return DB
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return errors.New("database connection is not initialized")
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetConnectionStats 获取数据库连接池统计信息
func GetConnectionStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	stats := DB.Stats()
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
	}, nil
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
