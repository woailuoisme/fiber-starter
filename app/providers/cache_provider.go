package providers

import (
	"fiber-starter/app/helpers"
	"fiber-starter/config"
	"go.uber.org/dig"
)

// CacheProvider 缓存服务提供者
type CacheProvider struct {
	container *dig.Container
}

// NewCacheProvider 创建缓存服务提供者
func NewCacheProvider(container *dig.Container) *CacheProvider {
	return &CacheProvider{
		container: container,
	}
}

// Register 注册缓存相关的依赖
func (p *CacheProvider) Register() error {
	// 注册Redis配置
	if err := p.container.Provide(func(cfg *config.Config) *config.RedisConfig {
		return &cfg.Redis
	}); err != nil {
		return err
	}

	// 注册缓存服务
	if err := p.container.Provide(func(cfg *config.Config) helpers.CacheService {
		return helpers.NewCacheService(cfg)
	}); err != nil {
		return err
	}

	return nil
}
