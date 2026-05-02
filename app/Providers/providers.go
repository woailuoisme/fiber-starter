// Package providers groups explicit application wiring helpers.
package providers

import (
	"errors"
	"fmt"

	controllers "fiber-starter/app/Http/Controllers"
	httpservices "fiber-starter/app/Http/Services"
	services "fiber-starter/app/Services"
	helpers "fiber-starter/app/Support"
	supporti18n "fiber-starter/app/Support/i18n"
	"fiber-starter/config"
	"fiber-starter/database"
)

// Runtime holds the application dependencies that are wired at startup.
type Runtime struct {
	Config           *config.Config
	Connection       *database.Connection
	Cache            helpers.CacheService
	I18n             *supporti18n.Service
	AuthService      httpservices.AuthService
	UserService      httpservices.UserService
	EmailService     services.EmailService
	QueueService     services.QueueService
	SearchService    services.SearchService
	StorageService   *services.StorageService
	AuthController   *controllers.AuthController
	UserController   *controllers.UserController
	HealthController *controllers.HealthController
}

// Build wires application dependencies explicitly without a container.
func Build(cfg *config.Config) (*Runtime, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	connection, err := database.NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database connection: %w", err)
	}

	cache := helpers.NewCacheService(cfg)

	i18nService, err := supporti18n.Init(&cfg.I18n)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize i18n service: %w", err)
	}

	authService := httpservices.NewAuthService(connection, cfg, cache)
	userService := httpservices.NewUserService(connection)
	emailService := services.NewEmailService(cfg)
	queueService := services.NewQueueService(cfg)
	searchService := services.NewSearchService(cfg)
	storageService, err := services.NewStorageService(&cfg.Storage, &cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage service: %w", err)
	}

	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	healthController := controllers.NewHealthController(cfg, connection, cache)

	return &Runtime{
		Config:           cfg,
		Connection:       connection,
		Cache:            cache,
		I18n:             i18nService,
		AuthService:      authService,
		UserService:      userService,
		EmailService:     emailService,
		QueueService:     queueService,
		SearchService:    searchService,
		StorageService:   storageService,
		AuthController:   authController,
		UserController:   userController,
		HealthController: healthController,
	}, nil
}

// Close shuts down runtime resources in reverse initialization order.
func (r *Runtime) Close() error {
	if r == nil {
		return nil
	}

	var errs []error
	if r.StorageService != nil {
		if err := r.StorageService.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.QueueService != nil {
		if err := r.QueueService.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.Cache != nil {
		if err := r.Cache.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.Connection != nil {
		if err := r.Connection.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
