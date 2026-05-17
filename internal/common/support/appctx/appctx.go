package appctx

import (
	"sync"

	"fiber-starter/configs"
	authContracts "fiber-starter/internal/providers/auth/Contracts"
	cacheContracts "fiber-starter/internal/providers/cache/Contracts"
	configContracts "fiber-starter/internal/providers/config/Contracts"
	databaseContracts "fiber-starter/internal/providers/database/Contracts"
	hashContracts "fiber-starter/internal/providers/hash/Contracts"
	i18nContracts "fiber-starter/internal/providers/i18n/Contracts"
	loggingContracts "fiber-starter/internal/providers/logging/Contracts"
	mailContracts "fiber-starter/internal/providers/mail/Contracts"
	notificationContracts "fiber-starter/internal/providers/notification/Contracts"
	queueContracts "fiber-starter/internal/providers/queue/Contracts"
	ratelimiterContracts "fiber-starter/internal/providers/ratelimiter/Contracts"
	scheduleContracts "fiber-starter/internal/providers/schedule/Contracts"
	searchContracts "fiber-starter/internal/providers/search/Contracts"
	storageContracts "fiber-starter/internal/providers/storage/Contracts"
	validationContracts "fiber-starter/internal/providers/validation/Contracts"
)

// HealthChecker is an interface for components that can be checked for health.
type HealthChecker interface {
	HealthCheck() error
}

type ConfigProvider interface {
	AppConfig() *configs.Config
	ConfigRepository() configContracts.Repository
}

type DatabaseProvider interface {
	DatabaseManager() databaseContracts.Manager
	ConnectionValue() databaseContracts.Connection
}

type CacheProvider interface {
	CacheManagerValue() cacheContracts.Manager
	CacheStore() cacheContracts.Store
}

type AuthProvider interface {
	AuthManager() authContracts.Manager
}

type MailProvider interface {
	MailManagerValue() mailContracts.Manager
	EmailServiceValue() mailContracts.Mailer
}

type QueueProvider interface {
	QueueManagerValue() queueContracts.Manager
	QueueServiceValue() queueContracts.Queue
}

type ScheduleProvider interface {
	ScheduleManagerValue() scheduleContracts.Manager
	ScheduleServiceValue() scheduleContracts.Scheduler
}

type SearchProvider interface {
	SearchManagerValue() searchContracts.Manager
	SearchServiceValue() searchContracts.Engine
}

type StorageProvider interface {
	StorageValue() storageContracts.StorageManager
}

type HashProvider interface {
	HashService() hashContracts.Hasher
}

type NotificationProvider interface {
	NotificationService() notificationContracts.Dispatcher
}

type I18nProvider interface {
	TranslatorService() i18nContracts.Translator
}

type ValidationProvider interface {
	ValidationService() validationContracts.Factory
}

type LoggingProvider interface {
	LogService() loggingContracts.Logger
}

type RateLimiterProvider interface {
	RateLimiterService() ratelimiterContracts.Limiter
}

// Application is the shared application container contract used by facade packages.
type Application interface {
	ConfigProvider
	DatabaseProvider
	CacheProvider
	AuthProvider
	MailProvider
	QueueProvider
	ScheduleProvider
	SearchProvider
	StorageProvider
	HashProvider
	NotificationProvider
	I18nProvider
	ValidationProvider
	LoggingProvider
	RateLimiterProvider
}

var (
	mu       sync.RWMutex
	instance Application
)

// Set stores the global application container.
func Set(app Application) Application {
	mu.Lock()
	defer mu.Unlock()

	instance = app
	return instance
}

// App returns the global application container.
func App() Application {
	mu.RLock()
	defer mu.RUnlock()

	return instance
}

// Clear removes the global application container when the same instance is being closed.
func Clear(app Application) {
	mu.Lock()
	defer mu.Unlock()

	if instance == app {
		instance = nil
	}
}
