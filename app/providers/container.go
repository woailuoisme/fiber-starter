package providers

import (
	"fiber-starter/app/controllers"
	"fiber-starter/app/services"
	"fiber-starter/config"
	"fiber-starter/database"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	*dig.Container
}

// NewContainer 创建新的依赖注入容器
func NewContainer() *Container {
	return &Container{
		Container: dig.New(),
	}
}

// RegisterProviders 注册所有服务提供者
func (c *Container) RegisterProviders() error {
	// 注册配置
	if err := c.Provide(config.LoadConfig); err != nil {
		return fmt.Errorf("failed to provide config: %w", err)
	}

	// 注册Redis配置
	if err := c.Provide(func(cfg *config.Config) *config.RedisConfig {
		return &cfg.Redis
	}); err != nil {
		return fmt.Errorf("failed to provide redis config: %w", err)
	}

	// 注册验证器
	if err := c.Provide(validator.New); err != nil {
		return fmt.Errorf("failed to provide validator: %w", err)
	}

	// 注册数据库连接
	if err := c.Provide(database.NewConnection); err != nil {
		return fmt.Errorf("failed to provide database: %w", err)
	}

	// 注册GORM实例
	if err := c.Provide(func(conn *database.Connection) *gorm.DB {
		return conn.DB
	}); err != nil {
		return fmt.Errorf("failed to provide gorm db: %w", err)
	}

	// 注册服务
	if err := c.RegisterServices(); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}

	// 注册控制器
	if err := c.Provide(controllers.NewAuthController); err != nil {
		return fmt.Errorf("failed to provide auth controller: %w", err)
	}
	if err := c.Provide(controllers.NewUserController); err != nil {
		return fmt.Errorf("failed to provide user controller: %w", err)
	}
	if err := c.Provide(controllers.NewStorageController); err != nil {
		return fmt.Errorf("failed to provide storage controller: %w", err)
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

	// 缓存服务
	if err := c.Provide(services.NewCacheService); err != nil {
		return fmt.Errorf("failed to provide cache service: %w", err)
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

// Invoke 调用依赖注入的函数
func (c *Container) Invoke(function interface{}) error {
	return c.Container.Invoke(function)
}
