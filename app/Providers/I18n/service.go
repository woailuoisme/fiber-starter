package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"fiber-starter/configs"

	contribi18n "github.com/gofiber/contrib/v3/i18n"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/text/language"
)

// Service wraps the official Fiber i18n container.
type Service struct {
	translator *contribi18n.I18n
	cfg        *configs.I18nConfig
}

// Init initializes the official Fiber i18n container.
func Init(cfg *configs.I18nConfig) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("i18n config is nil")
	}

	rootPath := cfg.LanguageDir
	if rootPath == "" {
		rootPath = "./lang"
	}

	// Resolve relative paths against the current working directory first, then
	// walk upward so tests executed from package directories can still find the
	// repo-level lang directory.
	if !filepath.IsAbs(rootPath) {
		if resolved := resolveUpwardPath(rootPath, 4); resolved != "" {
			rootPath = resolved
		} else if abs, err := filepath.Abs(rootPath); err == nil {
			rootPath = abs
		}
	}

	if len(cfg.SupportedLanguages) == 0 {
		cfg.SupportedLanguages = []string{"en", "zh-CN"}
	}

	supported := make([]language.Tag, 0, len(cfg.SupportedLanguages))
	for _, lang := range cfg.SupportedLanguages {
		if tag, err := language.Parse(lang); err == nil {
			supported = append(supported, tag)
		}
	}

	defaultLang, err := language.Parse(cfg.DefaultLanguage)
	if err != nil {
		defaultLang = language.English
	}

	service := &Service{
		cfg: cfg,
	}
	service.translator = contribi18n.New(&contribi18n.Config{
		RootPath:         rootPath,
		AcceptLanguages:  supported,
		DefaultLanguage:  defaultLang,
		FormatBundleFile: "json",
		UnmarshalFunc:    legacyJSONUnmarshal,
		Loader:           contribi18n.LoaderFunc(os.ReadFile),
		LangHandler: func(c fiber.Ctx, defaultLang string) string {
			return GetCurrentLanguage(c, *cfg)
		},
	})

	return service, nil
}

// Localize resolves a message.
func (s *Service) Localize(c fiber.Ctx, params interface{}) (string, error) {
	if s == nil || s.translator == nil {
		return fmt.Sprintf("%v", params), nil
	}
	return s.translator.Localize(c, params)
}

// MustLocalize resolves a message and panics on error.
func (s *Service) MustLocalize(c fiber.Ctx, params interface{}) string {
	if s == nil || s.translator == nil {
		return fmt.Sprintf("%v", params)
	}
	return s.translator.MustLocalize(c, params)
}

// Middleware handles language persistence (cookies).
func (s *Service) Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if s == nil || s.cfg == nil {
			return c.Next()
		}

		// If lang is in query, persist it to cookie
		if lang := c.Query(s.cfg.CookieName); lang != "" {
			c.Cookie(&fiber.Cookie{
				Name:     s.cfg.CookieName,
				Value:    lang,
				MaxAge:   s.cfg.CookieMaxAge,
				HTTPOnly: true,
				Path:     "/",
			})
		}

		return c.Next()
	}
}

// legacyJSONUnmarshal maintains compatibility with the existing JSON structure.
func legacyJSONUnmarshal(data []byte, v interface{}) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	normalized := normalizeCatalog(raw)
	switch dst := v.(type) {
	case *interface{}:
		*dst = normalized
		return nil
	default:
		return json.Unmarshal(data, v)
	}
}

func normalizeCatalog(value interface{}) interface{} {
	switch data := value.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(data))
		for key, item := range data {
			normalized[key] = normalizeCatalog(item)
		}
		return normalized
	case string:
		return map[string]interface{}{
			"other": data,
		}
	default:
		return value
	}
}

func resolveUpwardPath(path string, depth int) string {
	candidate := path
	for i := 0; i <= depth; i++ {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		candidate = filepath.Join("..", candidate)
	}

	return ""
}
