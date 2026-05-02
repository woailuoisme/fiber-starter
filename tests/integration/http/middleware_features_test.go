package tests

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	middleware "fiber-starter/app/Http/Middleware"
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthLimiter_EnforcesRequestLimit(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.AuthLimiter(&config.Config{
		Security: config.SecurityConfig{
			RateLimit: config.RateLimitConfig{
				Max:    1,
				Window: 60,
			},
		},
	}))
	app.Get("/limited", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp1, err := app.Test(httptest.NewRequest("GET", "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp1.StatusCode)

	resp2, err := app.Test(httptest.NewRequest("GET", "/limited", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp2.StatusCode)
}

func TestIdempotencyMiddleware_ReusesCachedResponse(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.IdempotencyMiddleware())

	var count int32
	app.Post("/submit", func(c fiber.Ctx) error {
		n := atomic.AddInt32(&count, 1)
		return c.JSON(fiber.Map{"count": n})
	})

	req1 := httptest.NewRequest("POST", "/submit", bytes.NewBufferString(`{"name":"demo"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Idempotency-Key", "12345678-1234-1234-1234-123456789012")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)

	req2 := httptest.NewRequest("POST", "/submit", bytes.NewBufferString(`{"name":"demo"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Idempotency-Key", "12345678-1234-1234-1234-123456789012")
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	assert.Equal(t, string(body1), string(body2))
	assert.EqualValues(t, 1, atomic.LoadInt32(&count))
}

func TestAPIKeyAuth_ValidatesBearerAndHeaderTokens(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.APIKeyAuth("secret-token"))
	app.Get("/secure", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	missingResp, err := app.Test(httptest.NewRequest("GET", "/secure", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, missingResp.StatusCode)

	bearerReq := httptest.NewRequest("GET", "/secure", nil)
	bearerReq.Header.Set("Authorization", "Bearer secret-token")
	bearerResp, err := app.Test(bearerReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, bearerResp.StatusCode)

	headerReq := httptest.NewRequest("GET", "/secure", nil)
	headerReq.Header.Set("X-API-Key", "secret-token")
	headerResp, err := app.Test(headerReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, headerResp.StatusCode)

	wrongReq := httptest.NewRequest("GET", "/secure", nil)
	wrongReq.Header.Set("Authorization", "Bearer wrong-token")
	wrongResp, err := app.Test(wrongReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, wrongResp.StatusCode)
}

func TestIdempotencyMiddleware_AllowsSafeMethods(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.IdempotencyMiddleware())
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "ok", strings.TrimSpace(string(body)))
}
