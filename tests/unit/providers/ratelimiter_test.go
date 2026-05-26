package providers_test

import (
	"net/http/httptest"
	"testing"

	"lfiber/configs"
	ratelimiter "lfiber/internal/providers/ratelimiter"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterProvider_Register(t *testing.T) {
	service, err := ratelimiter.Register(configs.LimiterConfig{
		Default: "default",
		Strategies: map[string]configs.RateLimitConfig{
			"default": {
				Max:    2,
				Window: 60,
			},
			"auth": {
				Max:    1,
				Window: 60,
			},
		},
	})
	require.NoError(t, err)

	require.NotNil(t, service)

	strategy, ok := service.Strategy("auth")
	require.True(t, ok)
	assert.Equal(t, 1, strategy.Max)
	assert.Equal(t, 60, strategy.Window)

	_, ok = service.Strategy("missing")
	assert.False(t, ok)

	limitedApp := fiber.New()
	limitedApp.Use(service.Middleware("auth"))
	limitedApp.Get("/limited", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp1, err := limitedApp.Test(httptest.NewRequest("GET", "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp1.StatusCode)

	resp2, err := limitedApp.Test(httptest.NewRequest("GET", "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp2.StatusCode)

	fallbackApp := fiber.New()
	fallbackApp.Use(service.Middleware("missing"))
	fallbackApp.Get("/fallback", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	fallback1, err := fallbackApp.Test(httptest.NewRequest("GET", "/fallback", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, fallback1.StatusCode)

	fallback2, err := fallbackApp.Test(httptest.NewRequest("GET", "/fallback", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, fallback2.StatusCode)

	fallback3, err := fallbackApp.Test(httptest.NewRequest("GET", "/fallback", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, fallback3.StatusCode)
}
