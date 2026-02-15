// Package providers 管理应用程序的服务提供者
package providers

import (
	"fiber-starter/config"
	"github.com/go-playground/validator/v10"
	"go.uber.org/dig"
)

// AppProvider 应用服务提供者
type AppProvider struct {
	container *dig.Container
}

// NewAppProvider 创建应用服务提供者
func NewAppProvider(container *dig.Container) *AppProvider {
	return &AppProvider{
		container: container,
	}
}

// Register 注册应用相关的依赖
func (p *AppProvider) Register() error {
	// 注册配置
	if err := p.container.Provide(config.LoadConfig); err != nil {
		return err
	}

	// 注册验证器
	if err := p.container.Provide(validator.New); err != nil {
		return err
	}

	return nil
}
