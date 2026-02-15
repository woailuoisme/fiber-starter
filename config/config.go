// Package config 处理应用程序的配置加载和管理
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// MailConfig 邮件配置
type MailConfig struct {
	FromName    string `mapstructure:"from_name"`
	FromAddress string `mapstructure:"from_address"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Encryption  string `mapstructure:"encryption"`
	TLSInsecure bool   `mapstructure:"tls_insecure"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Driver     string              `mapstructure:"driver"`
	Database   string              `mapstructure:"database"`
	Reset      bool                `mapstructure:"reset"`
	GCInterval int                 `mapstructure:"gc_interval"`
	MinIO      *MinIOStorageConfig `mapstructure:"minio"`
	S3         *S3StorageConfig    `mapstructure:"s3"`
	R2         *S3StorageConfig    `mapstructure:"r2"`  // R2 复用 S3 配置
	OSS        *S3StorageConfig    `mapstructure:"oss"` // OSS 复用 S3 配置
}

// MinIOStorageConfig MinIO存储配置
type MinIOStorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
}

// S3StorageConfig AWS S3存储配置
type S3StorageConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	Endpoint        string `mapstructure:"endpoint"` // 可选，用于兼容S3的其他服务
}

// Config 应用程序配置结构体
type Config struct {
	App         AppConfig         `mapstructure:"app"`
	Database    DatabaseConfig    `mapstructure:"database"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Logger      LoggerConfig      `mapstructure:"logger"`
	Cache       CacheConfig       `mapstructure:"cache"`
	Mail        MailConfig        `mapstructure:"mail"`
	Queue       QueueConfig       `mapstructure:"queue"`
	Storage     StorageConfig     `mapstructure:"storage"`
	WebSocket   WebSocketConfig   `mapstructure:"websocket"`
	Payment     PaymentConfig     `mapstructure:"payment"`
	Business    BusinessConfig    `mapstructure:"business"`
	Security    SecurityConfig    `mapstructure:"security"`
	I18n        I18nConfig        `mapstructure:"i18n"`
	Meilisearch MeilisearchConfig `mapstructure:"meilisearch"`
}

// MeilisearchConfig Meilisearch配置
type MeilisearchConfig struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
}

// AppConfig 应用程序基础配置
type AppConfig struct {
	Name     string      `mapstructure:"name"`
	Env      string      `mapstructure:"env"`
	Debug    bool        `mapstructure:"debug"`
	Port     string      `mapstructure:"port"`
	Host     string      `mapstructure:"host"`
	Timezone string      `mapstructure:"timezone"`
	URL      string      `mapstructure:"url"`
	Fiber    FiberConfig `mapstructure:"fiber"`
}

// FiberConfig Fiber 框架配置
type FiberConfig struct {
	Prefork      bool   `mapstructure:"prefork"`
	ServerHeader string `mapstructure:"server_header"`
	BodyLimit    int    `mapstructure:"body_limit"`
	Concurrency  int    `mapstructure:"concurrency"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Default     string                  `mapstructure:"default"`
	Connections map[string]DBConnection `mapstructure:"connections"`
	Pool        DBPoolConfig            `mapstructure:"pool"`
	Read        DBReadConfig            `mapstructure:"read"`
	Write       DBWriteConfig           `mapstructure:"write"`
	Migrations  DBMigrationConfig       `mapstructure:"migrations"`
	Seeders     DBSeederConfig          `mapstructure:"seeders"`
	Redis       DBRedisConfig           `mapstructure:"redis"`
}

// DBConnection 单个数据库连接配置
type DBConnection struct {
	Driver    string            `mapstructure:"driver"`
	Host      string            `mapstructure:"host"`
	Port      string            `mapstructure:"port"`
	Database  string            `mapstructure:"database"`
	Username  string            `mapstructure:"username"`
	Password  string            `mapstructure:"password"`
	Charset   string            `mapstructure:"charset"`
	Collation string            `mapstructure:"collation"`
	Prefix    string            `mapstructure:"prefix"`
	Strict    bool              `mapstructure:"strict"`
	Timezone  string            `mapstructure:"timezone"`
	Schema    string            `mapstructure:"schema"`
	SSLMode   string            `mapstructure:"sslmode"`
	Options   map[string]string `mapstructure:"options"`
}

// DBPoolConfig 数据库连接池配置
type DBPoolConfig struct {
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"`
}

