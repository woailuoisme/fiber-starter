package i18n

import (
	"fiber-starter/configs"
	i18nContracts "fiber-starter/internal/providers/i18n/contracts"

	"github.com/gofiber/fiber/v3"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

// TranslatorWrapper wraps the Support/i18n.Service to implement the Translator contract.
type TranslatorWrapper struct {
	service *Service
}

// NewTranslatorWrapper creates a new TranslatorWrapper.
func NewTranslatorWrapper(service *Service) *TranslatorWrapper {
	return &TranslatorWrapper{service: service}
}

// Trans translates the given message key.
func (w *TranslatorWrapper) Trans(c fiber.Ctx, key string, params map[string]interface{}, locale ...string) string {
	return w.service.MustLocalize(c, &goi18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: params,
		DefaultMessage: &goi18n.Message{
			ID: key,
		},
	})
}

// Note: Locale override via LocalizeConfig might need specific handling if the middleware doesn't support it directly in params.
// But usually, it uses the language from the context.

// Choice translates the given message key with pluralization.
func (w *TranslatorWrapper) Choice(c fiber.Ctx, key string, number int, params map[string]interface{}, locale ...string) string {
	if params == nil {
		params = make(map[string]interface{})
	}
	params["Count"] = number
	return w.Trans(c, key, params, locale...)
}

// GetLocale returns the current locale.
func (w *TranslatorWrapper) GetLocale(c fiber.Ctx) string {
	return GetCurrentLanguage(c, *w.service.cfg)
}

// SetLocale sets the current locale and persists it (e.g., via cookie).
func (w *TranslatorWrapper) SetLocale(c fiber.Ctx, locale string) error {
	return SetLanguage(c, *w.service.cfg, locale)
}

// Middleware returns the language handling middleware.
func (w *TranslatorWrapper) Middleware() fiber.Handler {
	return w.service.Middleware()
}

// RegisterI18n initializes the I18n service and returns the manager and translator.
func RegisterI18n(cfg *configs.Config) (*Service, i18nContracts.Translator, error) {
	service, err := Init(&cfg.I18n)
	if err != nil {
		return nil, nil, err
	}
	return service, NewTranslatorWrapper(service), nil
}
