package providers

import (
	"fiber-starter/app/controllers"
	"fiber-starter/app/services"
	"fiber-starter/config"
	"fmt"

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
	// 注册所有Provider
	for _, provider := range c.providers {
		if err := provider.Register(); err != nil {
			return fmt.Errorf("failed to register provider: %w", err)
		}
	}

	// 注册服务
	if err := c.RegisterServices(); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}

	// 注册控制器
	if err := c.RegisterControllers(); err != nil {
		return fmt.Errorf("failed to register controllers: %w", err)
	}

	return nil
}

// RegisterServices 注册所有服务
func (c *Container) RegisterServices() error {
	// 认证服务
	if err := c.Provide(services.NewAuthService); err != nil {
		return fmt.Errorf("failed to provide auth service: %w", err)
	}

	// 用户服务
	if err := c.Provide(services.NewUserService); err != nil {
		return fmt.Errorf("failed to provide user service: %w", err)
	}

	// 邮件服务
	if err := c.Provide(services.NewEmailService); err != nil {
		return fmt.Errorf("failed to provide email service: %w", err)
	}

	// 队列服务
	if err := c.Provide(services.NewQueueService); err != nil {
		return fmt.Errorf("failed to provide queue service: %w", err)
	}

	// 存储服务 - 需要StorageConfig和RedisConfig两个参数
	if err := c.Provide(func(cfg *config.Config, redisCfg *config.RedisConfig) (*services.StorageService, error) {
		return services.NewStorageService(&cfg.Storage, redisCfg)
	}); err != nil {
		return fmt.Errorf("failed to provide storage service: %w", err)
	}

	return nil
}

// RegisterControllers 注册所有控制器
func (c *Container) RegisterControllers() error {
	// 认证控制器
	if err := c.Provide(controllers.NewAuthController); err != nil {
		return fmt.Errorf("failed to provide auth controller: %w", err)
	}

	// 用户控制器
	if err := c.Provide(controllers.NewUserController); err != nil {
		return fmt.Errorf("failed to provide user controller: %w", err)
	}

	// 存储控制器
	if err := c.Provide(controllers.NewStorageController); err != nil {
		return fmt.Errorf("failed to provide storage controller: %w", err)
	}

	return nil
}

// Invoke 调用依赖注入的函数
func (c *Container) Invoke(function any) error {
	return c.Container.Invoke(function)
}
