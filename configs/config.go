// Package configs 处理应用程序的配置加载和管理
package configs

import (
	"fmt"

	"lfiber/configs/internal"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

// Type aliases to preserve backward compatibility for the rest of the application.
type (
	Config                     = internal.Config
	AppConfig                  = internal.AppConfig
	FiberConfig                = internal.FiberConfig
	DatabaseConfig             = internal.DatabaseConfig
	DBConnection               = internal.DBConnection
	DBPoolConfig               = internal.DBPoolConfig
	DBReadConfig               = internal.DBReadConfig
	DBWriteConfig              = internal.DBWriteConfig
	DBMigrationConfig          = internal.DBMigrationConfig
	DBSeederConfig             = internal.DBSeederConfig
	DBRedisConfig              = internal.DBRedisConfig
	AuthConfig                 = internal.AuthConfig
	AuthorizationConfig        = internal.AuthorizationConfig
	GuardConfig                = internal.GuardConfig
	ProviderConfig             = internal.ProviderConfig
	JWTConfig                  = internal.JWTConfig
	RedisConfig                = internal.RedisConfig
	LoggerConfig               = internal.LoggerConfig
	CacheConfig                = internal.CacheConfig
	MailConfig                 = internal.MailConfig
	NotificationConfig         = internal.NotificationConfig
	GotifyNotificationConfig   = internal.GotifyNotificationConfig
	TelegramNotificationConfig = internal.TelegramNotificationConfig
	QueueConfig                = internal.QueueConfig
	StorageConfig              = internal.StorageConfig
	MediaLibraryConfig         = internal.MediaLibraryConfig
	BackupConfig               = internal.BackupConfig
	BackupNotificationConfig   = internal.BackupNotificationConfig
	BackupBinaryConfig         = internal.BackupBinaryConfig
	LocalStorageConfig         = internal.LocalStorageConfig
	GarageStorageConfig        = internal.GarageStorageConfig
	S3StorageConfig            = internal.S3StorageConfig
	WebSocketConfig            = internal.WebSocketConfig
	PaymentConfig              = internal.PaymentConfig
	WechatPaymentConfig        = internal.WechatPaymentConfig
	AlipayPaymentConfig        = internal.AlipayPaymentConfig
	BusinessConfig             = internal.BusinessConfig
	OrderConfig                = internal.OrderConfig
	DeviceConfig               = internal.DeviceConfig
	SecurityConfig             = internal.SecurityConfig
	CORSConfig                 = internal.CORSConfig
	RateLimitConfig            = internal.RateLimitConfig
	LimiterConfig              = internal.LimiterConfig
	LoadShedConfig             = internal.LoadShedConfig
	CircuitBreakerConfig       = internal.CircuitBreakerConfig
	LogViewerConfig            = internal.LogViewerConfig
	I18nConfig                 = internal.I18nConfig
	SearchConfig               = internal.SearchConfig
	HashConfig                 = internal.HashConfig
	BcryptHashConfig           = internal.BcryptHashConfig
	Argon2HashConfig           = internal.Argon2HashConfig
	OTELConfig                 = internal.OTELConfig
	ServicesConfig             = internal.ServicesConfig
	ServiceDependencyConfig    = internal.ServiceDependencyConfig
)

// GlobalConfig stores the globally accessible configuration instance.
var GlobalConfig *Config

// Init loads the application config and stores it in GlobalConfig.
func Init() error {
	_, _, err := LoadConfig()
	return err
}

// LoadConfig initializes the configuration by merging defaults, file-based configs, and environment variables.
func LoadConfig() (*Config, *koanf.Koanf, error) {
	internal.LoadEnvFile()

	k := koanf.New(".")
	if err := k.Load(confmap.Provider(internal.DefaultConfigMap(), "."), nil); err != nil {
		return nil, nil, err
	}
	if err := internal.LoadConfigFiles(k); err != nil {
		internal.LogWarn(fmt.Sprintf("Warning: failed to load config files from configs/: %v", err))
	}
	if err := k.Load(confmap.Provider(internal.EnvConfigMap(), "."), nil); err != nil {
		return nil, nil, err
	}

	cfg := &internal.Config{}
	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "mapstructure"}); err != nil {
		return nil, nil, err
	}
	applyUnmarshalFallbacks(k, cfg)

	if cfg.Database.Default == "" {
		cfg.Database.Default = "postgres"
	}

	// Validate config
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, nil, fmt.Errorf("config validation failed: %w", err)
	}

	GlobalConfig = cfg
	return cfg, k, nil
}

func applyUnmarshalFallbacks(k *koanf.Koanf, cfg *internal.Config) {
	if k == nil || cfg == nil {
		return
	}
	if cfg.Notification.Telegram.APIURL == "" {
		cfg.Notification.Telegram.APIURL = k.String("notification.telegram.api_url")
	}
	if cfg.Notification.Telegram.BotToken == "" {
		cfg.Notification.Telegram.BotToken = k.String("notification.telegram.bot_token")
	}
	if cfg.Notification.Telegram.ChatID == "" {
		cfg.Notification.Telegram.ChatID = k.String("notification.telegram.chat_id")
	}
	if cfg.Notification.Telegram.ParseMode == "" {
		cfg.Notification.Telegram.ParseMode = k.String("notification.telegram.parse_mode")
	}
	if cfg.Logger.Output == "" {
		cfg.Logger.Output = k.String("logger.channel")
	}
	cfg.OTEL.TraceEnabled = k.Bool("otel.trace_enabled")
	if cfg.OTEL.ServiceName == "" {
		cfg.OTEL.ServiceName = k.String("otel.service_name")
	}
	if cfg.OTEL.ExporterType == "" {
		cfg.OTEL.ExporterType = k.String("otel.exporter_type")
	}
	if cfg.OTEL.Endpoint == "" {
		cfg.OTEL.Endpoint = k.String("otel.endpoint")
	}
	cfg.OTEL.OTLPInsecure = k.Bool("otel.otlp_insecure")
	cfg.OTEL.TraceSampleRatio = k.Float64("otel.trace_sample_ratio")
	cfg.OTEL.MetricsEnabled = k.Bool("otel.metrics_enabled")
	if cfg.OTEL.MetricsPath == "" {
		cfg.OTEL.MetricsPath = k.String("otel.metrics_path")
	}
}

// LoadDatabaseConfig specifically loads the database configuration, useful for migration tools.
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	internal.LoadEnvFile()

	dbConfig := &internal.DatabaseConfig{}
	k := koanf.New(".")
	if err := k.Load(confmap.Provider(internal.DefaultConfigMap(), "."), nil); err != nil {
		return nil, err
	}
	if err := internal.LoadConfigFiles(k); err != nil {
		internal.LogWarn(fmt.Sprintf("Warning: failed to load config files from configs/: %v", err))
	}
	if err := k.Load(confmap.Provider(internal.EnvConfigMap(), "."), nil); err != nil {
		return nil, err
	}
	if err := k.UnmarshalWithConf("database", dbConfig, koanf.UnmarshalConf{Tag: "mapstructure"}); err != nil {
		return nil, err
	}

	if dbConfig.Default == "" {
		dbConfig.Default = "postgres"
	}

	return dbConfig, nil
}

// SetLogger 允许高层注入日志接口以避免循环依赖
func SetLogger(infoLog, warnLog func(string)) {
	internal.LogInfo = infoLog
	internal.LogWarn = warnLog
}
