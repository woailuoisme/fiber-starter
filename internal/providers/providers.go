package providers

import (
	"errors"
	"fmt"

	"lfiber/configs"
	artisan "lfiber/internal/providers/artisan"
	artisanContracts "lfiber/internal/providers/artisan/contracts"
	auth "lfiber/internal/providers/auth"
	authContracts "lfiber/internal/providers/auth/contracts"
	authorization "lfiber/internal/providers/authorization"
	authorizationContracts "lfiber/internal/providers/authorization/contracts"
	backupProvider "lfiber/internal/providers/backup"
	cache "lfiber/internal/providers/cache"
	cacheContracts "lfiber/internal/providers/cache/contracts"
	config "lfiber/internal/providers/config"
	configContracts "lfiber/internal/providers/config/contracts"
	database "lfiber/internal/providers/database"
	databaseContracts "lfiber/internal/providers/database/contracts"
	hash "lfiber/internal/providers/hash"
	hashContracts "lfiber/internal/providers/hash/contracts"
	i18n "lfiber/internal/providers/i18n"
	i18nContracts "lfiber/internal/providers/i18n/contracts"
	logging "lfiber/internal/providers/logging"
	loggingContracts "lfiber/internal/providers/logging/contracts"
	mail "lfiber/internal/providers/mail"
	mailContracts "lfiber/internal/providers/mail/contracts"
	notification "lfiber/internal/providers/notification"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	queue "lfiber/internal/providers/queue"
	queueContracts "lfiber/internal/providers/queue/contracts"
	ratelimiter "lfiber/internal/providers/ratelimiter"
	ratelimiterContracts "lfiber/internal/providers/ratelimiter/contracts"
	realtime "lfiber/internal/providers/realtime"
	realtimeContracts "lfiber/internal/providers/realtime/contracts"
	schedule "lfiber/internal/providers/schedule"
	scheduleContracts "lfiber/internal/providers/schedule/contracts"
	storage "lfiber/internal/providers/storage"
	storageContracts "lfiber/internal/providers/storage/contracts"
	helpers "lfiber/internal/support"
	"lfiber/internal/support/appctx"
	backup "lfiber/pkg/backup"
	lock "lfiber/pkg/lock"
	medialibrary "lfiber/pkg/medialibrary"
	search "lfiber/pkg/search"
)

// Runtime holds the application infrastructure dependencies (Providers).
type Runtime struct {
	Config          *configs.Config
	Artisan         artisanContracts.Artisan
	ConfigRepo      configContracts.Repository
	Database        databaseContracts.Manager
	Connection      databaseContracts.Connection
	CacheManager    cacheContracts.Manager
	Cache           cacheContracts.Store
	Auth            authContracts.Manager
	Authorization   authorizationContracts.Authorizer
	MailManager     mailContracts.Manager
	EmailService    mailContracts.Mailer
	Realtime        realtimeContracts.Manager
	Locker          lock.Locker
	QueueManager    queueContracts.Manager
	QueueService    queueContracts.Queue
	ScheduleManager scheduleContracts.Manager
	ScheduleService scheduleContracts.Scheduler
	SearchManager   search.Manager
	SearchService   search.Engine
	Storage         storageContracts.StorageManager
	Hash            hashContracts.Hasher
	Notification    notificationContracts.Dispatcher
	Translator      i18nContracts.Translator
	Log             loggingContracts.Logger
	RateLimiter     ratelimiterContracts.Limiter
	MediaLibrary    *medialibrary.Service
	Backup          *backup.Service
	Degraded        map[string]string
}

// App returns the global application container instance.
func App() *Runtime {
	return appctx.App().(*Runtime)
}

