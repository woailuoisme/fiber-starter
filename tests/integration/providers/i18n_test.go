package providers

import (
	"os"
	"path/filepath"
	"testing"

	"lfiber/configs"
	i18n "lfiber/internal/providers/i18n"
	"lfiber/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestI18nProvider(t *testing.T) {
	cfg := &configs.Config{
		I18n: configs.I18nConfig{
			Enabled:         true,
			DefaultLanguage: "en",
			LanguageDir:     "../../../lang",
		},
	}

	_, translator, err := i18n.RegisterI18n(cfg)
	require.NoError(t, err)

	c, app := testkit.AcquireCtx(t)
	defer app.ReleaseCtx(c)

	// Test GetLocale
	assert.Equal(t, "en", translator.GetLocale(c))

	// Test Trans (English default)
	msg := translator.Trans(c, "app.welcome", nil)
	assert.Equal(t, "Welcome to Lunchbox Vending System", msg)

	// Test Trans with specific locale
	msgZh := translator.Trans(c, "app.welcome", nil, "zh-CN")
	assert.Equal(t, "\u6b22\u8fce\u4f7f\u7528\u996d\u76d2\u552e\u8d27\u673a\u7cfb\u7edf", msgZh)
	assert.NotEqual(t, msg, msgZh)

	assert.Equal(t, "Profile", translator.Trans(c, "modules.user.profile", nil))

	choice := translator.Choice(c, "validation.required", 2, map[string]interface{}{"attribute": "email"})
	assert.NotEmpty(t, choice)
}

func TestI18nProvider_DirectoryBundleFailures(t *testing.T) {
	t.Run("DuplicateKey", func(t *testing.T) {
		langDir := t.TempDir()
		writeLangFile(t, langDir, "en", "app.toml", "[\"app.welcome\"]\nother = \"Welcome\"\n")
		writeLangFile(t, langDir, "en", "duplicate.toml", "[\"app.welcome\"]\nother = \"Duplicate\"\n")

		_, _, err := i18n.RegisterI18n(i18nTestConfig(langDir, []string{"en"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate i18n message")
	})

	t.Run("MissingLanguageDirectory", func(t *testing.T) {
		_, _, err := i18n.RegisterI18n(i18nTestConfig(t.TempDir(), []string{"en"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read language directory")
	})

	t.Run("EmptyLanguageDirectory", func(t *testing.T) {
		langDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(langDir, "en"), 0o755))

		_, _, err := i18n.RegisterI18n(i18nTestConfig(langDir, []string{"en"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contains no toml files")
	})

	t.Run("InvalidTOML", func(t *testing.T) {
		langDir := t.TempDir()
		writeLangFile(t, langDir, "en", "app.toml", "[\"app.welcome\"\nother = \"Welcome\"\n")

		_, _, err := i18n.RegisterI18n(i18nTestConfig(langDir, []string{"en"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "initialize i18n translator")
	})
}

func i18nTestConfig(langDir string, supported []string) *configs.Config {
	return &configs.Config{
		I18n: configs.I18nConfig{
			Enabled:            true,
			DefaultLanguage:    "en",
			SupportedLanguages: supported,
			LanguageDir:        langDir,
			CookieName:         "lang",
			CookieMaxAge:       86400,
		},
	}
}

func writeLangFile(t *testing.T, root, locale, name, body string) {
	t.Helper()

	dir := filepath.Join(root, locale)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
}
