// Package database 处理数据库连接和管理
package database

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fiber-starter/app/helpers"
	"fiber-starter/app/models"
	"fiber-starter/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// Connection 数据库连接结构体
type Connection struct {
	db     *gorm.DB
	config *config.Config
	mu     sync.RWMutex
}

// NewConnection 创建新的数据库连接管理器
func NewConnection(cfg *config.Config) (*Connection, error) {
	// 仅保存配置，不进行实际连接
	return &Connection{
		config: cfg,
	}, nil
}

// GetDB 获取数据库实例（懒加载）
func (c *Connection) GetDB() (*gorm.DB, error) {
	c.mu.RLock()
	if c.db != nil {
		c.mu.RUnlock()
		return c.db, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if c.db != nil {
		return c.db, nil
	}

	// 获取默认连接配置
	defaultConn := c.config.Database.Default
	connConfig, exists := c.config.Database.Connections[defaultConn]
	if !exists {
		return nil, fmt.Errorf("数据库连接配置 '%s' 不存在", defaultConn)
	}

	dsn := buildDSN(connConfig)

	// 创建 GORM DB 实例
	db, err := createGormDB(connConfig.Driver, dsn, c.config.App.Debug)
	if err != nil {
		helpers.LogError("数据库连接失败",
			zap.Error(err),
			zap.String("host", connConfig.Host),
			zap.String("port", connConfig.Port),
			zap.String("database", connConfig.Database))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	if err := configureConnectionPool(db, c.config.Database.Pool); err != nil {
		return nil, err
	}

	helpers.Info("数据库已连接",
		zap.String("database", connConfig.Database),
		zap.String("driver", connConfig.Driver))

	c.db = db
	// 设置全局DB实例（为了兼容性）
	DB = db

	return c.db, nil
}

// createGormDB 根据驱动创建 GORM DB 实例
func createGormDB(driver, dsn string, debug bool) (*gorm.DB, error) {
	logLevel := getLogLevel(debug)
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
	}

	switch strings.ToLower(driver) {
	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Open(dsn), gormConfig)
	case "postgres", "postgresql":
		return gorm.Open(postgres.Open(dsn), gormConfig)
	default:
		// 默认使用SQLite
		return gorm.Open(sqlite.Open(dsn), gormConfig)
	}
}

// configureConnectionPool 配置数据库连接池
func configureConnectionPool(db *gorm.DB, poolConfig config.DBPoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		helpers.LogError("获取底层sql.DB对象失败", zap.Error(err))
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(poolConfig.ConnMaxLifetime) * time.Second)
	if poolConfig.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(poolConfig.ConnMaxIdleTime) * time.Second)
	}

	helpers.Info("数据库连接池配置完成",
		zap.Int("maxIdleConns", poolConfig.MaxIdleConns),
		zap.Int("maxOpenConns", poolConfig.MaxOpenConns),
		zap.Int("connMaxLifetime", poolConfig.ConnMaxLifetime))

	return nil
}

// testConnection 测试数据库连接
func testConnection(db *gorm.DB, connConfig config.DBConnection) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		helpers.LogError("数据库连接测试失败",
			zap.Error(err),
			zap.String("host", connConfig.Host),
			zap.String("port", connConfig.Port),
			zap.String("database", connConfig.Database))
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// buildDSN 构建数据库连接字符串
func buildDSN(cfg config.DBConnection) string {
	// 根据数据库类型构建不同的DSN
	switch strings.ToLower(cfg.Driver) {
	case "sqlite", "sqlite3":
		return cfg.Database
	case "postgres", "postgresql":
		sslmode := cfg.SSLMode
		if sslmode == "" {
			sslmode = "disable"
		}
		timezone := cfg.Timezone
		if timezone == "" {
			timezone = "UTC"
		}
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			cfg.Host,
			cfg.Username,
			cfg.Password,
			cfg.Database,
			cfg.Port,
			sslmode,
			timezone,
		)
	default:
		// 默认使用SQLite
		return cfg.Database
	}
}

// getLogLevel 根据调试模式获取日志级别
func getLogLevel(debug bool) gormLogger.LogLevel {
	if debug {
		return gormLogger.Info
	}
	return gormLogger.Error
}

// Close 关闭数据库连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		helpers.LogError("获取底层sql.DB对象失败，无法关闭连接", zap.Error(err))
		return err
	}

	if err := sqlDB.Close(); err != nil {
		helpers.LogError("关闭数据库连接失败", zap.Error(err))
		return err
	}

	c.db = nil
	helpers.Info("数据库连接已关闭")
	return nil
}

// HealthCheck 检查数据库连接健康状态
func (c *Connection) HealthCheck() error {
	// 尝试获取连接（如果未初始化会尝试连接）
	db, err := c.GetDB()
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		helpers.LogError("获取底层sql.DB对象失败", zap.Error(err))
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 执行 Ping 测试连接
	if err := sqlDB.Ping(); err != nil {
		helpers.LogError("数据库健康检查失败", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	helpers.Info("数据库连接健康检查通过")
	return nil
}

// GetStats 获取数据库连接池统计信息
func (c *Connection) GetStats() (map[string]interface{}, error) {
	db, err := c.GetDB()
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
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

// AutoMigrate 自动迁移数据库表
func (c *Connection) AutoMigrate(models ...interface{}) error {
	db, err := c.GetDB()
	if err != nil {
		return err
	}

	helpers.Info("开始数据库表自动迁移", zap.Int("modelCount", len(models)))

	if err := db.AutoMigrate(models...); err != nil {
		helpers.LogError("数据库表自动迁移失败", zap.Error(err))
		return err
	}

	helpers.Info("数据库表自动迁移完成")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		helpers.LogError("获取底层sql.DB对象失败", zap.Error(err))
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 执行 Ping 测试连接
	if err := sqlDB.Ping(); err != nil {
		helpers.LogError("数据库健康检查失败", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 获取连接池统计信息
	stats := sqlDB.Stats()
	helpers.Info("数据库健康检查通过",
		zap.Int("openConnections", stats.OpenConnections),
		zap.Int("inUse", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("waitCount", stats.WaitCount),
		zap.Duration("waitDuration", stats.WaitDuration),
		zap.Int64("maxIdleClosed", stats.MaxIdleClosed),
		zap.Int64("maxLifetimeClosed", stats.MaxLifetimeClosed),
	)

	return nil
}

// GetConnectionStats 获取数据库连接池统计信息
func GetConnectionStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
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

// AutoMigrate 自动迁移所有数据库表
func AutoMigrate() error {
	if DB == nil {
		helpers.LogError("数据库连接未初始化，无法执行自动迁移")
		return fmt.Errorf("数据库连接未初始化")
	}

	helpers.Info("开始执行全局数据库表自动迁移")

	if err := DB.AutoMigrate(
		&models.User{},
		// 在这里添加其他模型
		// &models.Post{},
		// &models.Comment{},
	); err != nil {
		helpers.LogError("全局数据库表自动迁移失败", zap.Error(err))
		return err
	}

	helpers.Info("全局数据库表自动迁移完成")
	return nil
}
