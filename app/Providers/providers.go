package providers

import (
	"fmt"

	controllers "fiber-starter/app/Http/Controllers"
	httpservices "fiber-starter/app/Http/Services"
	services "fiber-starter/app/Services"
	helpers "fiber-starter/app/Support"
	supporti18n "fiber-starter/app/Support/i18n"
	"fiber-starter/config"
	"fiber-starter/database"

	"go.uber.org/dig"
)

type Container struct {
	*dig.Container
	providers []Provider
}

type Provider interface {
	Register() error
}

func NewContainer() *Container {
	container := dig.New()
	return &Container{
		Container: container,
		providers: []Provider{
			NewAppProvider(container),
			NewDatabaseProvider(container),
			NewCacheProvider(container),
		},
	}
}

func (c *Container) RegisterProviders() error {
	if err := c.registerProviders(c.providers); err != nil {
		return fmt.Errorf("failed to register providers: %w", err)
	}
	if err := c.registerServices(); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}
	if err := c.registerControllers(); err != nil {
		return fmt.Errorf("failed to register controllers: %w", err)
	}
	return nil
}

func (c *Container) Invoke(function any) error {
	return c.Container.Invoke(function)
}

func (c *Container) registerProviders(providers []Provider) error {
	for _, provider := range providers {
		if err := provider.Register(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Container) registerServices() error {
	providers := []any{
		httpservices.NewAuthService,
		httpservices.NewUserService,
		services.NewEmailService,
		services.NewQueueService,
		services.NewSearchService,
		func(cfg *config.Config, redisCfg *config.RedisConfig) (*services.StorageService, error) {
			return services.NewStorageService(&cfg.Storage, redisCfg)
		},
	}
	return c.provideAll(providers, "service")
}

func (c *Container) registerControllers() error {
	return c.provideAll([]any{
		controllers.NewAuthController,
		controllers.NewUserController,
		controllers.NewHealthController,
	}, "controller")
}

func (c *Container) provideAll(providers []any, kind string) error {
	for _, provider := range providers {
		if err := c.Provide(provider); err != nil {
			return fmt.Errorf("failed to provide %s: %w", kind, err)
		}
	}
	return nil
}

type (
	AppProvider      struct{ container *dig.Container }
	DatabaseProvider struct{ container *dig.Container }
	CacheProvider    struct{ container *dig.Container }
)

func NewAppProvider(container *dig.Container) *AppProvider { return &AppProvider{container: container} }
func NewDatabaseProvider(container *dig.Container) *DatabaseProvider {
	return &DatabaseProvider{container: container}
}

func NewCacheProvider(container *dig.Container) *CacheProvider {
	return &CacheProvider{container: container}
}

func (p *AppProvider) Register() error {
	if err := p.container.Provide(config.LoadConfig); err != nil {
		return err
	}
	if err := p.container.Provide(func(cfg *config.Config) (*supporti18n.Service, error) {
		return supporti18n.Init(&cfg.I18n)
	}); err != nil {
		return err
	}
	return nil
}

func (p *DatabaseProvider) Register() error {
	return p.container.Provide(database.NewConnection)
}

func (p *CacheProvider) Register() error {
	if err := p.container.Provide(func(cfg *config.Config) *config.RedisConfig {
		return &cfg.Redis
	}); err != nil {
		return err
	}
	return p.container.Provide(helpers.NewCacheService)
}
