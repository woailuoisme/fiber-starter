package internal

type HashConfig struct {
	Driver string           `mapstructure:"driver"`
	Bcrypt BcryptHashConfig `mapstructure:"bcrypt"`
	Argon2 Argon2HashConfig `mapstructure:"argon2"`
}

type BcryptHashConfig struct {
	Rounds int `mapstructure:"rounds"`
}

type Argon2HashConfig struct {
	Memory      uint32 `mapstructure:"memory"`
	Iterations  uint32 `mapstructure:"iterations"`
	Parallelism uint8  `mapstructure:"parallelism"`
}

type MailConfig struct {
	Enabled     bool            `mapstructure:"enabled"`
	Default     string          `mapstructure:"default"`
	FromName    string          `mapstructure:"from_name"`
	FromAddress string          `mapstructure:"from_address"`
	ReplyTo     string          `mapstructure:"reply_to"`
	APIKey      string          `mapstructure:"api_key"`
	Host        string          `mapstructure:"host"`
	Port        int             `mapstructure:"port"`
	Username    string          `mapstructure:"username"`
	Password    string          `mapstructure:"password"`
	Encryption  string          `mapstructure:"encryption"`
	Theme       MailThemeConfig `mapstructure:"theme"`
}

type MailThemeConfig struct {
	PrimaryColor string `mapstructure:"primary_color"`
	SuccessColor string `mapstructure:"success_color"`
	DangerColor  string `mapstructure:"danger_color"`
	WarningColor string `mapstructure:"warning_color"`
	BgColor      string `mapstructure:"bg_color"`
}

type NotificationConfig struct {
	Gotify   GotifyNotificationConfig   `mapstructure:"gotify"`
	Telegram TelegramNotificationConfig `mapstructure:"telegram"`
}

type GotifyNotificationConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	URL      string `mapstructure:"url"`
	Token    string `mapstructure:"token"`
	Title    string `mapstructure:"title"`
	Priority int    `mapstructure:"priority"`
}

type TelegramNotificationConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	APIURL    string `mapstructure:"api_url"`
	BotToken  string `mapstructure:"bot_token"`
	ChatID    string `mapstructure:"chat_id"`
	ParseMode string `mapstructure:"parse_mode"`
}

type StorageConfig struct {
	Enabled    bool                 `mapstructure:"enabled"`
	Driver     string               `mapstructure:"driver"`
	Database   string               `mapstructure:"database"`
	Reset      bool                 `mapstructure:"reset"`
	GCInterval int                  `mapstructure:"gc_interval"`
	Garage     *GarageStorageConfig `mapstructure:"garage"`
	MinIO      *GarageStorageConfig `mapstructure:"minio"`
	S3         *S3StorageConfig     `mapstructure:"s3"`
	R2         *S3StorageConfig     `mapstructure:"r2"`
	OSS        *S3StorageConfig     `mapstructure:"oss"`
	Local      *LocalStorageConfig  `mapstructure:"local"`
	Public     *LocalStorageConfig  `mapstructure:"public"`
}

type MediaLibraryConfig struct {
	ConversionMode string `mapstructure:"conversion_mode"`
	Queue          string `mapstructure:"queue"`
}

type BackupConfig struct {
	Disk          string                   `mapstructure:"disk"`
	Path          string                   `mapstructure:"path"`
	TempPath      string                   `mapstructure:"temp_path"`
	Notifications BackupNotificationConfig `mapstructure:"notifications"`
	Binaries      BackupBinaryConfig       `mapstructure:"binaries"`
}

type BackupNotificationConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	NotifySuccess bool     `mapstructure:"notify_success"`
	Channels      []string `mapstructure:"channels"`
	MailTo        string   `mapstructure:"mail_to"`
}

type BackupBinaryConfig struct {
	PgDump  string `mapstructure:"pg_dump"`
	Psql    string `mapstructure:"psql"`
	SQLite3 string `mapstructure:"sqlite3"`
}

