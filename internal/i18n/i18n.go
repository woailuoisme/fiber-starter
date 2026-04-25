// Package i18n Handles internationalization and localization logic
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

// Bundle Global translation bundle
var Bundle *i18n.Bundle

// SupportedLanguages List of supported languages
var SupportedLanguages []string

// DefaultLanguage Default language
var DefaultLanguage string

// Init Initialize i18n system
func Init() error {
	cfg := config.GlobalConfig.I18n
	DefaultLanguage = cfg.DefaultLanguage
	SupportedLanguages = cfg.SupportedLanguages

	// Parse default language tag
	defaultLang, err := language.Parse(DefaultLanguage)
	if err != nil {
		helpers.Error("Failed to parse default language", zap.String("language", DefaultLanguage), zap.Error(err))
		defaultLang = language.Chinese
	}

	// Create Bundle
	Bundle = i18n.NewBundle(defaultLang)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load language files
	if err := LoadLanguageFiles(); err != nil {
		helpers.Error("Failed to load language files", zap.Error(err))
		return err
	}

	helpers.Info("i18n system initialized successfully",
		zap.String("default_language", DefaultLanguage),
		zap.Strings("supported_languages", SupportedLanguages))

	return nil
}

// LoadLanguageFiles Load language files
func LoadLanguageFiles() error {
	languageDir := config.GlobalConfig.I18n.LanguageDir

	if err := ensureLanguageDir(languageDir); err != nil {
		return err
	}

	for _, lang := range SupportedLanguages {
		if err := loadLanguageFile(languageDir, lang); err != nil {
			helpers.Warn("Skipping language file",
				zap.String("language", lang),
				zap.Error(err))
		}
	}

	return nil
}

// GetLocalizer Get Localizer for specified language
func GetLocalizer(lang string) *i18n.Localizer {
	if Bundle == nil {
		helpers.Error("i18n Bundle not initialized")
		return nil
	}

	// If language is not supported, use default language
	if !IsSupported(lang) {
		helpers.Warn("Requested language not supported, using default language",
			zap.String("requested", lang),
			zap.String("default", DefaultLanguage))
		lang = DefaultLanguage
	}

	return i18n.NewLocalizer(Bundle, lang)
}

// IsSupported Check if language is supported
func IsSupported(lang string) bool {
	for _, supported := range SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

func ensureLanguageDir(languageDir string) error {
	if _, err := os.Stat(languageDir); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	helpers.Warn("Language directory does not exist, will create directory", zap.String("dir", languageDir))
	return os.MkdirAll(languageDir, 0o750)
}

func loadLanguageFile(languageDir, lang string) error {
	filename := filepath.Join(languageDir, fmt.Sprintf("%s.json", lang))
	if _, err := os.Stat(filename); err != nil {
		return err
	}

	if _, err := Bundle.LoadMessageFile(filename); err != nil {
		return err
	}

	helpers.Info("Language file loaded successfully",
		zap.String("language", lang),
		zap.String("file", filename))
	return nil
}

// Reload Reload language files (for hot reload)
func Reload() error {
	helpers.Info("Reloading language files")
	return LoadLanguageFiles()
}

// GetSupportedLanguages Get list of supported languages
func GetSupportedLanguages() []string {
	return SupportedLanguages
}

// GetDefaultLanguage Get default language
func GetDefaultLanguage() string {
	return DefaultLanguage
}
