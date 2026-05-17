package providers

import (
	"errors"
	"fmt"

	auth "fiber-starter/app/Providers/Auth"
	authContracts "fiber-starter/app/Providers/Auth/Contracts"
	cache "fiber-starter/app/Providers/Cache"
	cacheContracts "fiber-starter/app/Providers/Cache/Contracts"
	config "fiber-starter/app/Providers/Config"
	configContracts "fiber-starter/app/Providers/Config/Contracts"
	database "fiber-starter/app/Providers/Database"
	databaseContracts "fiber-starter/app/Providers/Database/Contracts"
	hash "fiber-starter/app/Providers/Hash"
	hashContracts "fiber-starter/app/Providers/Hash/Contracts"
	i18n "fiber-starter/app/Providers/I18n"
	i18nContracts "fiber-starter/app/Providers/I18n/Contracts"
	logging "fiber-starter/app/Providers/Logging"
	loggingContracts "fiber-starter/app/Providers/Logging/Contracts"
	mail "fiber-starter/app/Providers/Mail"
	mailContracts "fiber-starter/app/Providers/Mail/Contracts"
	notification "fiber-starter/app/Providers/Notification"
	notificationContracts "fiber-starter/app/Providers/Notification/Contracts"
	queue "fiber-starter/app/Providers/Queue"
	queueContracts "fiber-starter/app/Providers/Queue/Contracts"
	ratelimiter "fiber-starter/app/Providers/RateLimiter"
	ratelimiterContracts "fiber-starter/app/Providers/RateLimiter/Contracts"
	realtime "fiber-starter/app/Providers/Realtime"
	realtimeContracts "fiber-starter/app/Providers/Realtime/Contracts"
	schedule "fiber-starter/app/Providers/Schedule"
	scheduleContracts "fiber-starter/app/Providers/Schedule/Contracts"
	search "fiber-starter/app/Providers/Search"
	searchContracts "fiber-starter/app/Providers/Search/Contracts"
	storage "fiber-starter/app/Providers/Storage"
	storageContracts "fiber-starter/app/Providers/Storage/Contracts"
	validation "fiber-starter/app/Providers/Validation"
	validationContracts "fiber-starter/app/Providers/Validation/Contracts"
	"fiber-starter/app/Support/appctx"
	"fiber-starter/configs"
)

// Runtime holds the application infrastructure dependencies (Providers).
type Runtime struct {
	Config          *configs.Config
	ConfigRepo      configContracts.Repository
	Database        databaseContracts.Manager
	Connection      databaseContracts.Connection
	CacheManager    cacheContracts.Manager
	Cache           cacheContracts.Store
	Auth            authContracts.Manager
	MailManager     mailContracts.Manager
	EmailService    mailContracts.Mailer
	Realtime        realtimeContracts.Manager
	QueueManager    queueContracts.Manager
	QueueService    queueContracts.Queue
	ScheduleManager scheduleContracts.Manager
	ScheduleService scheduleContracts.Scheduler
	SearchManager   searchContracts.Manager
	SearchService   searchContracts.Engine
	Storage         storageContracts.StorageManager
	Hash            hashContracts.Hasher
	Notification    notificationContracts.Dispatcher
	Translator      i18nContracts.Translator
	Validation      validationContracts.Factory
	Log             loggingContracts.Logger
	RateLimiter     ratelimiterContracts.Limiter
}

// App returns the global application container instance.
func App() *Runtime {
	return appctx.App().(*Runtime)
}

// SetInstance sets the shared application container instance.
func SetInstance(rt *Runtime) *Runtime {
	appctx.Set(rt)
	return rt
}

