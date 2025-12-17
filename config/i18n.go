package config

// I18nConfig i18n 国际化配置
type I18nConfig struct {
	DefaultLanguage    string   `mapstructure:"default_language"`    // 默认语言
	SupportedLanguages []string `mapstructure:"supported_languages"` // 支持的语言列表
	LanguageDir        string   `mapstructure:"language_dir"`        // 语言文件目录
	CookieName         string   `mapstructure:"cookie_name"`         // 语言 Cookie 名称
	CookieMaxAge       int      `mapstructure:"cookie_max_age"`      // Cookie 过期时间（秒）
}
