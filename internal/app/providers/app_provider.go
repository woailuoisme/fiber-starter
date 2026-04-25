// Package providers 管理应用程序的服务提供者
package providers

import (
	"fiber-starter/internal/config"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
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

	// 注册验证器，支持多语言翻译
	if err := p.container.Provide(func() (*validator.Validate, *ut.UniversalTranslator) {
		validate := validator.New()

		// 注册通用翻译器
		enLocale := en.New()
		zhLocale := zh.New()
		uni := ut.New(enLocale, enLocale, zhLocale)

		return validate, uni
	}); err != nil {
		return err
	}

	return nil
}
