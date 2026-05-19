package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"fiber-starter/configs"
	"fiber-starter/internal/bootstrap"
	exceptions "fiber-starter/internal/common/exceptions"
	middleware "fiber-starter/internal/common/middleware"
	"fiber-starter/internal/common/requests"
	ratelimiter "fiber-starter/internal/providers/ratelimiter"
	"fiber-starter/internal/providers/validation"
	helpers "fiber-starter/internal/support"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrottle_EnforcesRequestLimit(t *testing.T) {
	limiter, err := ratelimiter.Register(configs.LimiterConfig{
		Default: "auth",
		Strategies: map[string]configs.RateLimitConfig{
			"auth": {
				Max:    1,
				Window: 60,
			},
		},
	})
	require.NoError(t, err)

	app := fiber.New()
	app.Use(middleware.Throttle(limiter, "auth"))
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

func TestRequestBehavior_InvalidJSONUsesErrorEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{
		JSONDecoder:  json.Unmarshal,
		ErrorHandler: helpers.HandleHTTPError,
	})
	app.Post("/json", func(c fiber.Ctx) error {
		var payload struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(c.Body(), &payload); err != nil {
			return exceptions.BadRequestWithDetails("Invalid request body", err.Error())
		}
		return nil
	})

	resp := testkit.DoRequest(t, app, "POST", "/json", `{"name":`)
	testkit.AssertErrorEnvelope(t, resp, fiber.StatusBadRequest)
}

func TestRequestBehavior_BodyLimitUsesErrorEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Post("/limited-body", func(c fiber.Ctx) error {
		return fiber.ErrRequestEntityTooLarge
	})

	resp := testkit.DoRequest(t, app, "POST", "/limited-body", `{"name":"too-large"}`)
	testkit.AssertErrorEnvelope(t, resp, fiber.StatusRequestEntityTooLarge)
}

func TestRequestBehavior_ValidationFailureUsesErrorEnvelope(t *testing.T) {
	v, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)
	requests.InitValidator(v)
	t.Cleanup(func() {
		requests.InitValidator(nil)
	})

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Post("/validate", func(c fiber.Ctx) error {
		var payload struct {
			Email string `json:"email" validate:"required,email"`
		}
		return requests.BindAndValidateBody(c, &payload)
	})

	resp := testkit.DoRequest(t, app, "POST", "/validate", `{}`)
	testkit.AssertErrorEnvelope(t, resp, fiber.StatusUnprocessableEntity)
}

func TestHTTPStartupAndMiddlewareBaseline(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	start := time.Now()
	app := bootstrap.NewHTTPApp(cfg)
	middleware.SetupMiddleware(app, cfg)
	app.Get("/", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "ok", fiber.Map{"ready": true})
	})
	startupDuration := time.Since(start)

	require.Less(t, startupDuration, 30*time.Second)

	resp := testkit.DoRequest(t, app, "GET", "/", "")
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
	assert.NotEmpty(t, resp.Header.Get("X-Content-Type-Options"))
}
