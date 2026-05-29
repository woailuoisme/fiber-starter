package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func EnvConfigMap() map[string]any {
	m := map[string]any{}
	set := func(envKey, target string) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			m[target] = value
		}
	}

	// 1. Direct mappings (leveraging Koanf's Unmarshal for type conversion)
	mappings := map[string]string{
		// App
		"APP_NAME":                      "app.name",
		"APP_ENV":                       "app.env",
		"APP_DEBUG":                     "app.debug",
		"APP_PORT":                      "app.port",
		"APP_HOST":                      "app.host",
		"APP_TIMEZONE":                  "app.timezone",
		"APP_URL":                       "app.url",
		"APP_FIBER_PREFORK":             "app.fiber.prefork",
		"APP_FIBER_SERVER_HEADER":       "app.fiber.server_header",
		"APP_FIBER_BODY_LIMIT":          "app.fiber.body_limit",
		"APP_FIBER_CONCURRENCY":         "app.fiber.concurrency",
		"APP_FIBER_READ_BUFFER_SIZE":    "app.fiber.read_buffer_size",
		"APP_FIBER_READ_TIMEOUT":        "app.fiber.read_timeout",
		"APP_FIBER_WRITE_TIMEOUT":       "app.fiber.write_timeout",
		"APP_FIBER_IDLE_TIMEOUT":        "app.fiber.idle_timeout",
		"APP_FIBER_TRUST_PROXY":         "app.fiber.trust_proxy",
		"APP_FIBER_PROXY_HEADER":        "app.fiber.proxy_header",
		"APP_FIBER_STREAM_REQUEST_BODY": "app.fiber.stream_request_body",
		"APP_FIBER_IMMUTABLE":           "app.fiber.immutable",

		// Redis & Queue
		"REDIS_HOST":        "redis.host",
		"REDIS_PORT":        "redis.port",
		"REDIS_PASSWORD":    "redis.password",
		"REDIS_DB":          "redis.db",
		"QUEUE_CONCURRENCY": "queue.concurrency",

		// Cache
		"CACHE_DRIVER":  "cache.driver",
		"CACHE_PREFIX":  "cache.prefix",
		"CACHE_DEFAULT": "cache.default",
		"CACHE_TTL":     "cache.ttl",

		// JWT
		"JWT_SECRET":          "jwt.secret",
		"JWT_EXPIRATION_TIME": "jwt.expiration_time",
		"JWT_REFRESH_TIME":    "jwt.refresh_time",
		"JWT_EXPIRE_HOURS":    "jwt.expire_hours",
		"JWT_ISSUER":          "jwt.issuer",

		// Authorization
		"AUTHORIZATION_MODEL_FILE":  "authorization.model_file",
		"AUTHORIZATION_POLICY_FILE": "authorization.policy_file",

		// Logger
		"LOGGER_LEVEL":       "logger.level",
		"LOGGER_FORMAT":      "logger.format",
		"LOGGER_OUTPUT":      "logger.output",
		"LOGGER_MAX_SIZE":    "logger.max_size",
		"LOGGER_MAX_AGE":     "logger.max_age",
		"LOGGER_MAX_BACKUPS": "logger.max_backups",
		"LOGGER_COMPRESS":    "logger.compress",

		// Mail
		"MAIL_MAILER":       "mail.default",
		"MAIL_FROM_NAME":    "mail.from_name",
		"MAIL_FROM_ADDRESS": "mail.from_address",
		"MAIL_REPLY_TO":     "mail.reply_to",
		"MAIL_HOST":         "mail.host",
		"MAIL_PORT":         "mail.port",
		"MAIL_USERNAME":     "mail.username",
		"MAIL_PASSWORD":     "mail.password",
		"MAIL_ENCRYPTION":   "mail.encryption",
		"RESEND_API_KEY":    "mail.api_key",

		// Notification
		"NOTIFICATION_GOTIFY_ENABLED":      "notification.gotify.enabled",
		"NOTIFICATION_GOTIFY_URL":          "notification.gotify.url",
		"NOTIFICATION_GOTIFY_TOKEN":        "notification.gotify.token",
		"NOTIFICATION_GOTIFY_TITLE":        "notification.gotify.title",
		"NOTIFICATION_GOTIFY_PRIORITY":     "notification.gotify.priority",
		"NOTIFICATION_TELEGRAM_ENABLED":    "notification.telegram.enabled",
		"NOTIFICATION_TELEGRAM_API_URL":    "notification.telegram.api_url",
		"NOTIFICATION_TELEGRAM_BOT_TOKEN":  "notification.telegram.bot_token",
		"NOTIFICATION_TELEGRAM_CHAT_ID":    "notification.telegram.chat_id",
		"NOTIFICATION_TELEGRAM_PARSE_MODE": "notification.telegram.parse_mode",

		// Storage
		"STORAGE_DRIVER":                "storage.driver",
		"STORAGE_DATABASE":              "storage.database",
		"STORAGE_RESET":                 "storage.reset",
		"STORAGE_GC_INTERVAL":           "storage.gc_interval",
		"GARAGE_ENDPOINT":               "storage.garage.endpoint",
		"GARAGE_ACCESS_KEY_ID":          "storage.garage.access_key_id",
		"GARAGE_SECRET_ACCESS_KEY":      "storage.garage.secret_access_key",
		"GARAGE_USE_SSL":                "storage.garage.use_ssl",
		"GARAGE_BUCKET":                 "storage.garage.bucket",
		"GARAGE_REGION":                 "storage.garage.region",
		"MINIO_ENDPOINT":                "storage.minio.endpoint",
		"MINIO_ACCESS_KEY_ID":           "storage.minio.access_key_id",
		"MINIO_SECRET_ACCESS_KEY":       "storage.minio.secret_access_key",
		"MINIO_USE_SSL":                 "storage.minio.use_ssl",
		"MINIO_BUCKET":                  "storage.minio.bucket",
		"MINIO_REGION":                  "storage.minio.region",
		"S3_ACCESS_KEY_ID":              "storage.s3.access_key_id",
		"S3_SECRET_ACCESS_KEY":          "storage.s3.secret_access_key",
		"S3_REGION":                     "storage.s3.region",
		"S3_BUCKET":                     "storage.s3.bucket",
		"S3_ENDPOINT":                   "storage.s3.endpoint",
		"R2_ACCESS_KEY_ID":              "storage.r2.access_key_id",
		"R2_SECRET_ACCESS_KEY":          "storage.r2.secret_access_key",
		"R2_REGION":                     "storage.r2.region",
		"R2_BUCKET":                     "storage.r2.bucket",
		"R2_ENDPOINT":                   "storage.r2.endpoint",
		"OSS_ACCESS_KEY_ID":             "storage.oss.access_key_id",
		"OSS_SECRET_ACCESS_KEY":         "storage.oss.secret_access_key",
		"OSS_REGION":                    "storage.oss.region",
		"OSS_BUCKET":                    "storage.oss.bucket",
		"OSS_ENDPOINT":                  "storage.oss.endpoint",
		"STORAGE_LOCAL_ROOT":            "storage.local.root",
		"STORAGE_LOCAL_URL":             "storage.local.url",
		"STORAGE_PUBLIC_ROOT":           "storage.public.root",
		"STORAGE_PUBLIC_URL":            "storage.public.url",
		"MEDIA_LIBRARY_CONVERSION_MODE": "media_library.conversion_mode",
		"MEDIA_LIBRARY_QUEUE":           "media_library.queue",
		"BACKUP_DISK":                   "backup.disk",
		"BACKUP_PATH":                   "backup.path",
		"BACKUP_TEMP_PATH":              "backup.temp_path",
		"BACKUP_NOTIFICATIONS_ENABLED":  "backup.notifications.enabled",
		"BACKUP_NOTIFY_SUCCESS":         "backup.notifications.notify_success",
		"BACKUP_NOTIFICATION_CHANNELS":  "backup.notifications.channels",
		"BACKUP_NOTIFICATION_MAIL_TO":   "backup.notifications.mail_to",
		"BACKUP_PG_DUMP_BINARY":         "backup.binaries.pg_dump",
		"BACKUP_PSQL_BINARY":            "backup.binaries.psql",
		"BACKUP_SQLITE3_BINARY":         "backup.binaries.sqlite3",

		// WebSocket
		"WEBSOCKET_PORT":               "websocket.port",
		"WEBSOCKET_PATH":               "websocket.path",
		"WEBSOCKET_AUTH_PATH":          "websocket.auth_path",
		"WEBSOCKET_APP_ID":             "websocket.app_id",
		"WEBSOCKET_APP_KEY":            "websocket.app_key",
		"WEBSOCKET_APP_SECRET":         "websocket.app_secret",
		"WEBSOCKET_BUS_MODE":           "websocket.bus_mode",
		"WEBSOCKET_REDIS_PREFIX":       "websocket.redis_prefix",
		"WEBSOCKET_WRITE_QUEUE_SIZE":   "websocket.write_queue_size",
		"WEBSOCKET_MAX_MESSAGE_SIZE":   "websocket.max_message_size",
		"WEBSOCKET_HEARTBEAT_INTERVAL": "websocket.heartbeat_interval",
		"WEBSOCKET_PRESENCE_TTL":       "websocket.presence_ttl",
		"REVERB_APP_ID":                "websocket.app_id",
		"REVERB_APP_KEY":               "websocket.app_key",
		"REVERB_APP_SECRET":            "websocket.app_secret",
		"REVERB_PATH":                  "websocket.path",

		// Payment
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

		// Business
		"ORDER_PAYMENT_TIMEOUT":       "business.order.payment_timeout",
		"ORDER_PICKUP_TIMEOUT":        "business.order.pickup_timeout",
		"DEVICE_CHANNEL_COUNT":        "business.device.channel_count",
		"DEVICE_CHANNEL_MAX_CAPACITY": "business.device.channel_max_capacity",

		// Security
		"SECURITY_CORS_ALLOWED_ORIGINS": "security.cors.allowed_origins",
		"SECURITY_CORS_ALLOWED_METHODS": "security.cors.allowed_methods",
		"SECURITY_CORS_ALLOWED_HEADERS": "security.cors.allowed_headers",
		"SECURITY_RATE_LIMIT_MAX":       "security.rate_limit.max",
		"SECURITY_RATE_LIMIT_WINDOW":    "security.rate_limit.window",
		"SECURITY_LOAD_SHED_ENABLED":    "security.load_shed.enabled",
		"SECURITY_LOAD_SHED_LOWER":      "security.load_shed.lower_threshold",
		"SECURITY_LOAD_SHED_UPPER":      "security.load_shed.upper_threshold",
		"LOG_VIEWER_ENABLED":            "security.log_viewer.enabled",
		"LOG_VIEWER_PATH":               "security.log_viewer.path",
		"LOG_VIEWER_USERNAME":           "security.log_viewer.username",
		"LOG_VIEWER_PASSWORD":           "security.log_viewer.password",
		"LOG_VIEWER_ALLOW_DELETE":       "security.log_viewer.allow_delete",

		// Dependency criticality
		"SERVICE_DATABASE_CRITICAL": "services.dependencies.database.critical",
		"SERVICE_CACHE_CRITICAL":    "services.dependencies.cache.critical",
		"SERVICE_MAIL_CRITICAL":     "services.dependencies.mail.critical",
		"SERVICE_QUEUE_CRITICAL":    "services.dependencies.queue.critical",
		"SERVICE_SEARCH_CRITICAL":   "services.dependencies.search.critical",
		"SERVICE_STORAGE_CRITICAL":  "services.dependencies.storage.critical",
		"SERVICE_REALTIME_CRITICAL": "services.dependencies.realtime.critical",
		"SERVICE_BACKUP_CRITICAL":   "services.dependencies.backup.critical",

		// I18n
		"I18N_ENABLED":          "i18n.enabled",
		"I18N_DEFAULT_LANGUAGE": "i18n.default_language",
		"I18N_LANGUAGE_DIR":     "i18n.language_dir",
		"I18N_COOKIE_NAME":      "i18n.cookie_name",
		"I18N_COOKIE_MAX_AGE":   "i18n.cookie_max_age",

		// Meilisearch & OTEL
		"SEARCH_DRIVER":           "search.default",
		"MEILISEARCH_HOST":        "search.host",
		"MEILISEARCH_API_KEY":     "search.api_key",
		"OTEL_TRACE_ENABLED":      "otel.trace_enabled",
		"OTEL_SERVICE_NAME":       "otel.service_name",
		"OTEL_EXPORTER_TYPE":      "otel.exporter_type",
		"OTEL_ENDPOINT":           "otel.endpoint",
		"OTEL_OTLP_INSECURE":      "otel.otlp_insecure",
		"OTEL_TRACE_SAMPLE_RATIO": "otel.trace_sample_ratio",
		"OTEL_METRICS_ENABLED":    "otel.metrics_enabled",
		"OTEL_METRICS_PATH":       "otel.metrics_path",

		// Database common
		"DB_POOL_MAX_OPEN_CONNS":     "database.pool.max_open_conns",
		"DB_POOL_MAX_IDLE_CONNS":     "database.pool.max_idle_conns",
		"DB_POOL_CONN_MAX_LIFETIME":  "database.pool.conn_max_lifetime",
		"DB_POOL_CONN_MAX_IDLE_TIME": "database.pool.conn_max_idle_time",
		"DB_RETRY_ATTEMPTS":          "database.retry_attempts",
		"DB_RETRY_BACKOFF_MS":        "database.retry_backoff_ms",
		"DB_RETRY_BACKOFF_FACTOR":    "database.retry_backoff_factor",
		"DB_RETRY_ON_OPEN":           "database.retry_on_open",
		"DB_RETRY_ON_QUERY":          "database.retry_on_query",
		"DB_LOG_QUERIES":             "database.log_queries",
		"DB_SLOW_QUERY_THRESHOLD_MS": "database.slow_query_threshold_ms",
		"DB_MIGRATIONS_TABLE":        "database.migrations.table",
		"DB_MIGRATIONS_PATH":         "database.migrations.path",
		"DB_SEEDERS_PATH":            "database.seeders.path",
		"EXCEL_TEMP_PATH":            "excel.temp_path",
	}

	for env, key := range mappings {
		set(env, key)
	}

	// Backward compatibility: LOG_CHANNEL was the previous name for logger output.
	// LOGGER_OUTPUT is canonical and wins when both are present.
	if _, ok := m["logger.output"]; !ok {
		set("LOG_CHANNEL", "logger.output")
	}

	// Laravel-compatible aliases take precedence over older project-specific names.
	set("CACHE_STORE", "cache.driver")
	set("FILESYSTEM_DISK", "storage.driver")
	set("AWS_ACCESS_KEY_ID", "storage.s3.access_key_id")
	set("AWS_SECRET_ACCESS_KEY", "storage.s3.secret_access_key")
	set("AWS_DEFAULT_REGION", "storage.s3.region")
	set("AWS_BUCKET", "storage.s3.bucket")
	set("AWS_ENDPOINT", "storage.s3.endpoint")

	// 2. Specialized Database Logic
	dbDefault := strings.ToLower(strings.TrimSpace(os.Getenv("DB_CONNECTION")))
	if dbDefault == "" {
		dbDefault = "postgres"
	}
	m["database.default"] = dbDefault

	// Sync main DB envs to both postgres and pgsql connection aliases
	for _, conn := range []string{"postgres", "pgsql"} {
		set("DB_HOST", fmt.Sprintf("database.connections.%s.host", conn))
		set("DB_PORT", fmt.Sprintf("database.connections.%s.port", conn))
		set("DB_DATABASE", fmt.Sprintf("database.connections.%s.database", conn))
		set("DB_USERNAME", fmt.Sprintf("database.connections.%s.username", conn))
		set("DB_PASSWORD", fmt.Sprintf("database.connections.%s.password", conn))
		set("DB_CHARSET", fmt.Sprintf("database.connections.%s.charset", conn))
		set("DB_COLLATION", fmt.Sprintf("database.connections.%s.collation", conn))
		set("DB_PREFIX", fmt.Sprintf("database.connections.%s.prefix", conn))
		set("DB_STRICT", fmt.Sprintf("database.connections.%s.strict", conn))
		set("DB_TIMEZONE", fmt.Sprintf("database.connections.%s.timezone", conn))
		set("DB_SCHEMA", fmt.Sprintf("database.connections.%s.schema", conn))
		set("DB_SSLMODE", fmt.Sprintf("database.connections.%s.sslmode", conn))
	}

	for _, conn := range []string{"postgres", "pgsql", "sqlite"} {
		setConnTuningEnv(m, conn)
	}

	if dbDefault == "sqlite" {
		set("DB_DATABASE", "database.connections.sqlite.database")
		set("DB_SQLITE_DATABASE", "database.connections.sqlite.database")
	}

	// 3. Complex parsing (Slices)
	setBool := func(envKey, target string) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			if parsed, err := strconv.ParseBool(value); err == nil {
				m[target] = parsed
			}
		}
	}
	setFloat := func(envKey, target string) {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			if parsed, err := strconv.ParseFloat(value, 64); err == nil {
				m[target] = parsed
			}
		}
	}

	setBool("OTEL_TRACE_ENABLED", "otel.trace_enabled")
	setBool("OTEL_METRICS_ENABLED", "otel.metrics_enabled")
	setBool("OTEL_OTLP_INSECURE", "otel.otlp_insecure")
	setFloat("OTEL_TRACE_SAMPLE_RATIO", "otel.trace_sample_ratio")

	if val := strings.TrimSpace(os.Getenv("I18N_SUPPORTED_LANGUAGES")); val != "" {
		supported := splitCommaList(val)
		if len(supported) > 0 {
			m["i18n.supported_languages"] = supported
		}
	}

	return m
}

