package Contracts

import fiber "github.com/gofiber/fiber/v3"

// Translator defines the interface for internationalization and localization.
type Translator interface {
	// Trans translates the given message key.
	Trans(c fiber.Ctx, key string, params map[string]interface{}, locale ...string) string

	// Choice translates the given message key with pluralization (if supported).
	Choice(c fiber.Ctx, key string, number int, params map[string]interface{}, locale ...string) string

	// GetLocale returns the current locale.
	GetLocale(c fiber.Ctx) string

	// SetLocale sets the current locale and persists it (e.g., via cookie).
	SetLocale(c fiber.Ctx, locale string) error

	// Middleware returns the language handling middleware.
	Middleware() fiber.Handler
}
