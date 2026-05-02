// Package config 处理应用程序的配置加载和管理
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

type MailConfig struct {
	FromName    string `mapstructure:"from_name"`
	FromAddress string `mapstructure:"from_address"`
	ReplyTo     string `mapstructure:"reply_to"`
	APIKey      string `mapstructure:"api_key"`
}

type StorageConfig struct {
	Driver     string               `mapstructure:"driver"`
	Database   string               `mapstructure:"database"`
	Reset      bool                 `mapstructure:"reset"`
	GCInterval int                  `mapstructure:"gc_interval"`
	Garage     *GarageStorageConfig `mapstructure:"garage"`
	MinIO      *GarageStorageConfig `mapstructure:"minio"`
	S3         *S3StorageConfig     `mapstructure:"s3"`
	R2         *S3StorageConfig     `mapstructure:"r2"`
	OSS        *S3StorageConfig     `mapstructure:"oss"`
}

type GarageStorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
}

type S3StorageConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	Endpoint        string `mapstructure:"endpoint"`
}

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

type MeilisearchConfig struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
}

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

type FiberConfig struct {
	Prefork           bool   `mapstructure:"prefork"`
	ServerHeader      string `mapstructure:"server_header"`
	BodyLimit         int    `mapstructure:"body_limit"`
	Concurrency       int    `mapstructure:"concurrency"`
	ReadBufferSize    int    `mapstructure:"read_buffer_size"`
	ReadTimeout       int    `mapstructure:"read_timeout"`
	WriteTimeout      int    `mapstructure:"write_timeout"`
	IdleTimeout       int    `mapstructure:"idle_timeout"`
	TrustProxy        bool   `mapstructure:"trust_proxy"`
	ProxyHeader       string `mapstructure:"proxy_header"`
	StreamRequestBody bool   `mapstructure:"stream_request_body"`
	Immutable         bool   `mapstructure:"immutable"`
}

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

type DBPoolConfig struct {
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"`
}

type DBReadConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	Port     string   `mapstructure:"port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

type DBWriteConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type DBMigrationConfig struct {
	Table string `mapstructure:"table"`
	Path  string `mapstructure:"path"`
}

type DBSeederConfig struct {
	Path string `mapstructure:"path"`
}

type DBRedisConfig struct {
	Client  string                 `mapstructure:"client"`
	Options map[string]interface{} `mapstructure:"options"`
	Default map[string]interface{} `mapstructure:"default"`
}

type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	ExpirationTime int    `mapstructure:"expiration_time"`
	RefreshTime    int    `mapstructure:"refresh_time"`
	ExpireHours    int    `mapstructure:"expire_hours"`
	Issuer         string `mapstructure:"issuer"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type CacheConfig struct {
	Driver  string `mapstructure:"driver"`
	Prefix  string `mapstructure:"prefix"`
	Default int    `mapstructure:"default"`
	TTL     int    `mapstructure:"ttl"`
}

type QueueConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

type WebSocketConfig struct {
	Port              string `mapstructure:"port"`
	Path              string `mapstructure:"path"`
	HeartbeatInterval int    `mapstructure:"heartbeat_interval"`
}

type PaymentConfig struct {
	Wechat WechatPaymentConfig `mapstructure:"wechat"`
	Alipay AlipayPaymentConfig `mapstructure:"alipay"`
}

type WechatPaymentConfig struct {
	AppID     string `mapstructure:"app_id"`
	MchID     string `mapstructure:"mch_id"`
	APIKey    string `mapstructure:"api_key"`
	CertPath  string `mapstructure:"cert_path"`
	KeyPath   string `mapstructure:"key_path"`
	NotifyURL string `mapstructure:"notify_url"`
}

type AlipayPaymentConfig struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
	NotifyURL  string `mapstructure:"notify_url"`
}

type BusinessConfig struct {
	Order  OrderConfig  `mapstructure:"order"`
	Device DeviceConfig `mapstructure:"device"`
}

type OrderConfig struct {
	PaymentTimeout int `mapstructure:"payment_timeout"`
	PickupTimeout  int `mapstructure:"pickup_timeout"`
}

