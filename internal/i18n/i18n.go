// Package i18n Handles internationalization and localization logic
package i18n

import (
	"encoding/json"
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
	// Get config
	cfg := config.GlobalConfig.I18n

	// Set default language and supported languages list
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
	cfg := config.GlobalConfig.I18n
	languageDir := cfg.LanguageDir

	// Check if language directory exists
	if _, err := os.Stat(languageDir); os.IsNotExist(err) {
		helpers.Warn("Language directory does not exist, will create directory", zap.String("dir", languageDir))
		if err := os.MkdirAll(languageDir, 0750); err != nil {
			return err
		}
	}

	// Load each supported language file
	for _, lang := range SupportedLanguages {
		filename := filepath.Join(languageDir, lang+".json")

		// Check if file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			helpers.Warn("Language file does not exist",
				zap.String("language", lang),
				zap.String("file", filename))
			continue
		}

		// Load language file
		if _, err := Bundle.LoadMessageFile(filename); err != nil {
			helpers.Error("Failed to load language file",
				zap.String("language", lang),
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		helpers.Info("Language file loaded successfully",
			zap.String("language", lang),
			zap.String("file", filename))
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
