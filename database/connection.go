package database

import (
	"fmt"
	"strings"
	"time"

	"fiber-starter/app/logger"
	"fiber-starter/app/models"
	"fiber-starter/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// Connection 数据库连接结构体
type Connection struct {
	*gorm.DB
}

// NewConnection 创建新的数据库连接
func NewConnection(cfg *config.Config) (*Connection, error) {
	// 获取默认连接配置
	defaultConn := cfg.Database.Default
	connConfig, exists := cfg.Database.Connections[defaultConn]
	if !exists {
		return nil, fmt.Errorf("数据库连接配置 '%s' 不存在", defaultConn)
	}

	dsn := buildDSN(connConfig)

	var db *gorm.DB
	var err error

	// 根据数据库类型选择驱动
	switch strings.ToLower(connConfig.Driver) {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(getLogLevel(cfg.App.Debug)),
		})
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(getLogLevel(cfg.App.Debug)),
		})
	default:
		// 默认使用MySQL
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(getLogLevel(cfg.App.Debug)),
		})
	}

	if err != nil {
		logger.Error("数据库连接失败",
			zap.Error(err),
			zap.String("host", connConfig.Host),
			zap.String("port", connConfig.Port),
			zap.String("database", connConfig.Database))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("获取底层sql.DB对象失败", zap.Error(err))
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.Pool.MaxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetMaxOpenConns(cfg.Database.Pool.MaxOpenConns)                                    // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.Pool.ConnMaxLifetime) * time.Second) // 连接最大生存时间
	if cfg.Database.Pool.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.Pool.ConnMaxIdleTime) * time.Second)
	}

	logger.Info("数据库连接池配置完成",
		zap.Int("maxIdleConns", cfg.Database.Pool.MaxIdleConns),
		zap.Int("maxOpenConns", cfg.Database.Pool.MaxOpenConns),
		zap.Int("connMaxLifetime", cfg.Database.Pool.ConnMaxLifetime))

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		logger.Error("数据库连接测试失败",
			zap.Error(err),
			zap.String("host", connConfig.Host),
			zap.String("port", connConfig.Port),
			zap.String("database", connConfig.Database))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("数据库连接成功",
		zap.String("database", connConfig.Database),
		zap.String("driver", connConfig.Driver))

	// 设置全局DB实例
	DB = db

	return &Connection{DB: db}, nil
}

// buildDSN 构建数据库连接字符串
func buildDSN(cfg config.DBConnection) string {
	// 根据数据库类型构建不同的DSN
	switch strings.ToLower(cfg.Driver) {
	case "mysql":
		charset := cfg.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		timezone := cfg.Timezone
		if timezone == "" {
			timezone = "Local"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			charset,
			timezone,
		)
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
		// 默认使用MySQL格式
		charset := cfg.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		timezone := cfg.Timezone
		if timezone == "" {
			timezone = "Local"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			charset,
			timezone,
		)
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
	sqlDB, err := c.DB.DB()
	if err != nil {
		logger.Error("获取底层sql.DB对象失败，无法关闭连接", zap.Error(err))
		return err
	}

	if err := sqlDB.Close(); err != nil {
		logger.Error("关闭数据库连接失败", zap.Error(err))
		return err
	}

	logger.Info("数据库连接已关闭")
	return nil
}

// HealthCheck 检查数据库连接健康状态
func (c *Connection) HealthCheck() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		logger.Error("获取底层sql.DB对象失败", zap.Error(err))
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 执行 Ping 测试连接
	if err := sqlDB.Ping(); err != nil {
		logger.Error("数据库健康检查失败", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("数据库连接健康检查通过")
	return nil
}

// GetStats 获取数据库连接池统计信息
func (c *Connection) GetStats() (map[string]interface{}, error) {
	sqlDB, err := c.DB.DB()
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
	logger.Info("开始数据库表自动迁移", zap.Int("modelCount", len(models)))

	if err := c.DB.AutoMigrate(models...); err != nil {
		logger.Error("数据库表自动迁移失败", zap.Error(err))
		return err
	}

	logger.Info("数据库表自动迁移完成")
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
		logger.Error("获取底层sql.DB对象失败", zap.Error(err))
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 执行 Ping 测试连接
	if err := sqlDB.Ping(); err != nil {
		logger.Error("数据库健康检查失败", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 获取连接池统计信息
	stats := sqlDB.Stats()
	logger.Info("数据库健康检查通过",
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
		logger.Error("数据库连接未初始化，无法执行自动迁移")
		return fmt.Errorf("数据库连接未初始化")
	}

	logger.Info("开始执行全局数据库表自动迁移")

	if err := DB.AutoMigrate(
		&models.User{},
		// 在这里添加其他模型
		// &models.Post{},
		// &models.Comment{},
	); err != nil {
		logger.Error("全局数据库表自动迁移失败", zap.Error(err))
		return err
	}

	logger.Info("全局数据库表自动迁移完成")
	return nil
}
