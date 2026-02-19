package i18n

import (
	"strings"
	"time"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Middleware Language detection middleware
// Detect user language by priority: Query > Cookie > Accept-Language Header > Default
func Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Detect user language
		lang := detectLanguage(c)

		// Create translator and set to context
		translator := NewTranslator(lang)
		SetToContext(c, translator)

		// If language comes from query parameter, set Cookie
		if c.Query("lang") != "" {
			setLanguageCookie(c, lang)
		}

		return c.Next()
	}
}

// detectLanguage Detect user language preference
// Priority: Query > Cookie > Accept-Language Header > Default
func detectLanguage(c fiber.Ctx) string {
	// 1. Check query parameter
	if lang := c.Query("lang"); lang != "" {
		if IsSupported(lang) {
			helpers.Debug("Detected language from query parameter", zap.String("lang", lang))
			return lang
		}
		helpers.Warn("Language in query parameter not supported", zap.String("lang", lang))
	}

	// 2. Check Cookie
	cfg := config.GlobalConfig.I18n
	if lang := c.Cookies(cfg.CookieName); lang != "" {
		if IsSupported(lang) {
			helpers.Debug("Detected language from Cookie", zap.String("lang", lang))
			return lang
		}
		helpers.Warn("Language in Cookie not supported", zap.String("lang", lang))
	}

	// 3. Check Accept-Language header
	if acceptLang := c.Get("Accept-Language"); acceptLang != "" {
		langs := parseAcceptLanguage(acceptLang)
		for _, lang := range langs {
			if IsSupported(lang) {
				helpers.Debug("Detected language from Accept-Language", zap.String("lang", lang))
				return lang
			}
		}
		helpers.Debug("No supported language in Accept-Language", zap.String("accept-language", acceptLang))
	}

	// 4. Use default language
	helpers.Debug("Using default language", zap.String("lang", DefaultLanguage))
	return DefaultLanguage
}

// parseAcceptLanguage Parse Accept-Language header
// Returns list of languages sorted by priority
func parseAcceptLanguage(header string) []string {
	if header == "" {
		return []string{}
	}

	var languages []string

	// Split multiple languages
	parts := strings.Split(header, ",")
	for _, part := range parts {
		// Remove spaces
		part = strings.TrimSpace(part)

		// Split language and weight (e.g., zh-CN;q=0.9)
		langParts := strings.Split(part, ";")
		if len(langParts) > 0 {
			lang := strings.TrimSpace(langParts[0])

			// Normalize language code
			lang = normalizeLanguageCode(lang)

			if lang != "" {
				languages = append(languages, lang)
			}
		}
	}

	return languages
}

// normalizeLanguageCode Normalize language code
// Example: zh-cn -> zh-CN, en-us -> en
func normalizeLanguageCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}

	// Convert to lowercase
	code = strings.ToLower(code)

	// Handle common language codes
	switch {
	case strings.HasPrefix(code, "zh-cn") || code == "zh":
		return "zh-CN"
	case strings.HasPrefix(code, "zh-tw") || strings.HasPrefix(code, "zh-hk"):
		return "zh-TW"
	case strings.HasPrefix(code, "en"):
		return "en"
	case strings.HasPrefix(code, "ja"):
		return "ja"
	case strings.HasPrefix(code, "ko"):
		return "ko"
	default:
		// For other languages, keep original but capitalize first letter
		parts := strings.Split(code, "-")
		if len(parts) == 2 {
			return parts[0] + "-" + strings.ToUpper(parts[1])
		}
		return code
	}
}

// setLanguageCookie Set language cookie
func setLanguageCookie(c fiber.Ctx, lang string) {
	cfg := config.GlobalConfig.I18n

	c.Cookie(&fiber.Cookie{
		Name:     cfg.CookieName,
		Value:    lang,
		MaxAge:   cfg.CookieMaxAge,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Expires:  time.Now().Add(time.Duration(cfg.CookieMaxAge) * time.Second),
	})

	helpers.Debug("Setting language cookie", zap.String("lang", lang))
}

// GetLanguageCookie Get language cookie
func GetLanguageCookie(c fiber.Ctx) string {
	cfg := config.GlobalConfig.I18n
	return c.Cookies(cfg.CookieName)
}

// ClearLanguageCookie Clear language cookie
func ClearLanguageCookie(c fiber.Ctx) {
	cfg := config.GlobalConfig.I18n

	c.Cookie(&fiber.Cookie{
		Name:     cfg.CookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		Expires:  time.Now().Add(-time.Hour),
	})

	helpers.Debug("Clearing language cookie")
}

// SetLanguage Set user language (via Cookie)
func SetLanguage(c fiber.Ctx, lang string) error {
	if !IsSupported(lang) {
		return fiber.NewError(fiber.StatusBadRequest, "Unsupported language: "+lang)
	}

	setLanguageCookie(c, lang)

	// Update current request's translator
	translator := NewTranslator(lang)
	SetToContext(c, translator)

	return nil
}

// GetCurrentLanguage Get current request's language
func GetCurrentLanguage(c fiber.Ctx) string {
	translator := GetFromContext(c)
	if translator != nil {
		return translator.GetLanguage()
	}
	return DefaultLanguage
}
