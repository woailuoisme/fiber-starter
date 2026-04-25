package providers

import (
	"fmt"

	"fiber-starter/internal/config"
	"fiber-starter/internal/services"
	"fiber-starter/internal/transport/http/controllers"

	"go.uber.org/dig"
)

// Container 依赖注入容器
type Container struct {
	*dig.Container
	providers []Provider
}

// Provider 服务提供者接口
type Provider interface {
	Register() error
}

// NewContainer 创建新的依赖注入容器
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

// RegisterProviders 注册所有服务提供者
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

// Invoke 调用依赖注入的函数
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
		services.NewAuthService,
		services.NewUserService,
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
