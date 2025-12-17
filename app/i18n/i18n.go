package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"

	"fiber-starter/app/helpers"
	"fiber-starter/config"
)

// Bundle 全局翻译包
var Bundle *i18n.Bundle

// SupportedLanguages 支持的语言列表
var SupportedLanguages []string

// DefaultLanguage 默认语言
var DefaultLanguage string

// Init 初始化 i18n 系统
func Init() error {
	// 获取配置
	cfg := config.GlobalConfig.I18n

	// 设置默认语言和支持的语言列表
	DefaultLanguage = cfg.DefaultLanguage
	SupportedLanguages = cfg.SupportedLanguages

	// 解析默认语言标签
	defaultLang, err := language.Parse(DefaultLanguage)
	if err != nil {
		helpers.Error("解析默认语言失败", zap.String("language", DefaultLanguage), zap.Error(err))
		defaultLang = language.Chinese
	}

	// 创建 Bundle
	Bundle = i18n.NewBundle(defaultLang)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 加载语言文件
	if err := LoadLanguageFiles(); err != nil {
		helpers.Error("加载语言文件失败", zap.Error(err))
		return err
	}

	helpers.Info("i18n 系统初始化成功",
		zap.String("default_language", DefaultLanguage),
		zap.Strings("supported_languages", SupportedLanguages))

	return nil
}

// LoadLanguageFiles 加载语言文件
func LoadLanguageFiles() error {
	cfg := config.GlobalConfig.I18n
	languageDir := cfg.LanguageDir

	// 检查语言目录是否存在
	if _, err := os.Stat(languageDir); os.IsNotExist(err) {
		helpers.Warn("语言目录不存在，将创建目录", zap.String("dir", languageDir))
		if err := os.MkdirAll(languageDir, 0755); err != nil {
			return err
		}
	}

	// 加载每个支持的语言文件
	for _, lang := range SupportedLanguages {
		filename := filepath.Join(languageDir, lang+".json")

		// 检查文件是否存在
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			helpers.Warn("语言文件不存在",
				zap.String("language", lang),
				zap.String("file", filename))
			continue
		}

		// 加载语言文件
		if _, err := Bundle.LoadMessageFile(filename); err != nil {
			helpers.Error("加载语言文件失败",
				zap.String("language", lang),
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		helpers.Info("语言文件加载成功",
			zap.String("language", lang),
			zap.String("file", filename))
	}

	return nil
}

// GetLocalizer 获取指定语言的 Localizer
func GetLocalizer(lang string) *i18n.Localizer {
	if Bundle == nil {
		helpers.Error("i18n Bundle 未初始化")
		return nil
	}

	// 如果语言不支持，使用默认语言
	if !IsSupported(lang) {
		helpers.Warn("请求的语言不支持，使用默认语言",
			zap.String("requested", lang),
			zap.String("default", DefaultLanguage))
		lang = DefaultLanguage
	}

	return i18n.NewLocalizer(Bundle, lang)
}

// IsSupported 检查语言是否支持
func IsSupported(lang string) bool {
	for _, supported := range SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// Reload 重新加载语言文件（用于热重载）
func Reload() error {
	helpers.Info("重新加载语言文件")
	return LoadLanguageFiles()
}

// GetSupportedLanguages 获取支持的语言列表
func GetSupportedLanguages() []string {
	return SupportedLanguages
}

// GetDefaultLanguage 获取默认语言
func GetDefaultLanguage() string {
	return DefaultLanguage
}
