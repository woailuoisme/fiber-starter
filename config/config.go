// Package config 处理应用程序的配置加载和管理
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
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
	Prefork        bool   `mapstructure:"prefork"`
	ServerHeader   string `mapstructure:"server_header"`
	BodyLimit      int    `mapstructure:"body_limit"`
	Concurrency    int    `mapstructure:"concurrency"`
	ReadBufferSize int    `mapstructure:"read_buffer_size"`
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

		var envFiles []string
		if appEnv != "" {
			envFiles = []string{
				fmt.Sprintf(".env.%s", appEnv),
				".env",
			}
		} else {
			envFiles = []string{".env"}
		}

		var loaded bool
		for _, envFile := range envFiles {
			envPaths := []string{
				envFile,
				fmt.Sprintf("./config/%s", envFile),
				fmt.Sprintf("./configs/%s", envFile),
				fmt.Sprintf("../%s", envFile),
			}

			for _, path := range envPaths {
				if fileExists(path) {
					v := viper.New()
					v.SetConfigFile(path)
					v.SetConfigType("env")
					if err := v.ReadInConfig(); err == nil {
						for k, v := range v.AllSettings() {
							if strVal, ok := v.(string); ok {
								if os.Getenv(k) == "" {
									_ = os.Setenv(k, strVal)
								}
							}
						}
						log.Printf("Successfully loaded environment file: %s", path) //nolint:gosec // environment file path is controlled by local project config
						loaded = true
						break
					}
				}
			}

			if loaded {
				break
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
	dbConfig := &DatabaseConfig{}
	loadEnvFile()
	viper.SetConfigName("database")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DB")
	setDatabaseDefaults()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
			log.Printf("Database configuration file not found, using default configuration and environment variables")
		} else {
			return nil, err
		}
	}
	replaceEnvVars()
	if err := viper.Unmarshal(dbConfig); err != nil {
		return nil, err
	}
	if dbConnection := os.Getenv("DB_CONNECTION"); dbConnection != "" {
		dbConfig.Default = dbConnection
	}
	return dbConfig, nil
}

func setDatabaseDefaults() {
	defaults := map[string]interface{}{
		"default":                     "mysql",
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
		"connections.pgsql.driver":    "postgres",
		"connections.pgsql.host":      "localhost",
		"connections.pgsql.port":      "5432",
		"connections.pgsql.database":  "fiber_starter",
		"connections.pgsql.username":  "postgres",
		"connections.pgsql.password":  "",
		"connections.pgsql.charset":   "utf8",
		"connections.pgsql.prefix":    "",
		"connections.pgsql.schema":    "public",
		"connections.pgsql.sslmode":   "disable",
		"connections.pgsql.timezone":  "UTC",
		"connections.sqlite.driver":   "sqlite",
		"connections.sqlite.database": "./database/database.sqlite",
		"connections.sqlite.prefix":   "",
		"pool.max_open_conns":         100,
		"pool.max_idle_conns":         10,
		"pool.conn_max_lifetime":      3600,
		"pool.conn_max_idle_time":     600,
		"migrations.table":            "migrations",
		"migrations.path":             "./database/migrations",
		"seeders.path":                "./database/seeders",
	}
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}
}

func replaceEnvVars() {
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

func processEnvVarSubstitution(valueStr string) (string, bool) {
	if strings.Contains(valueStr, "${") && strings.Contains(valueStr, "}") {
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
			envValue := os.Getenv(envKey)
			if envValue == "" {
				envValue = defaultValue
			}
			return envValue, true
		}
	}
	return "", false
}

func LoadConfig() (*Config, error) {
	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		log.Printf("Failed to load database configuration: %v", err)
		dbConfig = &DatabaseConfig{}
	}

	config := &Config{Database: *dbConfig}
	loadEnvFile()

	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	setAppDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
			log.Printf("App configuration file not found, using default configuration and environment variables")
		} else {
			return nil, err
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	GlobalConfig = config
	return config, nil
}

// Init loads the application config and stores it in GlobalConfig.
func Init() error {
	_, err := LoadConfig()
	return err
}

func setAppDefaults() {
	viper.SetDefault("app.name", "Fiber Starter")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("app.timezone", "UTC")
	viper.SetDefault("app.url", "http://localhost:8080")
	viper.SetDefault("app.fiber.prefork", false)
	viper.SetDefault("app.fiber.server_header", "")
	viper.SetDefault("app.fiber.body_limit", 4*1024*1024)
	viper.SetDefault("app.fiber.concurrency", 256*1024)
	viper.SetDefault("app.fiber.read_buffer_size", 16*1024)
}