type DeviceConfig struct {
	ChannelCount       int `mapstructure:"channel_count"`
	ChannelMaxCapacity int `mapstructure:"channel_max_capacity"`
}

type SecurityConfig struct {
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
	AllowedMethods string `mapstructure:"allowed_methods"`
	AllowedHeaders string `mapstructure:"allowed_headers"`
}

type RateLimitConfig struct {
	Max    int `mapstructure:"max"`
	Window int `mapstructure:"window"`
}

var GlobalConfig *Config

var loadEnvOnce sync.Once

func loadEnvFile() {
	loadEnvOnce.Do(func() {
		appEnv := os.Getenv("APP_ENV")
		if appEnv == "" {
			appEnv = "development"
		}

		files := []string{fmt.Sprintf(".env.%s.local", appEnv)}
		if appEnv != "test" {
			files = append(files, ".env.local")
		}
		files = append(files, fmt.Sprintf(".env.%s", appEnv), ".env")

		loaded := false
		for _, file := range files {
			for _, path := range envFileCandidates(file) {
				if fileExists(path) {
					if err := godotenv.Load(path); err == nil {
						log.Printf("Successfully loaded environment file: %s", path) //nolint:gosec // environment file path is controlled by local project config
						loaded = true
						break
					}
				}
			}
		}

		if !loaded {
			log.Printf("Environment file not found, will use environment variables and default configuration")
		}
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path) //nolint:gosec // path is derived from local config/env file lookup
	return !os.IsNotExist(err)
}

func LoadDatabaseConfig() (*DatabaseConfig, error) {
	loadEnvFile()

	dbConfig := &DatabaseConfig{}
	k := koanf.New(".")
	if err := k.Load(confmap.Provider(defaultConfigMap(), "."), nil); err != nil {
		return nil, err
	}
	if err := k.Load(confmap.Provider(envConfigMap(), "."), nil); err != nil {
		return nil, err
	}
	if err := k.Unmarshal("", dbConfig); err != nil {
		return nil, err
	}
	return dbConfig, nil
}

func LoadConfig() (*Config, error) {
	loadEnvFile()

	k := koanf.New(".")
	if err := k.Load(confmap.Provider(defaultConfigMap(), "."), nil); err != nil {
		return nil, err
	}
	if err := k.Load(confmap.Provider(envConfigMap(), "."), nil); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}

	if cfg.Database.Default == "" {
		cfg.Database.Default = "postgres"
	}

	GlobalConfig = cfg
	return cfg, nil
}

// Init loads the application config and stores it in GlobalConfig.
func Init() error {
	_, err := LoadConfig()
	return err
}

