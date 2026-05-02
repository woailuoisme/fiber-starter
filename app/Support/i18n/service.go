package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fiber-starter/config"

	contribi18n "github.com/gofiber/contrib/v3/i18n"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/text/language"
)

// Service wraps the Fiber official i18n container and project-specific helpers.
type Service struct {
	translator *contribi18n.I18n
	cfg        config.I18nConfig
}

var (
	defaultMu      sync.RWMutex
	defaultService *Service
)

// Init builds the shared i18n service and stores it as the default singleton.
func Init(cfg *config.I18nConfig) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("i18n config is nil")
	}

	resolvedRoot, err := resolveRootPath(cfg.LanguageDir)
	if err != nil {
		return nil, err
	}

	serviceCfg := cloneConfig(cfg)
	defaultLanguage := resolveDefaultLanguage(serviceCfg.DefaultLanguage, serviceCfg.SupportedLanguages)
	acceptLanguages := parseSupportedLanguages(serviceCfg.SupportedLanguages)

	service := &Service{
		cfg: serviceCfg,
	}

	service.translator = contribi18n.New(&contribi18n.Config{
		RootPath:         resolvedRoot,
		AcceptLanguages:  acceptLanguages,
		FormatBundleFile: "json",
		DefaultLanguage:  defaultLanguage,
		Loader:           contribi18n.LoaderFunc(os.ReadFile),
		UnmarshalFunc:    legacyJSONUnmarshal,
		LangHandler: func(ctx fiber.Ctx, fallback string) string {
			return service.resolveLanguage(ctx, fallback)
		},
	})

	defaultMu.Lock()
	defaultService = service
	defaultMu.Unlock()

	return service, nil
}

// Default returns the initialized singleton service.
func Default() *Service {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultService
}

// Localize resolves a message with the shared service.
func Localize(c fiber.Ctx, params interface{}) (string, error) {
	svc := Default()
	if svc == nil {
		return "", errors.New("i18n service is not initialized")
	}
	return svc.Localize(c, params)
}

// MustLocalize resolves a message and panics on error.
func MustLocalize(c fiber.Ctx, params interface{}) string {
	svc := Default()
	if svc == nil {
		panic("i18n service is not initialized")
	}
	return svc.MustLocalize(c, params)
}

// Localize resolves a message using the shared i18n container.
func (s *Service) Localize(c fiber.Ctx, params interface{}) (string, error) {
	if s == nil || s.translator == nil {
		return "", errors.New("i18n service is not initialized")
	}
	return s.translator.Localize(c, params)
}

// MustLocalize resolves a message and panics if translation fails.
func (s *Service) MustLocalize(c fiber.Ctx, params interface{}) string {
	if s == nil || s.translator == nil {
		panic("i18n service is not initialized")
	}
	return s.translator.MustLocalize(c, params)
}

func (s *Service) resolveLanguage(c fiber.Ctx, fallback string) string {
	cfg := s.cfg
	if cfg.CookieName == "" {
		cfg.CookieName = "lang"
	}

	if lang := resolveQueryLanguage(c, cfg); lang != "" {
		setLanguageCookie(c, cfg, lang)
		return lang
	}

	if lang := resolveCookieLanguage(c, cfg); lang != "" {
		return lang
	}

	if lang := resolveAcceptLanguage(c, cfg); lang != "" {
		return lang
	}

	if lang := matchSupportedLanguage(fallback, cfg.SupportedLanguages); lang != "" {
		return lang
	}

	if lang := normalizeLanguageCode(cfg.DefaultLanguage); lang != "" {
		if supported := matchSupportedLanguage(lang, cfg.SupportedLanguages); supported != "" {
			return supported
		}
		return lang
	}

	return "en"
}

func resolveRootPath(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = "./lang"
	}

	candidates := []string{root}
	if !filepath.IsAbs(root) {
		candidates = append(candidates, filepath.Join("..", root))
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return filepath.Abs(candidate)
		}
	}

	return filepath.Abs(root)
}

func cloneConfig(cfg *config.I18nConfig) config.I18nConfig {
	cloned := *cfg
	if cfg.SupportedLanguages != nil {
		cloned.SupportedLanguages = append([]string(nil), cfg.SupportedLanguages...)
	}
	return cloned
}

func resolveDefaultLanguage(defaultLanguage string, supported []string) language.Tag {
	if lang := matchSupportedLanguage(defaultLanguage, supported); lang != "" {
		if tag, err := language.Parse(lang); err == nil {
			return tag
		}
	}

	if len(supported) > 0 {
		if tag, err := language.Parse(normalizeLanguageCode(supported[0])); err == nil {
			return tag
		}
	}

	return language.English
}

func parseSupportedLanguages(languages []string) []language.Tag {
	if len(languages) == 0 {
		return []language.Tag{language.English}
	}

	tags := make([]language.Tag, 0, len(languages))
	for _, item := range languages {
		lang := normalizeLanguageCode(item)
		if lang == "" {
			continue
		}
		tag, err := language.Parse(lang)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		return []language.Tag{language.English}
	}

	return tags
}

func legacyJSONUnmarshal(data []byte, v interface{}) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	normalized := normalizeLegacyCatalog(raw)
	switch dst := v.(type) {
	case *interface{}:
		*dst = normalized
		return nil
	default:
		return json.Unmarshal(data, v)
	}
}

func normalizeLegacyCatalog(value interface{}) interface{} {
	switch data := value.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(data))
		for key, item := range data {
			switch strings.ToLower(key) {
			case "description", "translation":
				if _, ok := item.(string); ok {
					normalized[key] = map[string]interface{}{
						"other": item,
					}
					continue
				}
			}
			normalized[key] = normalizeLegacyCatalog(item)
		}
		return normalized
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{}, len(data))
		for key, item := range data {
			strKey, ok := key.(string)
			if !ok {
				strKey = fmt.Sprintf("%v", key)
			}
			switch strings.ToLower(strKey) {
			case "description", "translation":
				if _, ok := item.(string); ok {
					normalized[strKey] = map[string]interface{}{
						"other": item,
					}
					continue
				}
			}
			normalized[strKey] = normalizeLegacyCatalog(item)
		}
		return normalized
	case []interface{}:
		normalized := make([]interface{}, 0, len(data))
		for _, item := range data {
			normalized = append(normalized, normalizeLegacyCatalog(item))
		}
		return normalized
	default:
		return value
	}
}