// Build wires infrastructure providers explicitly and sets the global container.
func Build() (*Runtime, error) {
	cfg, k, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	rt := &Runtime{Config: cfg}

	// Phase 1: Register (Object creation, dependency wiring)
	configRepo, err := config.RegisterConfig(k)
	if err != nil {
		return nil, fmt.Errorf("register config provider: %w", err)
	}
	rt.ConfigRepo = configRepo

	_, translator, err := i18n.RegisterI18n(cfg)
	if err != nil {
		return nil, fmt.Errorf("register i18n provider: %w", err)
	}
	rt.Translator = translator

	databaseManager, connection, err := database.RegisterDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("register database provider: %w", err)
	}
	rt.Database = databaseManager
	rt.Connection = connection

	cacheManager, cacheStore, err := cache.RegisterCache(cfg)
	if err != nil {
		return nil, fmt.Errorf("register cache provider: %w", err)
	}
	rt.CacheManager = cacheManager
	rt.Cache = cacheStore

	hashManager, err := hash.RegisterHash(cfg)
	if err != nil {
		return nil, fmt.Errorf("register hash provider: %w", err)
	}
	rt.Hash = hashManager

	authManager, err := auth.Register(cfg, connection, hashManager)
	if err != nil {
		return nil, fmt.Errorf("register auth provider: %w", err)
	}
	rt.Auth = authManager

	mailManager, emailService, err := mail.Register(cfg)
	if err != nil {
		return nil, fmt.Errorf("register mail provider: %w", err)
	}
	rt.MailManager = mailManager
	rt.EmailService = emailService

	realtimeManager, err := realtime.RegisterRealtime(cfg)
	if err != nil {
		return nil, fmt.Errorf("register realtime provider: %w", err)
	}
	rt.Realtime = realtimeManager

	queueManager, queueService, err := queue.RegisterQueue(cfg)
	if err != nil {
		return nil, fmt.Errorf("register queue provider: %w", err)
	}
	rt.QueueManager = queueManager
	rt.QueueService = queueService

	scheduleManager, scheduleService, err := schedule.RegisterSchedule(cfg)
	if err != nil {
		return nil, fmt.Errorf("register schedule provider: %w", err)
	}
	rt.ScheduleManager = scheduleManager
	rt.ScheduleService = scheduleService

	searchManager, searchService, err := search.Register(cfg)
	if err != nil {
		return nil, fmt.Errorf("register search provider: %w", err)
	}
	rt.SearchManager = searchManager
	rt.SearchService = searchService

	storageManager, err := storage.Register(cfg)
	if err != nil {
		return nil, fmt.Errorf("register storage provider: %w", err)
	}
	rt.Storage = storageManager

	notificationManager, notificationService, err := notification.RegisterNotification(emailService)
	if err != nil {
		return nil, fmt.Errorf("register notification provider: %w", err)
	}
	rt.Notification = notificationService

	if err := notification.RegisterConfiguredChannels(cfg, notificationManager); err != nil {
		return nil, fmt.Errorf("register notification channels: %w", err)
	}

	validationService, err := validation.RegisterValidation(cfg)
	if err != nil {
		return nil, fmt.Errorf("register validation provider: %w", err)
	}
	rt.Validation = validationService

	loggingService, err := logging.Register(cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("register logging provider: %w", err)
	}
	rt.Log = loggingService

	rateLimiterService, err := ratelimiter.Register(cfg.Limiter)
	if err != nil {
		return nil, fmt.Errorf("register rate limiter provider: %w", err)
	}
	rt.RateLimiter = rateLimiterService

	return SetInstance(rt), nil
}

// Boot performs IO connection establishment and resource initialization.
func (rt *Runtime) Boot() error {
	return nil
}

// Close shuts down runtime resources in reverse initialization order.
func (rt *Runtime) Close() error {
	if rt == nil {
		return nil
	}

	defer appctx.Clear(rt)

	var errs []error
	if rt.Storage != nil {
		if err := rt.Storage.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.SearchManager != nil {
		if err := rt.SearchManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.MailManager != nil {
		if err := rt.MailManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.Realtime != nil {
		if err := rt.Realtime.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.ScheduleManager != nil {
		if err := rt.ScheduleManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.QueueManager != nil {
		if err := rt.QueueManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.CacheManager != nil {
		if err := rt.CacheManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if rt.Database != nil {
		if err := rt.Database.CloseAll(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (rt *Runtime) AppConfig() *configs.Config { return rt.Config }

func (rt *Runtime) ConfigRepository() configContracts.Repository { return rt.ConfigRepo }

func (rt *Runtime) DatabaseManager() databaseContracts.Manager { return rt.Database }

func (rt *Runtime) ConnectionValue() databaseContracts.Connection { return rt.Connection }

func (rt *Runtime) CacheManagerValue() cacheContracts.Manager { return rt.CacheManager }

func (rt *Runtime) CacheStore() cacheContracts.Store { return rt.Cache }

func (rt *Runtime) AuthManager() authContracts.Manager { return rt.Auth }

func (rt *Runtime) MailManagerValue() mailContracts.Manager { return rt.MailManager }

func (rt *Runtime) EmailServiceValue() mailContracts.Mailer { return rt.EmailService }

func (rt *Runtime) QueueManagerValue() queueContracts.Manager { return rt.QueueManager }

func (rt *Runtime) QueueServiceValue() queueContracts.Queue { return rt.QueueService }

func (rt *Runtime) ScheduleManagerValue() scheduleContracts.Manager { return rt.ScheduleManager }

func (rt *Runtime) ScheduleServiceValue() scheduleContracts.Scheduler { return rt.ScheduleService }

func (rt *Runtime) SearchManagerValue() searchContracts.Manager { return rt.SearchManager }

func (rt *Runtime) SearchServiceValue() searchContracts.Engine { return rt.SearchService }

func (rt *Runtime) StorageValue() storageContracts.StorageManager { return rt.Storage }

func (rt *Runtime) HashService() hashContracts.Hasher { return rt.Hash }

func (rt *Runtime) NotificationService() notificationContracts.Dispatcher { return rt.Notification }

func (rt *Runtime) TranslatorService() i18nContracts.Translator { return rt.Translator }

func (rt *Runtime) ValidationService() validationContracts.Factory { return rt.Validation }

func (rt *Runtime) LogService() loggingContracts.Logger { return rt.Log }

func (rt *Runtime) RateLimiterService() ratelimiterContracts.Limiter { return rt.RateLimiter }
