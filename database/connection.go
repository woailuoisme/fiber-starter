package database

import (
	"fmt"
	"log"
	"time"

	"fiber-starter/config"

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
	dsn := buildDSN(cfg.Database)
	// 创建数据库连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(getLogLevel(cfg.App.Debug)),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("数据库连接成功: %s", cfg.Database.Database)
	
	// 设置全局DB实例
	DB = db

	return &Connection{DB: db}, nil
}

// buildDSN 构建数据库连接字符串
func buildDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		cfg.Host,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.Port,
		cfg.Timezone,
	)
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
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate 自动迁移数据库表
func (c *Connection) AutoMigrate(models ...interface{}) error {
	return c.DB.AutoMigrate(models...)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}