func defaultConfigMap() map[string]any {
	return map[string]any{
		"app.name":                               "Fiber Starter",
		"app.env":                                "development",
		"app.debug":                              true,
		"app.port":                               "8080",
		"app.host":                               "0.0.0.0",
		"app.timezone":                           "UTC",
		"app.url":                                "http://localhost:8080",
		"app.fiber.prefork":                      false,
		"app.fiber.server_header":                "",
		"app.fiber.body_limit":                   4 * 1024 * 1024,
		"app.fiber.concurrency":                  256 * 1024,
		"app.fiber.read_buffer_size":             16 * 1024,
		"app.fiber.read_timeout":                 30,
		"app.fiber.write_timeout":                30,
		"app.fiber.idle_timeout":                 120,
		"app.fiber.trust_proxy":                  false,
		"app.fiber.proxy_header":                 "X-Forwarded-For",
		"app.fiber.stream_request_body":          false,
		"app.fiber.immutable":                    false,
		"database.default":                       "postgres",
		"database.connections.postgres.driver":   "postgres",
		"database.connections.postgres.host":     "localhost",
		"database.connections.postgres.port":     "5432",
		"database.connections.postgres.database": "fiber_starter",
		"database.connections.postgres.username": "postgres",
		"database.connections.postgres.password": "",
		"database.connections.postgres.charset":  "utf8",
		"database.connections.postgres.prefix":   "",
		"database.connections.postgres.schema":   "public",
		"database.connections.postgres.sslmode":  "disable",
		"database.connections.postgres.timezone": "UTC",
		"database.connections.pgsql.driver":      "postgres",
		"database.connections.pgsql.host":        "localhost",
		"database.connections.pgsql.port":        "5432",
		"database.connections.pgsql.database":    "fiber_starter",
		"database.connections.pgsql.username":    "postgres",
		"database.connections.pgsql.password":    "",
		"database.connections.pgsql.charset":     "utf8",
		"database.connections.pgsql.prefix":      "",
		"database.connections.pgsql.schema":      "public",
		"database.connections.pgsql.sslmode":     "disable",
		"database.connections.pgsql.timezone":    "UTC",
		"database.connections.sqlite.driver":     "sqlite",
		"database.connections.sqlite.database":   "./database/database.sqlite",
		"database.connections.sqlite.prefix":     "",
		"database.pool.max_open_conns":           100,
		"database.pool.max_idle_conns":           10,
		"database.pool.conn_max_lifetime":        3600,
		"database.pool.conn_max_idle_time":       600,
		"database.migrations.table":              "migrations",
		"database.migrations.path":               "./database/migrations",
		"database.seeders.path":                  "./database/seeders",
		"database.redis.client":                  "default",
		"queue.concurrency":                      10,
		"cache.driver":                           "redis",
		"cache.prefix":                           "fiber_starter_cache",
		"cache.default":                          0,
		"cache.ttl":                              3600,
		"redis.host":                             "127.0.0.1",
		"redis.port":                             "6379",
		"redis.password":                         "",
		"redis.db":                               0,
		"jwt.secret":                             "change-me",
		"jwt.expiration_time":                    86400,
		"jwt.refresh_time":                       604800,
		"jwt.expire_hours":                       24,
		"jwt.issuer":                             "fiber-starter",
		"logger.level":                           "info",
		"logger.format":                          "json",
		"logger.output":                          "stdout",
		"logger.max_size":                        100,
		"logger.max_age":                         30,
		"logger.max_backups":                     10,
		"logger.compress":                        true,
		"mail.from_name":                         "Fiber Starter",
		"mail.from_address":                      "noreply@example.com",
		"mail.reply_to":                          "",
		"mail.api_key":                           "",
		"storage.driver":                         "redis",
		"storage.database":                       "./database/storage.db",
		"storage.reset":                          false,
		"storage.gc_interval":                    3600,
		"storage.garage.endpoint":                "",
		"storage.garage.access_key_id":           "",
		"storage.garage.secret_access_key":       "",
		"storage.garage.use_ssl":                 false,
		"storage.garage.bucket":                  "",
		"storage.garage.region":                  "",
		"storage.minio.endpoint":                 "",
		"storage.minio.access_key_id":            "",
		"storage.minio.secret_access_key":        "",
		"storage.minio.use_ssl":                  false,
		"storage.minio.bucket":                   "",
		"storage.minio.region":                   "",
		"storage.s3.access_key_id":               "",
		"storage.s3.secret_access_key":           "",
		"storage.s3.region":                      "",
		"storage.s3.bucket":                      "",
		"storage.s3.endpoint":                    "",
		"storage.r2.access_key_id":               "",
		"storage.r2.secret_access_key":           "",
		"storage.r2.region":                      "",
		"storage.r2.bucket":                      "",
		"storage.r2.endpoint":                    "",
		"storage.oss.access_key_id":              "",
		"storage.oss.secret_access_key":          "",
		"storage.oss.region":                     "",
		"storage.oss.bucket":                     "",
		"storage.oss.endpoint":                   "",
		"websocket.port":                         "3001",
		"websocket.path":                         "/ws",
		"websocket.heartbeat_interval":           30,
		"payment.wechat.app_id":                  "",
		"payment.wechat.mch_id":                  "",
		"payment.wechat.api_key":                 "",
		"payment.wechat.cert_path":               "",
		"payment.wechat.key_path":                "",
		"payment.wechat.notify_url":              "",
		"payment.alipay.app_id":                  "",
		"payment.alipay.private_key":             "",
		"payment.alipay.public_key":              "",
		"payment.alipay.notify_url":              "",
		"business.order.payment_timeout":         30,
		"business.order.pickup_timeout":          1440,
		"business.device.channel_count":          53,
		"business.device.channel_max_capacity":   4,
		"security.cors.allowed_origins":          "http://localhost:3000",
		"security.cors.allowed_methods":          "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		"security.cors.allowed_headers":          "Origin,Content-Type,Accept,Authorization,Cache-Control,X-Requested-With",
		"security.rate_limit.max":                100,
		"security.rate_limit.window":             60,
		"i18n.default_language":                  "en",
		"i18n.supported_languages":               []string{"en", "zh-CN"},
		"i18n.language_dir":                      "./lang",
		"i18n.cookie_name":                       "lang",
		"i18n.cookie_max_age":                    86400,
		"meilisearch.host":                       "",
		"meilisearch.api_key":                    "",
	}
}