type LocalStorageConfig struct {
	Root string `mapstructure:"root"`
	URL  string `mapstructure:"url"`
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

type AuthConfig struct {
	Default   string                    `mapstructure:"default" validate:"required"`
	Guards    map[string]GuardConfig    `mapstructure:"guards"`
	Providers map[string]ProviderConfig `mapstructure:"providers"`
}

type GuardConfig struct {
	Driver   string `mapstructure:"driver"`
	Provider string `mapstructure:"provider"`
}

type ProviderConfig struct {
	Driver string `mapstructure:"driver"`
	Table  string `mapstructure:"table"`
}

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Authorization AuthorizationConfig `mapstructure:"authorization"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Logger        LoggerConfig        `mapstructure:"logger"`
	Cache         CacheConfig         `mapstructure:"cache"`
	Mail          MailConfig          `mapstructure:"mail"`
	Notification  NotificationConfig  `mapstructure:"notification"`
	Queue         QueueConfig         `mapstructure:"queue"`
	Storage       StorageConfig       `mapstructure:"storage"`
	MediaLibrary  MediaLibraryConfig  `mapstructure:"media_library"`
	Backup        BackupConfig        `mapstructure:"backup"`
	WebSocket     WebSocketConfig     `mapstructure:"websocket"`
	Payment       PaymentConfig       `mapstructure:"payment"`
	Business      BusinessConfig      `mapstructure:"business"`
	Security      SecurityConfig      `mapstructure:"security"`
	I18n          I18nConfig          `mapstructure:"i18n"`
	Search        SearchConfig        `mapstructure:"search"`
	Hash          HashConfig          `mapstructure:"hash"`
	OTEL          OTELConfig          `mapstructure:"otel"`
	Limiter       LimiterConfig       `mapstructure:"limiter"`
	Services      ServicesConfig      `mapstructure:"services"`
	Excel         ExcelConfig         `mapstructure:"excel"`
}

type AuthorizationConfig struct {
	ModelFile  string `mapstructure:"model_file"`
	PolicyFile string `mapstructure:"policy_file"`
}

type ServicesConfig struct {
	Dependencies map[string]ServiceDependencyConfig `mapstructure:"dependencies"`
}

type ServiceDependencyConfig struct {
	Critical bool `mapstructure:"critical"`
}

type SearchConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Default string `mapstructure:"default"`
	Host    string `mapstructure:"host"`
	APIKey  string `mapstructure:"api_key"`
}

type AppConfig struct {
	Name     string      `mapstructure:"name" validate:"required"`
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
	Default              string                  `mapstructure:"default" validate:"required"`
	Connections          map[string]DBConnection `mapstructure:"connections"`
	Pool                 DBPoolConfig            `mapstructure:"pool"`
	RetryAttempts        int                     `mapstructure:"retry_attempts"`
	RetryBackoffMS       int                     `mapstructure:"retry_backoff_ms"`
	RetryBackoffFactor   float64                 `mapstructure:"retry_backoff_factor"`
	RetryOnOpen          bool                    `mapstructure:"retry_on_open"`
	RetryOnQuery         bool                    `mapstructure:"retry_on_query"`
	LogQueries           bool                    `mapstructure:"log_queries"`
	SlowQueryThresholdMS int                     `mapstructure:"slow_query_threshold_ms"`
	Read                 DBReadConfig            `mapstructure:"read"`
	Write                DBWriteConfig           `mapstructure:"write"`
	Migrations           DBMigrationConfig       `mapstructure:"migrations"`
	Seeders              DBSeederConfig          `mapstructure:"seeders"`
	Redis                DBRedisConfig           `mapstructure:"redis"`
}

type DBConnection struct {
	Driver               string            `mapstructure:"driver"`
	Host                 string            `mapstructure:"host"`
	Port                 string            `mapstructure:"port"`
	Database             string            `mapstructure:"database"`
	Username             string            `mapstructure:"username"`
	Password             string            `mapstructure:"password"`
	Charset              string            `mapstructure:"charset"`
	Collation            string            `mapstructure:"collation"`
	Prefix               string            `mapstructure:"prefix"`
	Strict               bool              `mapstructure:"strict"`
	Timezone             string            `mapstructure:"timezone"`
	Schema               string            `mapstructure:"schema"`
	SSLMode              string            `mapstructure:"sslmode"`
	Options              map[string]string `mapstructure:"options"`
	RetryAttempts        *int              `mapstructure:"retry_attempts"`
	RetryBackoffMS       *int              `mapstructure:"retry_backoff_ms"`
	RetryBackoffFactor   *float64          `mapstructure:"retry_backoff_factor"`
	RetryOnOpen          *bool             `mapstructure:"retry_on_open"`
	RetryOnQuery         *bool             `mapstructure:"retry_on_query"`
	LogQueries           *bool             `mapstructure:"log_queries"`
	SlowQueryThresholdMS *int              `mapstructure:"slow_query_threshold_ms"`
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
	Enabled  bool   `mapstructure:"enabled"`
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
	Enabled bool   `mapstructure:"enabled"`
	Driver  string `mapstructure:"driver"`
	Prefix  string `mapstructure:"prefix"`
	Default int    `mapstructure:"default"`
	TTL     int    `mapstructure:"ttl"`
}

type QueueConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	Concurrency int  `mapstructure:"concurrency"`
}

type WebSocketConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	Port              string `mapstructure:"port"`
	Path              string `mapstructure:"path"`
	AuthPath          string `mapstructure:"auth_path"`
	AppID             string `mapstructure:"app_id"`
	AppKey            string `mapstructure:"app_key"`
	AppSecret         string `mapstructure:"app_secret"`
	BusMode           string `mapstructure:"bus_mode"`
	RedisPrefix       string `mapstructure:"redis_prefix"`
	WriteQueueSize    int    `mapstructure:"write_queue_size"`
	MaxMessageSize    int    `mapstructure:"max_message_size"`
	HeartbeatInterval int    `mapstructure:"heartbeat_interval"`
	PresenceTTL       int    `mapstructure:"presence_ttl"`
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
	CORS           CORSConfig           `mapstructure:"cors"`
	RateLimit      RateLimitConfig      `mapstructure:"rate_limit"`
	LoadShed       LoadShedConfig       `mapstructure:"load_shed"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	LogViewer      LogViewerConfig      `mapstructure:"log_viewer"`
}

type LogViewerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Path        string `mapstructure:"path"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	AllowDelete bool   `mapstructure:"allow_delete"`
}

type CircuitBreakerConfig struct {
	Enabled               bool `mapstructure:"enabled"`
	FailureThreshold      int  `mapstructure:"failure_threshold"`
	Timeout               int  `mapstructure:"timeout"` // Timeout in seconds
	SuccessThreshold      int  `mapstructure:"success_threshold"`
	HalfOpenMaxConcurrent int  `mapstructure:"half_open_max_concurrent"`
}

type LoadShedConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	LowerThreshold float64 `mapstructure:"lower_threshold"`
	UpperThreshold float64 `mapstructure:"upper_threshold"`
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

type LimiterConfig struct {
	Default    string                     `mapstructure:"default"`
	Strategies map[string]RateLimitConfig `mapstructure:"strategies"`
}

type OTELConfig struct {
	TraceEnabled     bool    `mapstructure:"trace_enabled"` // 是否启用链路追踪 (Tracer 引擎)
	ServiceName      string  `mapstructure:"service_name"`
	ExporterType     string  `mapstructure:"exporter_type"`      // exporter_type 可选值: stdout (开发调试输出至控制台), otlp (生产推荐推送至 OTel Collector)
	Endpoint         string  `mapstructure:"endpoint"`           // OTLP Collector 的 gRPC 地址，仅当 ExporterType 为 otlp 且开启 trace_enabled 时生效
	OTLPInsecure     bool    `mapstructure:"otlp_insecure"`      // 是否使用非 TLS OTLP gRPC 连接，适合本地 Collector
	TraceSampleRatio float64 `mapstructure:"trace_sample_ratio"` // Trace 采样比例，范围 0.0-1.0
	MetricsEnabled   bool    `mapstructure:"metrics_enabled"`    // 是否启用 Prometheus 指标监控 (/metrics)
	MetricsPath      string  `mapstructure:"metrics_path"`       // Prometheus 抓取指标的路由路径
}

type I18nConfig struct {
	Enabled            bool     `mapstructure:"enabled"`
	DefaultLanguage    string   `mapstructure:"default_language"`
	SupportedLanguages []string `mapstructure:"supported_languages"`
	LanguageDir        string   `mapstructure:"language_dir"`
	CookieName         string   `mapstructure:"cookie_name"`
	CookieMaxAge       int      `mapstructure:"cookie_max_age"`
}

type ExcelConfig struct {
	TempPath string `mapstructure:"temp_path"`
}
