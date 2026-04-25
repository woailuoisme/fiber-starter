package i18n

import (
	"strings"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		lang := detectLanguage(c)
		translator := NewTranslator(lang)
		SetToContext(c, translator)

		if c.Query("lang") != "" {
			setLanguageCookie(c, lang)
		}

		return c.Next()
	}
}

func detectLanguage(c fiber.Ctx) string {
	if queryLang := c.Query("lang"); queryLang != "" {
		if lang, ok := supportedLanguage(queryLang); ok {
			helpers.Debug("Detected language from query parameter", zap.String("lang", lang))
			return lang
		}
		helpers.Warn("Language in query parameter not supported", zap.String("lang", queryLang))
	}

	cfg := config.GlobalConfig.I18n
	if cookieLang := c.Cookies(cfg.CookieName); cookieLang != "" {
		if lang, ok := supportedLanguage(cookieLang); ok {
			helpers.Debug("Detected language from Cookie", zap.String("lang", lang))
			return lang
		}
		helpers.Warn("Language in Cookie not supported", zap.String("lang", cookieLang))
	}

	if acceptLang := c.Get("Accept-Language"); acceptLang != "" {
		if lang, ok := firstSupportedLanguage(parseAcceptLanguage(acceptLang)); ok {
			helpers.Debug("Detected language from Accept-Language", zap.String("lang", lang))
			return lang
		}
		helpers.Debug("No supported language in Accept-Language", zap.String("accept-language", acceptLang))
	}

	helpers.Debug("Using default language", zap.String("lang", DefaultLanguage))
	return DefaultLanguage
}

func supportedLanguage(lang string) (string, bool) {
	if lang == "" || !IsSupported(lang) {
		return "", false
	}
	return lang, true
}

func firstSupportedLanguage(languages []string) (string, bool) {
	for _, lang := range languages {
		if IsSupported(lang) {
			return lang, true
		}
	}
	return "", false
}

func parseAcceptLanguage(header string) []string {
	if header == "" {
		return []string{}
	}

	var languages []string
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		langParts := strings.Split(part, ";")
		if len(langParts) > 0 {
			lang := strings.TrimSpace(langParts[0])
			lang = normalizeLanguageCode(lang)
			if lang != "" {
				languages = append(languages, lang)
			}
		}
	}

	return languages
}

func normalizeLanguageCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}

	code = strings.ToLower(code)

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
		parts := strings.Split(code, "-")
		if len(parts) == 2 {
			return parts[0] + "-" + strings.ToUpper(parts[1])
		}
		return code
	}
}

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

func GetLanguageCookie(c fiber.Ctx) string {
	cfg := config.GlobalConfig.I18n
	return c.Cookies(cfg.CookieName)
}

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

func SetLanguage(c fiber.Ctx, lang string) error {
	if !IsSupported(lang) {
		return fiber.NewError(fiber.StatusBadRequest, "Unsupported language: "+lang)
	}

	setLanguageCookie(c, lang)
	translator := NewTranslator(lang)
	SetToContext(c, translator)
	return nil
}

func GetCurrentLanguage(c fiber.Ctx) string {
	translator := GetFromContext(c)
	if translator != nil {
		return translator.GetLanguage()
	}
	return DefaultLanguage
}
