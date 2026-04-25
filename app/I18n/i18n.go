// Package i18n Handles internationalization and localization logic
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

var Bundle *i18n.Bundle
var SupportedLanguages []string
var DefaultLanguage string

func Init() error {
	cfg := config.GlobalConfig.I18n
	DefaultLanguage = cfg.DefaultLanguage
	SupportedLanguages = cfg.SupportedLanguages

	defaultLang, err := language.Parse(DefaultLanguage)
	if err != nil {
		helpers.Error("Failed to parse default language", zap.String("language", DefaultLanguage), zap.Error(err))
		defaultLang = language.Chinese
	}

	Bundle = i18n.NewBundle(defaultLang)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	if err := LoadLanguageFiles(); err != nil {
		helpers.Error("Failed to load language files", zap.Error(err))
		return err
	}

	helpers.Info("i18n system initialized successfully",
		zap.String("default_language", DefaultLanguage),
		zap.Strings("supported_languages", SupportedLanguages))

	return nil
}

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

func GetLocalizer(lang string) *i18n.Localizer {
	if Bundle == nil {
		helpers.Error("i18n Bundle not initialized")
		return nil
	}

	if !IsSupported(lang) {
		helpers.Warn("Requested language not supported, using default language",
			zap.String("requested", lang),
			zap.String("default", DefaultLanguage))
		lang = DefaultLanguage
	}

	return i18n.NewLocalizer(Bundle, lang)
}

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

func Reload() error {
	helpers.Info("Reloading language files")
	return LoadLanguageFiles()
}

func GetSupportedLanguages() []string {
	return SupportedLanguages
}

func GetDefaultLanguage() string {
	return DefaultLanguage
}