// DBReadConfig 读库配置
type DBReadConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	Port     string   `mapstructure:"port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

// DBWriteConfig 写库配置
type DBWriteConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// DBMigrationConfig 迁移配置
type DBMigrationConfig struct {
	Table string `mapstructure:"table"`
	Path  string `mapstructure:"path"`
}

// DBSeederConfig 填充配置
type DBSeederConfig struct {
	Path string `mapstructure:"path"`
}

// DBRedisConfig Redis配置
type DBRedisConfig struct {
	Client  string                 `mapstructure:"client"`
	Options map[string]interface{} `mapstructure:"options"`
	Default map[string]interface{} `mapstructure:"default"`
}

// JWTConfig JWT认证配置
type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	ExpirationTime int    `mapstructure:"expiration_time"`
	RefreshTime    int    `mapstructure:"refresh_time"`
	ExpireHours    int    `mapstructure:"expire_hours"`
	Issuer         string `mapstructure:"issuer"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Driver  string `mapstructure:"driver"`
	Prefix  string `mapstructure:"prefix"`
	Default int    `mapstructure:"default"`
	TTL     int    `mapstructure:"ttl"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Port              string `mapstructure:"port"`
	Path              string `mapstructure:"path"`
	HeartbeatInterval int    `mapstructure:"heartbeat_interval"`
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	Wechat WechatPaymentConfig `mapstructure:"wechat"`
	Alipay AlipayPaymentConfig `mapstructure:"alipay"`
}

// WechatPaymentConfig 微信支付配置
type WechatPaymentConfig struct {
	AppID     string `mapstructure:"app_id"`
	MchID     string `mapstructure:"mch_id"`
	APIKey    string `mapstructure:"api_key"`
	CertPath  string `mapstructure:"cert_path"`
	KeyPath   string `mapstructure:"key_path"`
	NotifyURL string `mapstructure:"notify_url"`
}

// AlipayPaymentConfig 支付宝配置
type AlipayPaymentConfig struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
	NotifyURL  string `mapstructure:"notify_url"`
}

// BusinessConfig 业务配置
type BusinessConfig struct {
	Order  OrderConfig  `mapstructure:"order"`
	Device DeviceConfig `mapstructure:"device"`
}

// OrderConfig 订单配置
type OrderConfig struct {
	PaymentTimeout int `mapstructure:"payment_timeout"`
	PickupTimeout  int `mapstructure:"pickup_timeout"`
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	ChannelCount       int `mapstructure:"channel_count"`
	ChannelMaxCapacity int `mapstructure:"channel_max_capacity"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
	AllowedMethods string `mapstructure:"allowed_methods"`
	AllowedHeaders string `mapstructure:"allowed_headers"`
}

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	Max    int `mapstructure:"max"`
	Window int `mapstructure:"window"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// loadEnvOnce 确保环境文件只加载一次
var loadEnvOnce sync.Once

// loadEnvFile 根据 APP_ENV 加载对应的 .env 文件
func loadEnvFile() {
	loadEnvOnce.Do(func() {
		// 获取环境变量 APP_ENV，默认为空（使用 .env）
		appEnv := os.Getenv("APP_ENV")

		var envFiles []string

		// 根据 APP_ENV 确定要加载的 .env 文件
		if appEnv != "" {
			// 如果设置了 APP_ENV，优先加载 .env.{APP_ENV}
			envFiles = []string{
				fmt.Sprintf(".env.%s", appEnv),
				".env",
			}
		} else {
			// 如果没有设置 APP_ENV，只加载 .env
			envFiles = []string{".env"}
		}

		// 尝试加载环境文件
		var loaded bool
		for _, envFile := range envFiles {
			// 尝试多个可能的路径
			envPaths := []string{
				envFile,
				fmt.Sprintf("./config/%s", envFile),
				fmt.Sprintf("../%s", envFile),
			}

			for _, path := range envPaths {
				if err := godotenv.Load(path); err == nil {
					log.Printf("成功加载环境文件: %s", path)
					loaded = true
					break
				}
			}

			if loaded {
				break
			}
		}

		if !loaded {
			log.Printf("未找到环境文件，将使用环境变量和默认配置")
		}
	})
}

// LoadDatabaseConfig 加载数据库配置文件
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	configPath := "./config"
	dbConfig := &DatabaseConfig{}

	// 加载环境文件
	loadEnvFile()

	// 设置数据库配置文件路径和名称
	viper.SetConfigName("database")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置环境变量前缀
	viper.SetEnvPrefix("DB")
	viper.AutomaticEnv()

	// 设置默认值
	setDatabaseDefaults()

	// 读取数据库配置文件
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("数据库配置文件未找到，使用默认配置和环境变量")
		} else {
			return nil, err
		}
	}

	// 手动处理环境变量替换
	replaceEnvVars()

	// 解析配置到结构体
	if err := viper.Unmarshal(dbConfig); err != nil {
		return nil, err
	}

	// 检查是否有 DB_CONNECTION 环境变量，如果有则覆盖 default
	if dbConnection := os.Getenv("DB_CONNECTION"); dbConnection != "" {
		dbConfig.Default = dbConnection
	}

	return dbConfig, nil
}

// setDatabaseDefaults 设置数据库默认配置值
func setDatabaseDefaults() {
	defaults := map[string]interface{}{
		// 默认数据库连接
		"default": "mysql",

		// MySQL 默认配置
		"connections.mysql.driver":    "mysql",
		"connections.mysql.host":      "localhost",
		"connections.mysql.port":      "3306",
		"connections.mysql.database":  "fiber_starter",
		"connections.mysql.username":  "root",
		"connections.mysql.password":  "",
		"connections.mysql.charset":   "utf8mb4",
		"connections.mysql.collation": "utf8mb4_unicode_ci",
		"connections.mysql.prefix":    "",
		"connections.mysql.strict":    true,
		"connections.mysql.timezone":  "Local",

		// PostgreSQL 默认配置
		"connections.pgsql.driver":   "postgres",
		"connections.pgsql.host":     "localhost",
		"connections.pgsql.port":     "5432",
		"connections.pgsql.database": "fiber_starter",
		"connections.pgsql.username": "postgres",
		"connections.pgsql.password": "",
		"connections.pgsql.charset":  "utf8",
		"connections.pgsql.prefix":   "",
		"connections.pgsql.schema":   "public",
		"connections.pgsql.sslmode":  "disable",
		"connections.pgsql.timezone": "UTC",

		// SQLite 默认配置
		"connections.sqlite.driver":   "sqlite",
		"connections.sqlite.database": "./database/database.sqlite",
		"connections.sqlite.prefix":   "",

		// 连接池默认配置
		"pool.max_open_conns":     100,
		"pool.max_idle_conns":     10,
		"pool.conn_max_lifetime":  3600,
		"pool.conn_max_idle_time": 600,

		// 迁移默认配置
		"migrations.table": "migrations",
		"migrations.path":  "./database/migrations",

		// 填充默认配置
		"seeders.path": "./database/seeders",
	}

	for key, value := range defaults {
		viper.SetDefault(key, value)
	}
}

// replaceEnvVars 手动处理环境变量替换
func replaceEnvVars() {
	// 获取所有连接配置
	connections := viper.GetStringMap("connections")

	for connName, connConfig := range connections {
		if connMap, ok := connConfig.(map[string]interface{}); ok {
			for key, value := range connMap {
				if valueStr, ok := value.(string); ok {
					if newValue, changed := processEnvVarSubstitution(valueStr); changed {
						viper.Set(fmt.Sprintf("connections.%s.%s", connName, key), newValue)
					}
				}
			}
		}
	}
}

// processEnvVarSubstitution 处理环境变量替换逻辑
func processEnvVarSubstitution(valueStr string) (string, bool) {
	// 检查是否包含环境变量占位符
	if strings.Contains(valueStr, "${") && strings.Contains(valueStr, "}") {
		// 提取环境变量名和默认值
		start := strings.Index(valueStr, "${") + 2
		end := strings.Index(valueStr, "}")
		if start > 1 && end > start {
			envPart := valueStr[start:end]
			parts := strings.SplitN(envPart, ":", 2)
			envKey := parts[0]
			defaultValue := ""
			if len(parts) > 1 {
				defaultValue = parts[1]
			}

			// 获取环境变量值
			envValue := os.Getenv(envKey)
			if envValue == "" {
				envValue = defaultValue
			}

			return envValue, true
		}
	}
	return "", false
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	// 首先加载数据库配置
	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		log.Printf("加载数据库配置失败: %v", err)
		// 如果数据库配置加载失败，使用默认配置
		dbConfig = &DatabaseConfig{}
	}

	config := &Config{
		Database: *dbConfig,
	}

	// 加载环境文件
	loadEnvFile()

	// 设置配置文件路径和名称
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("配置文件未找到，使用默认配置和环境变量")
		} else {
			return nil, err
		}
	}

	// 手动覆盖环境变量
	overrideWithEnv()

	// 解析配置到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	// 设置全局配置
	GlobalConfig = config

	return config, nil
}

// defaultConfigValues 默认配置值
var defaultConfigValues = map[string]interface{}{
	// App配置
	"app.name":     "fiber-starter",
	"app.env":      "development",
	"app.debug":    true,
	"app.port":     "3000",
	"app.host":     "0.0.0.0",
	"app.timezone": "UTC",
	"app.url":      "http://localhost:3000",

	// 日志配置
	"logger.level":       "info",
	"logger.format":      "json",
	"logger.output":      "stdout",
	"logger.max_size":    100,
	"logger.max_age":     30,
	"logger.max_backups": 10,
	"logger.compress":    true,

	// 邮件配置
	"mail.from_name":    "Fiber Starter",
	"mail.from_address": "noreply@example.com",
	"mail.host":         "smtp.example.com",
	"mail.port":         587,
	"mail.username":     "",
	"mail.password":     "",
	"mail.encryption":   "tls",
	"mail.tls_insecure": false,

	// JWT配置
	"jwt.secret":          "your-secret-key-change-in-production",
	"jwt.expiration_time": 3600,   // 1小时
	"jwt.refresh_time":    604800, // 7天
	"jwt.expire_hours":    24,     // 24小时
	"jwt.issuer":          "fiber-starter",

	// Redis配置
	"redis.host":     "localhost",
	"redis.port":     "6379",
	"redis.password": "",
	"redis.db":       0,

	// 缓存配置
	"cache.driver":  "redis",
	"cache.prefix":  "fiber:",
	"cache.default": 3600,
	"cache.ttl":     3600,

	// 队列配置
	"queue.concurrency": 10,

	// 存储配置
	"storage.driver":      "minio",
	"storage.database":    "./storage/storage.db",
	"storage.reset":       false,
	"storage.gc_interval": 10, // 10分钟

	// MinIO配置
	"storage.minio.endpoint":          "localhost:9000",
	"storage.minio.access_key_id":     "minioadmin",
	"storage.minio.secret_access_key": "minioadmin",
	"storage.minio.use_ssl":           false,
	"storage.minio.bucket":            "lunchbox-media",
	"storage.minio.region":            "us-east-1",

	// S3配置
	"storage.s3.region": "us-east-1",
	"storage.s3.bucket": "lunchbox-media",

	// WebSocket配置
	"websocket.port":               "3001",
	"websocket.path":               "/ws",
	"websocket.heartbeat_interval": 30,

	// 支付配置
	"payment.wechat.notify_url": "",
	"payment.alipay.notify_url": "",

	// 业务配置
	"business.order.payment_timeout":       30,
	"business.order.pickup_timeout":        1440,
	"business.device.channel_count":        53,
	"business.device.channel_max_capacity": 4,

	// 安全配置
	"security.cors.allowed_origins": "*",
	"security.cors.allowed_methods": "GET,POST,PUT,DELETE,OPTIONS",
	"security.cors.allowed_headers": "Origin,Content-Type,Accept,Authorization",
	"security.rate_limit.max":       60,
	"security.rate_limit.window":    60,

	// i18n 配置
	"i18n.default_language":    "zh-CN",
	"i18n.supported_languages": []string{"zh-CN", "en"},
	"i18n.language_dir":        "./lang",
	"i18n.cookie_name":         "lang",
	"i18n.cookie_max_age":      31536000, // 1年
}

// envVarMappings 环境变量映射
var envVarMappings = map[string]string{
	// App配置
	"APP_NAME":     "app.name",
	"APP_ENV":      "app.env",
	"APP_DEBUG":    "app.debug",
	"APP_PORT":     "app.port",
	"APP_HOST":     "app.host",
	"APP_TIMEZONE": "app.timezone",
	"APP_URL":      "app.url",

	// Meilisearch配置
	"MEILISEARCH_HOST":    "meilisearch.host",
	"MEILISEARCH_API_KEY": "meilisearch.api_key",

	// Fiber配置
	"FIBER_PREFORK": "app.fiber.prefork",

	// JWT配置
	"JWT_SECRET":          "jwt.secret",
	"JWT_EXPIRATION_TIME": "jwt.expiration_time",
	"JWT_REFRESH_TIME":    "jwt.refresh_time",
	"JWT_ISSUER":          "jwt.issuer",

	// Redis配置
	"REDIS_HOST":     "redis.host",
	"REDIS_PORT":     "redis.port",
	"REDIS_PASSWORD": "redis.password",
	"REDIS_DB":       "redis.db",

	// Logger配置
	"LOG_LEVEL":       "logger.level",
	"LOG_FORMAT":      "logger.format",
	"LOG_OUTPUT":      "logger.output",
	"LOG_MAX_SIZE":    "logger.max_size",
	"LOG_MAX_AGE":     "logger.max_age",
	"LOG_MAX_BACKUPS": "logger.max_backups",
	"LOG_COMPRESS":    "logger.compress",

	// Cache配置
	"CACHE_DRIVER":      "cache.driver",
	"CACHE_PREFIX":      "cache.prefix",
	"CACHE_DEFAULT_TTL": "cache.default",

	// Mail配置
	"MAIL_FROM_NAME":    "mail.from_name",
	"MAIL_FROM_ADDRESS": "mail.from_address",
	"MAIL_HOST":         "mail.host",
	"MAIL_PORT":         "mail.port",
	"MAIL_USERNAME":     "mail.username",
	"MAIL_PASSWORD":     "mail.password",
	"MAIL_ENCRYPTION":   "mail.encryption",
	"MAIL_TLS_INSECURE": "mail.tls_insecure",

	// Queue配置
	"QUEUE_CONCURRENCY": "queue.concurrency",

	// Storage配置
	"STORAGE_DRIVER": "storage.driver",

	// MinIO配置
	"MINIO_ENDPOINT":          "storage.minio.endpoint",
	"MINIO_ACCESS_KEY_ID":     "storage.minio.access_key_id",
	"MINIO_SECRET_ACCESS_KEY": "storage.minio.secret_access_key",
	"MINIO_USE_SSL":           "storage.minio.use_ssl",
	"MINIO_BUCKET":            "storage.minio.bucket",
	"MINIO_REGION":            "storage.minio.region",

	// S3配置
	"S3_ACCESS_KEY_ID":     "storage.s3.access_key_id",
	"S3_SECRET_ACCESS_KEY": "storage.s3.secret_access_key",
	"S3_REGION":            "storage.s3.region",
	"S3_BUCKET":            "storage.s3.bucket",
	"S3_ENDPOINT":          "storage.s3.endpoint",

	// WebSocket配置
	"WEBSOCKET_PORT":               "websocket.port",
	"WEBSOCKET_PATH":               "websocket.path",
	"WEBSOCKET_HEARTBEAT_INTERVAL": "websocket.heartbeat_interval",

	// Payment配置
	"WECHAT_APP_ID":      "payment.wechat.app_id",
	"WECHAT_MCH_ID":      "payment.wechat.mch_id",
	"WECHAT_API_KEY":     "payment.wechat.api_key",
	"WECHAT_CERT_PATH":   "payment.wechat.cert_path",
	"WECHAT_KEY_PATH":    "payment.wechat.key_path",
	"WECHAT_NOTIFY_URL":  "payment.wechat.notify_url",
	"ALIPAY_APP_ID":      "payment.alipay.app_id",
	"ALIPAY_PRIVATE_KEY": "payment.alipay.private_key",
	"ALIPAY_PUBLIC_KEY":  "payment.alipay.public_key",
	"ALIPAY_NOTIFY_URL":  "payment.alipay.notify_url",

	// Business配置
	"ORDER_PAYMENT_TIMEOUT":       "business.order.payment_timeout",
	"ORDER_PICKUP_TIMEOUT":        "business.order.pickup_timeout",
	"DEVICE_CHANNEL_COUNT":        "business.device.channel_count",
	"DEVICE_CHANNEL_MAX_CAPACITY": "business.device.channel_max_capacity",

	// Security配置
	"CORS_ALLOWED_ORIGINS": "security.cors.allowed_origins",
	"CORS_ALLOWED_METHODS": "security.cors.allowed_methods",
	"CORS_ALLOWED_HEADERS": "security.cors.allowed_headers",
	"RATE_LIMIT_MAX":       "security.rate_limit.max",
	"RATE_LIMIT_WINDOW":    "security.rate_limit.window",

	// i18n 配置
	"I18N_DEFAULT_LANGUAGE": "i18n.default_language",
	"I18N_LANGUAGE_DIR":     "i18n.language_dir",
	"I18N_COOKIE_NAME":      "i18n.cookie_name",
	"I18N_COOKIE_MAX_AGE":   "i18n.cookie_max_age",
}

// overrideWithEnv 使用环境变量覆盖配置
func overrideWithEnv() {
	for env, key := range envVarMappings {
		if v := os.Getenv(env); v != "" {
			viper.Set(key, v)
		}
	}

	// 特殊处理 CACHE_DEFAULT_TTL
	if v := os.Getenv("CACHE_DEFAULT_TTL"); v != "" {
		viper.Set("cache.ttl", v)
	}
}

// setDefaults 设置默认配置值
func setDefaults() {
	for key, value := range defaultConfigValues {
		viper.SetDefault(key, value)
	}
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Init 初始化配置
func Init() error {
	var err error
	GlobalConfig, err = LoadConfig()
	if err != nil {
		return err
	}
	return nil
}

// GetString 获取字符串配置
func GetString(key string) string {
	// 直接使用viper获取配置值，这样更可靠
	return viper.GetString(key)
}
