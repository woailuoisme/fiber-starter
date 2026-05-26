package providers

import (
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
	msg := translator.Trans(c, "validation.required", map[string]interface{}{"attribute": "email"})
	assert.NotEmpty(t, msg)

	// Test Trans with specific locale
	msgZh := translator.Trans(c, "validation.required", map[string]interface{}{"attribute": "email"}, "zh-CN")
	assert.NotEmpty(t, msgZh)

	choice := translator.Choice(c, "validation.required", 2, map[string]interface{}{"attribute": "email"})
	assert.NotEmpty(t, choice)
}
