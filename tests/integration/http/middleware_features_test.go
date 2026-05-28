package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"lfiber/configs"
	"lfiber/internal/bootstrap"
	exceptions "lfiber/internal/common/exceptions"
	middleware "lfiber/internal/common/middleware"
	"lfiber/internal/common/requests"
	ratelimiter "lfiber/internal/providers/ratelimiter"
	helpers "lfiber/internal/support"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
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
	app := fiber.New(fiber.Config{
		ErrorHandler:    helpers.HandleHTTPError,
		StructValidator: requests.NewStructValidator(),
	})
	app.Post("/validate", func(c fiber.Ctx) error {
		var payload struct {
			Email string `json:"email" validate:"required,email"`
		}
		return requests.Body(c, &payload)
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

func TestOTELMiddleware_RecordsBusinessRouteWithRequestID(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(noop.NewTracerProvider())
	})

	cfg := testkit.NewSQLiteConfig(t)
	cfg.OTEL.TraceEnabled = true
	cfg.OTEL.MetricsEnabled = false
	cfg.OTEL.MetricsPath = "/metrics"

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	middleware.SetupMiddleware(app, cfg)
	app.Get("/api/v1/ping", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/api/v1/ping", nil)
	req.Header.Set("X-Request-ID", "rid-otel")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	spans := recorder.Ended()
	require.NotEmpty(t, spans)

	var found bool
	for _, span := range spans {
		if span.Name() != "GET /api/v1/ping" {
			continue
		}
		found = true
		var requestID string
		for _, attr := range span.Attributes() {
			if string(attr.Key) == "request_id" {
				requestID = attr.Value.AsString()
			}
		}
		assert.Equal(t, "rid-otel", requestID)
	}
	require.True(t, found)
}

func TestOTELMiddleware_SkipsHealthDocsAndMetrics(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(noop.NewTracerProvider())
	})

	cfg := testkit.NewSQLiteConfig(t)
	cfg.OTEL.TraceEnabled = true
	cfg.OTEL.MetricsEnabled = true
	cfg.OTEL.MetricsPath = "/metrics"

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	middleware.SetupMiddleware(app, cfg)
	for _, path := range []string{"/health", "/docs", "/metrics"} {
		app.Get(path, func(c fiber.Ctx) error {
			return c.SendString("ok")
		})
	}

	for _, path := range []string{"/health", "/docs", "/metrics"} {
		resp, err := app.Test(httptest.NewRequest("GET", path, nil))
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	}

	assert.Empty(t, recorder.Ended())
}

func TestOTELMiddleware_AllowsMetricsOnlyMode(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	cfg.OTEL.TraceEnabled = false
	cfg.OTEL.MetricsEnabled = true
	cfg.OTEL.MetricsPath = "/metrics"

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	middleware.SetupMiddleware(app, cfg)
	app.Get("/api/v1/metrics-only", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/metrics-only", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