// SetInstance sets the shared application container instance.
func SetInstance(rt *Runtime) *Runtime {
	if rt == nil {
		appctx.Set(nil)
	} else {
		appctx.Set(rt)
	}
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

	rt := &Runtime{Config: cfg, Degraded: map[string]string{}}

	// Phase 1: Register (Object creation, dependency wiring)
	steps := []struct {
		name     string
		critical bool
		register func() error
	}{
		{"config", true, func() error {
			configRepo, err := config.RegisterConfig(k)
			if err == nil {
				rt.ConfigRepo = configRepo
			}
			return err
		}},
		{"artisan", true, func() error {
			artisanService, err := artisan.Register()
			if err == nil {
				rt.Artisan = artisanService
			}
			return err
		}},
		{"translator", true, func() error {
			_, translator, err := i18n.RegisterI18n(cfg)
			if err == nil {
				rt.Translator = translator
			}
			return err
		}},
		{"database", true, func() error {
			databaseManager, connection, err := database.RegisterDatabase(cfg)
			if err == nil {
				rt.Database = databaseManager
				rt.Connection = connection
			}
			return err
		}},
		{"cache", false, func() error {
			cacheManager, cacheStore, err := cache.RegisterCache(cfg)
			if err == nil {
				rt.CacheManager = cacheManager
				rt.Cache = cacheStore
			}
			return err
		}},
		{"hash", true, func() error {
			hashManager, err := hash.RegisterHash(cfg)
			if err == nil {
				rt.Hash = hashManager
			}
			return err
		}},
		{"auth", true, func() error {
			authManager, err := auth.Register(cfg, rt.Connection, rt.Hash)
			if err == nil {
				rt.Auth = authManager
			}
			return err
		}},
		{"authorization", true, func() error {
			authorizationService, err := authorization.Register(cfg.Authorization)
			if err == nil {
				rt.Authorization = authorizationService
			}
			return err
		}},
		{"mail", false, func() error {
			mailManager, emailService, err := mail.Register(cfg)
			if err == nil {
				rt.MailManager = mailManager
				rt.EmailService = emailService
			}
			return err
		}},
		{"realtime", false, func() error {
			realtimeManager, err := realtime.RegisterRealtime(cfg)
			if err == nil {
				rt.Realtime = realtimeManager
			}
			return err
		}},
		{"lock", false, func() error {
			lockerService, err := lock.Register(cfg)
			if err == nil {
				rt.Locker = lockerService
			}
			return err
		}},
		{"queue", false, func() error {
			queueManager, queueService, err := queue.RegisterQueue(cfg)
			if err == nil {
				rt.QueueManager = queueManager
				rt.QueueService = queueService
			}
			return err
		}},
		{"schedule", true, func() error {
			scheduleManager, scheduleService, err := schedule.RegisterSchedule(cfg)
			if err == nil {
				rt.ScheduleManager = scheduleManager
				rt.ScheduleService = scheduleService
			}
			return err
		}},
		{"search", false, func() error {
			searchManager, searchService, err := search.Register(cfg)
			if err == nil {
				rt.SearchManager = searchManager
				rt.SearchService = searchService
				search.SetDefaultManager(searchManager)
				search.SetDefaultEngine(searchService)
			}
			return err
		}},
		{"storage", false, func() error {
			storageManager, err := storage.Register(cfg)
			if err == nil {
				rt.Storage = storageManager
			}
			return err
		}},
		{"medialibrary", true, func() error {
			if rt.Connection == nil || rt.Storage == nil {
				return errors.New("database connection or storage is not registered")
			}
			defaultDisk := "local"
			if cfg.Storage.Driver != "" {
				defaultDisk = cfg.Storage.Driver
			}
			rt.MediaLibrary = medialibrary.NewServiceFromConnection(
				rt.Connection,
				rt.Storage,
				defaultDisk,
				medialibrary.WithConversionMode(cfg.MediaLibrary.ConversionMode),
				medialibrary.WithQueue(rt.QueueService, cfg.MediaLibrary.Queue),
			)
			return nil
		}},
		{"notification", true, func() error {
			notificationManager, notificationService, err := notification.RegisterNotification(rt.EmailService)
			if err == nil {
				rt.Notification = notificationService
				if err := notification.RegisterConfiguredChannels(cfg, notificationManager); err != nil {
					return err
				}
			}
			return err
		}},
		{"backup", false, func() error {
			if rt.Database == nil || rt.Storage == nil {
				return errors.New("database manager or storage is not registered")
			}
			backupService, err := backupProvider.Register(cfg, rt.Database, rt.Storage, rt.Notification)
			if err == nil {
				rt.Backup = backupService
			}
			return err
		}},
		{"logging", true, func() error {
			loggingService, err := logging.Register(cfg.Logger)
			if err == nil {
				rt.Log = loggingService
			}
			return err
		}},
		{"ratelimiter", true, func() error {
			rateLimiterService, err := ratelimiter.Register(cfg.Limiter)
			if err == nil {
				rt.RateLimiter = rateLimiterService
			}
			return err
		}},
	}

	for _, step := range steps {
		err := step.register()
		if err != nil {
			if step.critical || isDependencyCritical(cfg, step.name, false) {
				return nil, fmt.Errorf("register %s provider: %w", step.name, err)
			}
			rt.markDegraded(step.name, err)
		}
	}

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
		search.SetDefaultManager(nil)
		search.SetDefaultEngine(nil)
	}
	if rt.MailManager != nil {
		if err := rt.MailManager.Close(); err != nil {
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

func (rt *Runtime) markDegraded(name string, err error) {
	if rt.Degraded == nil {
		rt.Degraded = map[string]string{}
	}
	rt.Degraded[name] = helpers.RedactError(err)
}

func isDependencyCritical(cfg *configs.Config, name string, fallback bool) bool {
	if cfg == nil || cfg.Services.Dependencies == nil {
		return fallback
	}
	dependency, ok := cfg.Services.Dependencies[name]
	if !ok {
		return fallback
	}
	return dependency.Critical
}

func (rt *Runtime) DegradedProviders() map[string]string {
	if rt == nil || len(rt.Degraded) == 0 {
		return nil
	}

	degraded := make(map[string]string, len(rt.Degraded))
	for name, reason := range rt.Degraded {
		degraded[name] = reason
	}
	return degraded
}

func (rt *Runtime) AppConfig() *configs.Config { return rt.Config }

func (rt *Runtime) ArtisanService() artisanContracts.Artisan { return rt.Artisan }

func (rt *Runtime) ConfigRepository() configContracts.Repository { return rt.ConfigRepo }

func (rt *Runtime) DatabaseManager() databaseContracts.Manager { return rt.Database }

func (rt *Runtime) ConnectionValue() databaseContracts.Connection { return rt.Connection }

func (rt *Runtime) CacheManagerValue() cacheContracts.Manager { return rt.CacheManager }

func (rt *Runtime) CacheStore() cacheContracts.Store { return rt.Cache }

func (rt *Runtime) AuthManager() authContracts.Manager { return rt.Auth }

func (rt *Runtime) AuthorizationService() authorizationContracts.Authorizer { return rt.Authorization }

func (rt *Runtime) MailManagerValue() mailContracts.Manager { return rt.MailManager }

func (rt *Runtime) EmailServiceValue() mailContracts.Mailer { return rt.EmailService }

func (rt *Runtime) QueueManagerValue() queueContracts.Manager { return rt.QueueManager }

func (rt *Runtime) QueueServiceValue() queueContracts.Queue { return rt.QueueService }

func (rt *Runtime) ScheduleManagerValue() scheduleContracts.Manager { return rt.ScheduleManager }

func (rt *Runtime) ScheduleServiceValue() scheduleContracts.Scheduler { return rt.ScheduleService }

func (rt *Runtime) SearchManagerValue() search.Manager { return rt.SearchManager }

func (rt *Runtime) SearchServiceValue() search.Engine { return rt.SearchService }

func (rt *Runtime) StorageValue() storageContracts.StorageManager { return rt.Storage }

func (rt *Runtime) HashService() hashContracts.Hasher { return rt.Hash }

func (rt *Runtime) NotificationService() notificationContracts.Dispatcher { return rt.Notification }

func (rt *Runtime) TranslatorService() i18nContracts.Translator { return rt.Translator }

func (rt *Runtime) LogService() loggingContracts.Logger { return rt.Log }

func (rt *Runtime) RateLimiterService() ratelimiterContracts.Limiter { return rt.RateLimiter }

func (rt *Runtime) LockerValue() lock.Locker { return rt.Locker }

func (rt *Runtime) BackupService() *backup.Service { return rt.Backup }
