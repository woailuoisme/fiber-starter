package config

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Mail      MailConfig      `mapstructure:"mail"`
	Queue     QueueConfig     `mapstructure:"queue"`
	Storage   StorageConfig   `mapstructure:"storage"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	Payment   PaymentConfig   `mapstructure:"payment"`
	Business  BusinessConfig  `mapstructure:"business"`
	Security  SecurityConfig  `mapstructure:"security"`
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

// 全局配置实例
var GlobalConfig *Config

// LoadDatabaseConfig 加载数据库配置文件
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	configPath := "./config"
	dbConfig := &DatabaseConfig{}

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// 尝试加载其他可能的 .env 文件位置
		envPaths := []string{".env", "./config/.env", "../.env"}
		for _, path := range envPaths {
			if err := godotenv.Load(path); err == nil {
				log.Printf("成功加载 .env 文件: %s", path)
				break
			}
		}
		if err != nil {
			log.Printf("未找到 .env 文件，将使用环境变量和默认配置")
		}
	}

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
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
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

	return dbConfig, nil
}

// setDatabaseDefaults 设置数据库默认配置值
func setDatabaseDefaults() {
	// 默认数据库连接
	viper.SetDefault("default", "mysql")

	// MySQL 默认配置
	viper.SetDefault("connections.mysql.driver", "mysql")
	viper.SetDefault("connections.mysql.host", "localhost")
	viper.SetDefault("connections.mysql.port", "3306")
	viper.SetDefault("connections.mysql.database", "fiber_starter")
	viper.SetDefault("connections.mysql.username", "root")
	viper.SetDefault("connections.mysql.password", "")
	viper.SetDefault("connections.mysql.charset", "utf8mb4")
	viper.SetDefault("connections.mysql.collation", "utf8mb4_unicode_ci")
	viper.SetDefault("connections.mysql.prefix", "")
	viper.SetDefault("connections.mysql.strict", true)
	viper.SetDefault("connections.mysql.timezone", "Local")

	// PostgreSQL 默认配置
	viper.SetDefault("connections.pgsql.driver", "postgres")
	viper.SetDefault("connections.pgsql.host", "localhost")
	viper.SetDefault("connections.pgsql.port", "5432")
	viper.SetDefault("connections.pgsql.database", "fiber_starter")
	viper.SetDefault("connections.pgsql.username", "postgres")
	viper.SetDefault("connections.pgsql.password", "")
	viper.SetDefault("connections.pgsql.charset", "utf8")
	viper.SetDefault("connections.pgsql.prefix", "")
	viper.SetDefault("connections.pgsql.schema", "public")
	viper.SetDefault("connections.pgsql.sslmode", "disable")
	viper.SetDefault("connections.pgsql.timezone", "UTC")

	// SQLite 默认配置
	viper.SetDefault("connections.sqlite.driver", "sqlite")
	viper.SetDefault("connections.sqlite.database", "./database/database.sqlite")
	viper.SetDefault("connections.sqlite.prefix", "")

	// 连接池默认配置
	viper.SetDefault("pool.max_open_conns", 100)
	viper.SetDefault("pool.max_idle_conns", 10)
	viper.SetDefault("pool.conn_max_lifetime", 3600)
	viper.SetDefault("pool.conn_max_idle_time", 600)

	// 迁移默认配置
	viper.SetDefault("migrations.table", "migrations")
	viper.SetDefault("migrations.path", "./database/migrations")

	// 填充默认配置
	viper.SetDefault("seeders.path", "./database/seeders")
}

// replaceEnvVars 手动处理环境变量替换
func replaceEnvVars() {
	// 获取所有连接配置
	connections := viper.GetStringMap("connections")

	for connName, connConfig := range connections {
		if connMap, ok := connConfig.(map[string]interface{}); ok {
			for key, value := range connMap {
				if valueStr, ok := value.(string); ok {
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

							// 替换配置值
							viper.Set(fmt.Sprintf("connections.%s.%s", connName, key), envValue)
						}
					}
				}
			}
		}
	}
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

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// 尝试加载其他可能的 .env 文件位置
		envPaths := []string{".env", "./config/.env", "../.env"}
		for _, path := range envPaths {
			if err := godotenv.Load(path); err == nil {
				log.Printf("成功加载 .env 文件: %s", path)
				break
			}
		}
		if err != nil {
			log.Printf("未找到 .env 文件，将使用环境变量和默认配置")
		}
	}

	// 设置配置文件路径和名称
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
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

// overrideWithEnv 使用环境变量覆盖配置
func overrideWithEnv() {
	// App配置
	if v := os.Getenv("APP_NAME"); v != "" {
		viper.Set("app.name", v)
	}
	if v := os.Getenv("APP_ENV"); v != "" {
		viper.Set("app.env", v)
	}
	if v := os.Getenv("APP_DEBUG"); v != "" {
		viper.Set("app.debug", v)
	}
	if v := os.Getenv("APP_PORT"); v != "" {
		viper.Set("app.port", v)
	}
	if v := os.Getenv("APP_HOST"); v != "" {
		viper.Set("app.host", v)
	}
	if v := os.Getenv("APP_TIMEZONE"); v != "" {
		viper.Set("app.timezone", v)
	}
	if v := os.Getenv("APP_URL"); v != "" {
		viper.Set("app.url", v)
	}

	// Fiber配置
	if v := os.Getenv("FIBER_PREFORK"); v != "" {
		viper.Set("app.fiber.prefork", v)
	}

	// JWT配置
	if v := os.Getenv("JWT_SECRET"); v != "" {
		viper.Set("jwt.secret", v)
	}
	if v := os.Getenv("JWT_EXPIRATION_TIME"); v != "" {
		viper.Set("jwt.expiration_time", v)
	}
	if v := os.Getenv("JWT_REFRESH_TIME"); v != "" {
		viper.Set("jwt.refresh_time", v)
	}
	if v := os.Getenv("JWT_ISSUER"); v != "" {
		viper.Set("jwt.issuer", v)
	}

	// Redis配置
	if v := os.Getenv("REDIS_HOST"); v != "" {
		viper.Set("redis.host", v)
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		viper.Set("redis.port", v)
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		viper.Set("redis.password", v)
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		viper.Set("redis.db", v)
	}

	// Logger配置
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		viper.Set("logger.level", v)
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		viper.Set("logger.format", v)
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		viper.Set("logger.output", v)
	}
	if v := os.Getenv("LOG_MAX_SIZE"); v != "" {
		viper.Set("logger.max_size", v)
	}
	if v := os.Getenv("LOG_MAX_AGE"); v != "" {
		viper.Set("logger.max_age", v)
	}
	if v := os.Getenv("LOG_MAX_BACKUPS"); v != "" {
		viper.Set("logger.max_backups", v)
	}
	if v := os.Getenv("LOG_COMPRESS"); v != "" {
		viper.Set("logger.compress", v)
	}

	// Cache配置
	if v := os.Getenv("CACHE_DRIVER"); v != "" {
		viper.Set("cache.driver", v)
	}
	if v := os.Getenv("CACHE_PREFIX"); v != "" {
		viper.Set("cache.prefix", v)
	}
	if v := os.Getenv("CACHE_DEFAULT_TTL"); v != "" {
		viper.Set("cache.default", v)
		viper.Set("cache.ttl", v)
	}

	// Mail配置
	if v := os.Getenv("MAIL_FROM_NAME"); v != "" {
		viper.Set("mail.from_name", v)
	}
	if v := os.Getenv("MAIL_FROM_ADDRESS"); v != "" {
		viper.Set("mail.from_address", v)
	}
	if v := os.Getenv("MAIL_HOST"); v != "" {
		viper.Set("mail.host", v)
	}
	if v := os.Getenv("MAIL_PORT"); v != "" {
		viper.Set("mail.port", v)
	}
	if v := os.Getenv("MAIL_USERNAME"); v != "" {
		viper.Set("mail.username", v)
	}
	if v := os.Getenv("MAIL_PASSWORD"); v != "" {
		viper.Set("mail.password", v)
	}
	if v := os.Getenv("MAIL_ENCRYPTION"); v != "" {
		viper.Set("mail.encryption", v)
	}
	if v := os.Getenv("MAIL_TLS_INSECURE"); v != "" {
		viper.Set("mail.tls_insecure", v)
	}

	// Queue配置
	if v := os.Getenv("QUEUE_CONCURRENCY"); v != "" {
		viper.Set("queue.concurrency", v)
	}

	// Storage配置
	if v := os.Getenv("STORAGE_DRIVER"); v != "" {
		viper.Set("storage.driver", v)
	}

	// MinIO配置
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		viper.Set("storage.minio.endpoint", v)
	}
	if v := os.Getenv("MINIO_ACCESS_KEY_ID"); v != "" {
		viper.Set("storage.minio.access_key_id", v)
	}
	if v := os.Getenv("MINIO_SECRET_ACCESS_KEY"); v != "" {
		viper.Set("storage.minio.secret_access_key", v)
	}
	if v := os.Getenv("MINIO_USE_SSL"); v != "" {
		viper.Set("storage.minio.use_ssl", v)
	}
	if v := os.Getenv("MINIO_BUCKET"); v != "" {
		viper.Set("storage.minio.bucket", v)
	}
	if v := os.Getenv("MINIO_REGION"); v != "" {
		viper.Set("storage.minio.region", v)
	}

	// S3配置
	if v := os.Getenv("S3_ACCESS_KEY_ID"); v != "" {
		viper.Set("storage.s3.access_key_id", v)
	}
	if v := os.Getenv("S3_SECRET_ACCESS_KEY"); v != "" {
		viper.Set("storage.s3.secret_access_key", v)
	}
	if v := os.Getenv("S3_REGION"); v != "" {
		viper.Set("storage.s3.region", v)
	}
	if v := os.Getenv("S3_BUCKET"); v != "" {
		viper.Set("storage.s3.bucket", v)
	}
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		viper.Set("storage.s3.endpoint", v)
	}

	// WebSocket配置
	if v := os.Getenv("WEBSOCKET_PORT"); v != "" {
		viper.Set("websocket.port", v)
	}
	if v := os.Getenv("WEBSOCKET_PATH"); v != "" {
		viper.Set("websocket.path", v)
	}
	if v := os.Getenv("WEBSOCKET_HEARTBEAT_INTERVAL"); v != "" {
		viper.Set("websocket.heartbeat_interval", v)
	}

	// Payment配置
	if v := os.Getenv("WECHAT_APP_ID"); v != "" {
		viper.Set("payment.wechat.app_id", v)
	}
	if v := os.Getenv("WECHAT_MCH_ID"); v != "" {
		viper.Set("payment.wechat.mch_id", v)
	}
	if v := os.Getenv("WECHAT_API_KEY"); v != "" {
		viper.Set("payment.wechat.api_key", v)
	}
	if v := os.Getenv("WECHAT_CERT_PATH"); v != "" {
		viper.Set("payment.wechat.cert_path", v)
	}
	if v := os.Getenv("WECHAT_KEY_PATH"); v != "" {
		viper.Set("payment.wechat.key_path", v)
	}
	if v := os.Getenv("WECHAT_NOTIFY_URL"); v != "" {
		viper.Set("payment.wechat.notify_url", v)
	}
	if v := os.Getenv("ALIPAY_APP_ID"); v != "" {
		viper.Set("payment.alipay.app_id", v)
	}
	if v := os.Getenv("ALIPAY_PRIVATE_KEY"); v != "" {
		viper.Set("payment.alipay.private_key", v)
	}
	if v := os.Getenv("ALIPAY_PUBLIC_KEY"); v != "" {
		viper.Set("payment.alipay.public_key", v)
	}
	if v := os.Getenv("ALIPAY_NOTIFY_URL"); v != "" {
		viper.Set("payment.alipay.notify_url", v)
	}

	// Business配置
	if v := os.Getenv("ORDER_PAYMENT_TIMEOUT"); v != "" {
		viper.Set("business.order.payment_timeout", v)
	}
	if v := os.Getenv("ORDER_PICKUP_TIMEOUT"); v != "" {
		viper.Set("business.order.pickup_timeout", v)
	}
	if v := os.Getenv("DEVICE_CHANNEL_COUNT"); v != "" {
		viper.Set("business.device.channel_count", v)
	}
	if v := os.Getenv("DEVICE_CHANNEL_MAX_CAPACITY"); v != "" {
		viper.Set("business.device.channel_max_capacity", v)
	}

	// Security配置
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		viper.Set("security.cors.allowed_origins", v)
	}
	if v := os.Getenv("CORS_ALLOWED_METHODS"); v != "" {
		viper.Set("security.cors.allowed_methods", v)
	}
	if v := os.Getenv("CORS_ALLOWED_HEADERS"); v != "" {
		viper.Set("security.cors.allowed_headers", v)
	}
	if v := os.Getenv("RATE_LIMIT_MAX"); v != "" {
		viper.Set("security.rate_limit.max", v)
	}
	if v := os.Getenv("RATE_LIMIT_WINDOW"); v != "" {
		viper.Set("security.rate_limit.window", v)
	}
}

// setDefaults 设置默认配置值
func setDefaults() {
	viper.SetDefault("app.name", "fiber-starter")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("app.port", "3000")
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("app.timezone", "UTC")
	viper.SetDefault("app.url", "http://localhost:3000")

	// 日志配置默认值
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "console")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.max_size", 100)
	viper.SetDefault("logger.max_age", 28)
	viper.SetDefault("logger.max_backups", 3)
	viper.SetDefault("logger.compress", false)

	// 邮件默认配置
	viper.SetDefault("mail.from_name", "Fiber Starter")
	viper.SetDefault("mail.from_address", "noreply@example.com")
	viper.SetDefault("mail.host", "smtp.example.com")
	viper.SetDefault("mail.port", 587)
	viper.SetDefault("mail.username", "")
	viper.SetDefault("mail.password", "")
	viper.SetDefault("mail.encryption", "tls")
	viper.SetDefault("mail.tls_insecure", false)

	// JWT默认配置
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration_time", 3600) // 1小时
	viper.SetDefault("jwt.refresh_time", 604800)  // 7天
	viper.SetDefault("jwt.expire_hours", 24)      // 24小时
	viper.SetDefault("jwt.issuer", "fiber-starter")

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// 日志默认配置
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.max_size", 100)
	viper.SetDefault("logger.max_age", 30)
	viper.SetDefault("logger.max_backups", 10)
	viper.SetDefault("logger.compress", true)

	// 缓存默认配置
	viper.SetDefault("cache.driver", "redis")
	viper.SetDefault("cache.prefix", "fiber:")
	viper.SetDefault("cache.default", 3600)
	viper.SetDefault("cache.ttl", 3600)

	// 队列默认配置
	viper.SetDefault("queue.concurrency", 10)

	// 存储默认配置
	viper.SetDefault("storage.driver", "minio")
	viper.SetDefault("storage.database", "./storage/storage.db")
	viper.SetDefault("storage.reset", false)
	viper.SetDefault("storage.gc_interval", 10) // 10分钟

	// MinIO默认配置
	viper.SetDefault("storage.minio.endpoint", "localhost:9000")
	viper.SetDefault("storage.minio.access_key_id", "minioadmin")
	viper.SetDefault("storage.minio.secret_access_key", "minioadmin")
	viper.SetDefault("storage.minio.use_ssl", false)
	viper.SetDefault("storage.minio.bucket", "lunchbox-media")
	viper.SetDefault("storage.minio.region", "us-east-1")

	// S3默认配置
	viper.SetDefault("storage.s3.region", "us-east-1")
	viper.SetDefault("storage.s3.bucket", "lunchbox-media")

	// WebSocket默认配置
	viper.SetDefault("websocket.port", "3001")
	viper.SetDefault("websocket.path", "/ws")
	viper.SetDefault("websocket.heartbeat_interval", 30)

	// 支付默认配置
	viper.SetDefault("payment.wechat.notify_url", "")
	viper.SetDefault("payment.alipay.notify_url", "")

	// 业务默认配置
	viper.SetDefault("business.order.payment_timeout", 30)
	viper.SetDefault("business.order.pickup_timeout", 1440)
	viper.SetDefault("business.device.channel_count", 53)
	viper.SetDefault("business.device.channel_max_capacity", 4)

	// 安全默认配置
	viper.SetDefault("security.cors.allowed_origins", "*")
	viper.SetDefault("security.cors.allowed_methods", "GET,POST,PUT,DELETE,OPTIONS")
	viper.SetDefault("security.cors.allowed_headers", "Origin,Content-Type,Accept,Authorization")
	viper.SetDefault("security.rate_limit.max", 60)
	viper.SetDefault("security.rate_limit.window", 60)
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