func splitCommaList(value string) []string {
	var parts []string
	for _, part := range strings.Split(value, ",") {
		if s := strings.TrimSpace(part); s != "" {
			parts = append(parts, s)
		}
	}
	return parts
}

func setConnTuningEnv(m map[string]any, conn string) {
	upper := strings.ToUpper(strings.TrimSpace(conn))
	if upper == "" {
		return
	}

	set := func(env, key string) {
		if val := strings.TrimSpace(os.Getenv(env)); val != "" {
			m[key] = val
		}
	}

	set("DB_"+upper+"_RETRY_ATTEMPTS", fmt.Sprintf("database.connections.%s.retry_attempts", conn))
	set("DB_"+upper+"_RETRY_BACKOFF_MS", fmt.Sprintf("database.connections.%s.retry_backoff_ms", conn))
	set("DB_"+upper+"_RETRY_BACKOFF_FACTOR", fmt.Sprintf("database.connections.%s.retry_backoff_factor", conn))
	set("DB_"+upper+"_RETRY_ON_OPEN", fmt.Sprintf("database.connections.%s.retry_on_open", conn))
	set("DB_"+upper+"_RETRY_ON_QUERY", fmt.Sprintf("database.connections.%s.retry_on_query", conn))
	set("DB_"+upper+"_LOG_QUERIES", fmt.Sprintf("database.connections.%s.log_queries", conn))
	set("DB_"+upper+"_SLOW_QUERY_THRESHOLD_MS", fmt.Sprintf("database.connections.%s.slow_query_threshold_ms", conn))
}