func envConfigMap() map[string]any {
	m := map[string]any{}
	setString := func(envKey, target string) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			m[target] = value
		}
	}
	setBool := func(envKey, target string) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			if parsed, err := strconv.ParseBool(value); err == nil {
				m[target] = parsed
			}
		}
	}
	setInt := func(envKey, target string) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			if parsed, err := strconv.Atoi(value); err == nil {
				m[target] = parsed
			}
		}
	}

	setString("APP_NAME", "app.name")
	setString("APP_ENV", "app.env")
	setBool("APP_DEBUG", "app.debug")
	setString("APP_PORT", "app.port")
	setString("APP_HOST", "app.host")
	setString("APP_TIMEZONE", "app.timezone")
	setString("APP_URL", "app.url")
	setBool("APP_FIBER_PREFORK", "app.fiber.prefork")
	setBool("FIBER_PREFORK", "app.fiber.prefork")
	setString("APP_FIBER_SERVER_HEADER", "app.fiber.server_header")
	setInt("APP_FIBER_BODY_LIMIT", "app.fiber.body_limit")
	setInt("APP_FIBER_CONCURRENCY", "app.fiber.concurrency")
	setInt("APP_FIBER_READ_BUFFER_SIZE", "app.fiber.read_buffer_size")
	setInt("APP_FIBER_READ_TIMEOUT", "app.fiber.read_timeout")
	setInt("APP_FIBER_WRITE_TIMEOUT", "app.fiber.write_timeout")
	setInt("APP_FIBER_IDLE_TIMEOUT", "app.fiber.idle_timeout")
	setBool("APP_FIBER_TRUST_PROXY", "app.fiber.trust_proxy")
	setString("APP_FIBER_PROXY_HEADER", "app.fiber.proxy_header")
	setBool("APP_FIBER_STREAM_REQUEST_BODY", "app.fiber.stream_request_body")
	setBool("APP_FIBER_IMMUTABLE", "app.fiber.immutable")
	setInt("FIBER_READ_BUFFER_SIZE", "app.fiber.read_buffer_size")

	dbDefault := strings.ToLower(strings.TrimSpace(os.Getenv("DB_CONNECTION")))
	if dbDefault == "" {
		dbDefault = "postgres"
	}
	m["database.default"] = dbDefault
	for _, conn := range []string{"postgres", "pgsql"} {
		setString("DB_HOST", fmt.Sprintf("database.connections.%s.host", conn))
		setString("DB_PORT", fmt.Sprintf("database.connections.%s.port", conn))
		setString("DB_DATABASE", fmt.Sprintf("database.connections.%s.database", conn))
		setString("DB_USERNAME", fmt.Sprintf("database.connections.%s.username", conn))
		setString("DB_PASSWORD", fmt.Sprintf("database.connections.%s.password", conn))
		setString("DB_CHARSET", fmt.Sprintf("database.connections.%s.charset", conn))
		setString("DB_COLLATION", fmt.Sprintf("database.connections.%s.collation", conn))
		setString("DB_PREFIX", fmt.Sprintf("database.connections.%s.prefix", conn))
		setBool("DB_STRICT", fmt.Sprintf("database.connections.%s.strict", conn))
		setString("DB_TIMEZONE", fmt.Sprintf("database.connections.%s.timezone", conn))
		setString("DB_SCHEMA", fmt.Sprintf("database.connections.%s.schema", conn))
		setString("DB_SSLMODE", fmt.Sprintf("database.connections.%s.sslmode", conn))
	}
	if dbDefault == "sqlite" {
		setString("DB_DATABASE", "database.connections.sqlite.database")
		setString("DB_SQLITE_DATABASE", "database.connections.sqlite.database")
	}

	setInt("DB_POOL_MAX_OPEN_CONNS", "database.pool.max_open_conns")
	setInt("DB_POOL_MAX_IDLE_CONNS", "database.pool.max_idle_conns")
	setInt("DB_POOL_CONN_MAX_LIFETIME", "database.pool.conn_max_lifetime")
	setInt("DB_POOL_CONN_MAX_IDLE_TIME", "database.pool.conn_max_idle_time")
	setString("DB_MIGRATIONS_TABLE", "database.migrations.table")
	setString("DB_MIGRATIONS_PATH", "database.migrations.path")
	setString("DB_SEEDERS_PATH", "database.seeders.path")

	setString("REDIS_HOST", "redis.host")
	setString("REDIS_PORT", "redis.port")
	setString("REDIS_PASSWORD", "redis.password")
	setInt("REDIS_DB", "redis.db")

	setInt("QUEUE_CONCURRENCY", "queue.concurrency")

	setString("CACHE_DRIVER", "cache.driver")
	setString("CACHE_PREFIX", "cache.prefix")
	setInt("CACHE_DEFAULT", "cache.default")
	setInt("CACHE_DEFAULT_TTL", "cache.ttl")
	setInt("CACHE_TTL", "cache.ttl")

	setString("JWT_SECRET", "jwt.secret")
	setInt("JWT_EXPIRATION_TIME", "jwt.expiration_time")
	setInt("JWT_REFRESH_TIME", "jwt.refresh_time")
	setInt("JWT_EXPIRE_HOURS", "jwt.expire_hours")
	setString("JWT_ISSUER", "jwt.issuer")

	setString("LOGGER_LEVEL", "logger.level")
	setString("LOG_LEVEL", "logger.level")
	setString("LOGGER_FORMAT", "logger.format")
	setString("LOG_FORMAT", "logger.format")
	setString("LOGGER_OUTPUT", "logger.output")
	setString("LOG_OUTPUT", "logger.output")
	setInt("LOGGER_MAX_SIZE", "logger.max_size")
	setInt("LOG_MAX_SIZE", "logger.max_size")
	setInt("LOGGER_MAX_AGE", "logger.max_age")
	setInt("LOG_MAX_AGE", "logger.max_age")
	setInt("LOGGER_MAX_BACKUPS", "logger.max_backups")
	setInt("LOG_MAX_BACKUPS", "logger.max_backups")
	setBool("LOGGER_COMPRESS", "logger.compress")
	setBool("LOG_COMPRESS", "logger.compress")

	setString("MAIL_FROM_NAME", "mail.from_name")
	setString("MAIL_FROM_ADDRESS", "mail.from_address")
	setString("MAIL_REPLY_TO", "mail.reply_to")
	setString("RESEND_API_KEY", "mail.api_key")
	setString("MAIL_API_KEY", "mail.api_key")

	setString("STORAGE_DRIVER", "storage.driver")
	setString("STORAGE_DATABASE", "storage.database")
	setBool("STORAGE_RESET", "storage.reset")
	setInt("STORAGE_GC_INTERVAL", "storage.gc_interval")
	setString("GARAGE_ENDPOINT", "storage.garage.endpoint")
	setString("GARAGE_ACCESS_KEY_ID", "storage.garage.access_key_id")
	setString("GARAGE_SECRET_ACCESS_KEY", "storage.garage.secret_access_key")
	setBool("GARAGE_USE_SSL", "storage.garage.use_ssl")
	setString("GARAGE_BUCKET", "storage.garage.bucket")
	setString("GARAGE_REGION", "storage.garage.region")
	setString("MINIO_ENDPOINT", "storage.minio.endpoint")
	setString("MINIO_ACCESS_KEY_ID", "storage.minio.access_key_id")
	setString("MINIO_SECRET_ACCESS_KEY", "storage.minio.secret_access_key")
	setBool("MINIO_USE_SSL", "storage.minio.use_ssl")
	setString("MINIO_BUCKET", "storage.minio.bucket")
	setString("MINIO_REGION", "storage.minio.region")
	setString("S3_ACCESS_KEY_ID", "storage.s3.access_key_id")
	setString("S3_SECRET_ACCESS_KEY", "storage.s3.secret_access_key")
	setString("S3_REGION", "storage.s3.region")
	setString("S3_BUCKET", "storage.s3.bucket")
	setString("S3_ENDPOINT", "storage.s3.endpoint")
	setString("R2_ACCESS_KEY_ID", "storage.r2.access_key_id")
	setString("R2_SECRET_ACCESS_KEY", "storage.r2.secret_access_key")
	setString("R2_REGION", "storage.r2.region")
	setString("R2_BUCKET", "storage.r2.bucket")
	setString("R2_ENDPOINT", "storage.r2.endpoint")
	setString("OSS_ACCESS_KEY_ID", "storage.oss.access_key_id")
	setString("OSS_SECRET_ACCESS_KEY", "storage.oss.secret_access_key")
	setString("OSS_REGION", "storage.oss.region")
	setString("OSS_BUCKET", "storage.oss.bucket")
	setString("OSS_ENDPOINT", "storage.oss.endpoint")

	setString("WEBSOCKET_PORT", "websocket.port")
	setString("WEBSOCKET_PATH", "websocket.path")
	setInt("WEBSOCKET_HEARTBEAT_INTERVAL", "websocket.heartbeat_interval")

	setString("WECHAT_APP_ID", "payment.wechat.app_id")
	setString("WECHAT_MCH_ID", "payment.wechat.mch_id")
	setString("WECHAT_API_KEY", "payment.wechat.api_key")
	setString("WECHAT_CERT_PATH", "payment.wechat.cert_path")
	setString("WECHAT_KEY_PATH", "payment.wechat.key_path")
	setString("WECHAT_NOTIFY_URL", "payment.wechat.notify_url")
	setString("ALIPAY_APP_ID", "payment.alipay.app_id")
	setString("ALIPAY_PRIVATE_KEY", "payment.alipay.private_key")
	setString("ALIPAY_PUBLIC_KEY", "payment.alipay.public_key")
	setString("ALIPAY_NOTIFY_URL", "payment.alipay.notify_url")

	setInt("ORDER_PAYMENT_TIMEOUT", "business.order.payment_timeout")
	setInt("ORDER_PICKUP_TIMEOUT", "business.order.pickup_timeout")
	setInt("DEVICE_CHANNEL_COUNT", "business.device.channel_count")
	setInt("DEVICE_CHANNEL_MAX_CAPACITY", "business.device.channel_max_capacity")

	setString("SECURITY_CORS_ALLOWED_ORIGINS", "security.cors.allowed_origins")
	setString("SECURITY_CORS_ALLOWED_METHODS", "security.cors.allowed_methods")
	setString("SECURITY_CORS_ALLOWED_HEADERS", "security.cors.allowed_headers")
	setInt("SECURITY_RATE_LIMIT_MAX", "security.rate_limit.max")
	setInt("SECURITY_RATE_LIMIT_WINDOW", "security.rate_limit.window")

	setString("I18N_DEFAULT_LANGUAGE", "i18n.default_language")
	if value := strings.TrimSpace(os.Getenv("I18N_SUPPORTED_LANGUAGES")); value != "" {
		parts := strings.Split(value, ",")
		supported := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				supported = append(supported, trimmed)
			}
		}
		if len(supported) > 0 {
			m["i18n.supported_languages"] = supported
		}
	}
	setString("I18N_LANGUAGE_DIR", "i18n.language_dir")
	setString("I18N_COOKIE_NAME", "i18n.cookie_name")
	setInt("I18N_COOKIE_MAX_AGE", "i18n.cookie_max_age")

	setString("MEILISEARCH_HOST", "meilisearch.host")
	setString("MEILISEARCH_API_KEY", "meilisearch.api_key")

	return m
}

func envFileCandidates(file string) []string {
	return []string{
		file,
		filepath.Join("config", file),
		filepath.Join("configs", file),
		filepath.Join("..", file),
	}
}
