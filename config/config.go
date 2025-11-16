package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// MailConfig 邮件配置
type MailConfig struct {
	FromName       string `mapstructure:"from_name"`
	FromAddress    string `mapstructure:"from_address"`
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Encryption     string `mapstructure:"encryption"`
	TLSInsecure    bool   `mapstructure:"tls_insecure"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Driver     string `mapstructure:"driver"`
	Database   string `mapstructure:"database"`
	Reset      bool   `mapstructure:"reset"`
	GCInterval int    `mapstructure:"gc_interval"`
}

// Config 应用程序配置结构体
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Mail     MailConfig     `mapstructure:"mail"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

// AppConfig 应用程序基础配置
type AppConfig struct {
	Name   string `mapstructure:"name"`
	Env    string `mapstructure:"env"`
	Debug  bool   `mapstructure:"debug"`
	Port   string `mapstructure:"port"`
	Host   string `mapstructure:"host"`
	Timezone string `mapstructure:"timezone"`
	URL    string `mapstructure:"url"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Connection string `mapstructure:"connection"`
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	Database   string `mapstructure:"database"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	Charset    string `mapstructure:"charset"`
	Timezone   string `mapstructure:"timezone"`
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
	Driver   string `mapstructure:"driver"`
	Prefix   string `mapstructure:"prefix"`
	Default  int    `mapstructure:"default"`
	TTL      int    `mapstructure:"ttl"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

// 全局配置实例
var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	configPath := "./config"
	config := &Config{}

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
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置环境变量前缀
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

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

	// 解析配置到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 应用程序默认配置
	viper.SetDefault("app.name", "Fiber Starter")
	viper.SetDefault("app.env", "local")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("app.port", "3000")
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("app.timezone", "UTC")
	viper.SetDefault("app.url", "http://localhost:3000")

	// 数据库默认配置
	viper.SetDefault("database.connection", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.database", "fiber_starter")
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.timezone", "UTC")

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
	viper.SetDefault("storage.driver", "memory")
	viper.SetDefault("storage.database", "./storage/storage.db")
	viper.SetDefault("storage.reset", false)
	viper.SetDefault("storage.gc_interval", 10) // 10分钟
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