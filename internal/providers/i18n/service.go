package i18n

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"lfiber/configs"

	"github.com/BurntSushi/toml"
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
	translator, err := newTranslator(rootPath, supported, defaultLang, cfg)
	if err != nil {
		return nil, err
	}
	service.translator = translator

	return service, nil
}

func newTranslator(rootPath string, supported []language.Tag, defaultLang language.Tag, cfg *configs.I18nConfig) (translator *contribi18n.I18n, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("initialize i18n translator: %v", recovered)
		}
	}()

	translator = contribi18n.New(&contribi18n.Config{
		RootPath:         rootPath,
		AcceptLanguages:  supported,
		DefaultLanguage:  defaultLang,
		FormatBundleFile: "toml",
		UnmarshalFunc:    toml.Unmarshal,
		Loader:           contribi18n.LoaderFunc(loadLanguageBundle),
		LangHandler: func(c fiber.Ctx, defaultLang string) string {
			return GetCurrentLanguage(c, *cfg)
		},
	})

	return translator, nil
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

var messageIDPattern = regexp.MustCompile(`^\s*\["([^"]+)"\]\s*$`)

func loadLanguageBundle(path string) ([]byte, error) {
	localeDir := languageDirectoryForBundle(path)
	entries, err := os.ReadDir(localeDir)
	if err != nil {
		return nil, fmt.Errorf("read language directory %s: %w", localeDir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toml") {
			files = append(files, filepath.Join(localeDir, entry.Name()))
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("language directory %s contains no toml files", localeDir)
	}
	sort.Strings(files)

	var bundle bytes.Buffer
	seen := make(map[string]string)
	for _, file := range files {
		data, err := os.ReadFile(file) //nolint:gosec // file is discovered from the configured local language directory
		if err != nil {
			return nil, fmt.Errorf("read language file %s: %w", file, err)
		}
		if err := ensureUniqueMessageIDs(file, data, seen); err != nil {
			return nil, err
		}
		bundle.Write(data)
		if len(data) == 0 || data[len(data)-1] != '\n' {
			bundle.WriteByte('\n')
		}
		bundle.WriteByte('\n')
	}

	return bundle.Bytes(), nil
}

func languageDirectoryForBundle(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	locale := strings.TrimSuffix(base, ext)
	return filepath.Join(filepath.Dir(path), locale)
}

func ensureUniqueMessageIDs(file string, data []byte, seen map[string]string) error {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		matches := messageIDPattern.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}
		messageID := matches[1]
		if previous, ok := seen[messageID]; ok {
			return fmt.Errorf("duplicate i18n message %q in %s and %s", messageID, previous, file)
		}
		seen[messageID] = file
	}
	return nil
}
