package i18n

import (
	"strings"
	"time"

	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/text/language"
)

func resolveQueryLanguage(c fiber.Ctx, cfg config.I18nConfig) string {
	if c == nil {
		return ""
	}

	if queryLang := normalizeLanguageCode(c.Query("lang")); queryLang != "" {
		if lang := matchSupportedLanguage(queryLang, cfg.SupportedLanguages); lang != "" {
			return lang
		}
	}

	return ""
}

func resolveCookieLanguage(c fiber.Ctx, cfg config.I18nConfig) string {
	if c == nil {
		return ""
	}

	if cookieLang := normalizeLanguageCode(c.Cookies(cookieName(cfg))); cookieLang != "" {
		return matchSupportedLanguage(cookieLang, cfg.SupportedLanguages)
	}

	return ""
}

func resolveAcceptLanguage(c fiber.Ctx, cfg config.I18nConfig) string {
	if c == nil {
		return ""
	}

	header := strings.TrimSpace(c.Get("Accept-Language"))
	if header == "" {
		return ""
	}

	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return ""
	}

	for _, tag := range tags {
		if lang := matchSupportedLanguage(tag.String(), cfg.SupportedLanguages); lang != "" {
			return lang
		}
	}

	return ""
}

func matchSupportedLanguage(lang string, supported []string) string {
	normalized := normalizeLanguageCode(lang)
	if normalized == "" {
		return ""
	}

	for _, candidate := range supported {
		if normalizeLanguageCode(candidate) == normalized {
			return normalizeLanguageCode(candidate)
		}
	}

	if len(supported) == 0 {
		return ""
	}

	tag, err := language.Parse(normalized)
	if err != nil {
		return ""
	}
	base, _ := tag.Base()

	for _, candidate := range supported {
		candidateTag, err := language.Parse(normalizeLanguageCode(candidate))
		if err != nil {
			continue
		}
		candidateBase, _ := candidateTag.Base()
		if candidateBase == base {
			return normalizeLanguageCode(candidate)
		}
	}

	return ""
}

func normalizeLanguageCode(code string) string {
	code = strings.TrimSpace(strings.ReplaceAll(code, "_", "-"))
	if code == "" {
		return ""
	}

	tag, err := language.Parse(code)
	if err != nil {
		return strings.ToLower(code)
	}

	return tag.String()
}

func setLanguageCookie(c fiber.Ctx, cfg config.I18nConfig, lang string) {
	if c == nil || lang == "" {
		return
	}

	maxAge := cfg.CookieMaxAge
	if maxAge <= 0 {
		maxAge = 24 * 60 * 60
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookieName(cfg),
		Value:    lang,
		MaxAge:   maxAge,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
	})
}

func cookieName(cfg config.I18nConfig) string {
	name := strings.TrimSpace(cfg.CookieName)
	if name == "" {
		return "lang"
	}
	return name
}

// GetLanguageCookie returns the persisted language cookie value.
func GetLanguageCookie(c fiber.Ctx) string {
	cfg := currentConfig()
	if c == nil {
		return ""
	}
	return c.Cookies(cookieName(cfg))
}

// ClearLanguageCookie removes the persisted language cookie.
func ClearLanguageCookie(c fiber.Ctx) {
	if c == nil {
		return
	}

	cfg := currentConfig()
	c.Cookie(&fiber.Cookie{
		Name:     cookieName(cfg),
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		Expires:  time.Unix(0, 0),
	})
}

// SetLanguage persists the given language in the cookie jar.
func SetLanguage(c fiber.Ctx, lang string) error {
	cfg := currentConfig()
	matched := matchSupportedLanguage(lang, cfg.SupportedLanguages)
	if matched == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Unsupported language: "+lang)
	}

	setLanguageCookie(c, cfg, matched)
	return nil
}

// GetCurrentLanguage returns the active language for the request.
func GetCurrentLanguage(c fiber.Ctx) string {
	cfg := currentConfig()
	if current := resolveQueryLanguage(c, cfg); current != "" {
		return current
	}
	if current := resolveCookieLanguage(c, cfg); current != "" {
		return current
	}
	if current := resolveAcceptLanguage(c, cfg); current != "" {
		return current
	}
	if current := matchSupportedLanguage(cfg.DefaultLanguage, cfg.SupportedLanguages); current != "" {
		return current
	}
	return normalizeLanguageCode(cfg.DefaultLanguage)
}

func currentConfig() config.I18nConfig {
	if svc := Default(); svc != nil {
		return svc.cfg
	}

	if config.GlobalConfig != nil {
		return config.GlobalConfig.I18n
	}

	return config.I18nConfig{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "zh-CN"},
		LanguageDir:        "./lang",
		CookieName:         "lang",
		CookieMaxAge:       24 * 60 * 60,
	}
}
