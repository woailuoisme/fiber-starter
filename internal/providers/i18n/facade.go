package i18n

import (
	"errors"

	"fiber-starter/internal/providers/i18n/Contracts"
	"fiber-starter/internal/support/appctx"

	"github.com/gofiber/fiber/v3"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// translator returns the translator instance from the container.
func translator() Contracts.Translator {
	if app := appctx.App(); app != nil {
		return app.TranslatorService()
	}
	return nil
}

// Trans translates the given message key.
func Trans(c fiber.Ctx, key string, params map[string]interface{}, locale ...string) string {
	if t := translator(); t != nil {
		return t.Trans(c, key, params, locale...)
	}
	return key
}

// Choice translates the given message key with pluralization (if supported).
func Choice(c fiber.Ctx, key string, number int, params map[string]interface{}, locale ...string) string {
	if t := translator(); t != nil {
		return t.Choice(c, key, number, params, locale...)
	}
	return key
}

// GetLocale returns the current locale.
func GetLocale(c fiber.Ctx) string {
	if t := translator(); t != nil {
		return t.GetLocale(c)
	}
	return ""
}

// SetLocale sets the current locale and persists it (e.g., via cookie).
func SetLocale(c fiber.Ctx, locale string) error {
	if t := translator(); t != nil {
		return t.SetLocale(c, locale)
	}
	return ErrContainerNotInitialized
}

// Middleware returns the language handling middleware.
func Middleware() fiber.Handler {
	if t := translator(); t != nil {
		return t.Middleware()
	}
	// Fallback to empty middleware if container not initialized
	return func(c fiber.Ctx) error { return c.Next() }
}
