package tests

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	requests "fiber-starter/app/Http/Requests"
	helpers "fiber-starter/app/Support"
	supporti18n "fiber-starter/app/Support/i18n"
	"fiber-starter/config"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestI18n_LocalizeLanguagePriorityAndCookiePersistence(t *testing.T) {
	initTestI18n(t)

	app := fiber.New()
	app.Get("/welcome", func(c fiber.Ctx) error {
		message, err := supporti18n.Localize(c, "welcome.message")
		if err != nil {
			return err
		}
		return c.SendString(message)
	})

	zhResp := testkit.DoRequest(t, app, "GET", "/welcome?lang=zh-CN", "")
	require.Equal(t, fiber.StatusOK, zhResp.StatusCode)
	assert.Contains(t, testkit.ReadBody(t, zhResp), "\u6b22\u8fce\u4f7f\u7528\u996d\u76d2\u552e\u8d27\u673a\u7cfb\u7edf")
	value, ok := cookieValue(zhResp, "lang")
	require.True(t, ok)
	assert.Equal(t, "zh-CN", value)

	cookieReq, err := http.NewRequest("GET", "/welcome", nil)
	require.NoError(t, err)
	cookieReq.AddCookie(&http.Cookie{Name: "lang", Value: "zh-CN"})
	cookieResp, err := app.Test(cookieReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, cookieResp.StatusCode)
	assert.Contains(t, testkit.ReadBody(t, cookieResp), "\u6b22\u8fce\u4f7f\u7528\u996d\u76d2\u552e\u8d27\u673a\u7cfb\u7edf")

	enReq, err := http.NewRequest("GET", "/welcome", nil)
	require.NoError(t, err)
	enReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
	enResp, err := app.Test(enReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, enResp.StatusCode)
	assert.Contains(t, testkit.ReadBody(t, enResp), "Welcome to Lunchbox Vending System")
}

func TestI18n_ValidationErrorsUseRequestLanguage(t *testing.T) {
	initTestI18n(t)

	app := fiber.New()
	app.Post("/validate", func(c fiber.Ctx) error {
		var req struct {
			Email string `json:"email" validate:"required,email"`
		}

		if err := requests.BindAndValidateBody(c, &req); err != nil {
			return helpers.HandleAppError(c, err)
		}

		return c.SendString("ok")
	})

	resp := testkit.DoRequest(t, app, "POST", "/validate?lang=zh-CN", `{"email":""}`)
	require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	payload := testkit.JSONBody(t, resp)
	assert.Equal(t, false, payload["success"])
	assert.EqualValues(t, fiber.StatusUnprocessableEntity, payload["code"])
	assert.Equal(t, "Validation failed", payload["message"])

	errorsMap := payload["errors"].(map[string]any)
	emailErrors, ok := errorsMap["email"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, emailErrors)
	assert.Equal(t, "\u90ae\u7bb1 \u662f\u5fc5\u586b\u9879", strings.TrimSpace(emailErrors[0].(string)))
}

func initTestI18n(t *testing.T) {
	t.Helper()

	langDir := filepath.Join(testkit.RepoRoot(t), "lang")

	_, err := supporti18n.Init(&config.I18nConfig{
		DefaultLanguage:    "zh-CN",
		SupportedLanguages: []string{"en", "zh-CN"},
		LanguageDir:        langDir,
		CookieName:         "lang",
		CookieMaxAge:       86400,
	})
	require.NoError(t, err)
}

func cookieValue(resp *http.Response, name string) (string, bool) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			return cookie.Value, true
		}
	}
	return "", false
}
