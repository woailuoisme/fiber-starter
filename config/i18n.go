package config

type I18nConfig struct {
	DefaultLanguage    string   `mapstructure:"default_language"`
	SupportedLanguages []string `mapstructure:"supported_languages"`
	LanguageDir        string   `mapstructure:"language_dir"`
	CookieName         string   `mapstructure:"cookie_name"`
	CookieMaxAge       int      `mapstructure:"cookie_max_age"`
}
