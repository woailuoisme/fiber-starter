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
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if resp1.StatusCode != fiber.StatusOK {
		t.Fatalf("first request status incorrect: got=%d want=%d", resp1.StatusCode, fiber.StatusOK)
	}

	resp2, err := app.Test(httptest.NewRequest("GET", "/limited", nil))
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if resp2.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("second request status incorrect: got=%d want=%d", resp2.StatusCode, fiber.StatusTooManyRequests)
	}
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
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		t.Fatalf("read first response failed: %v", err)
	}

	req2 := httptest.NewRequest("POST", "/submit", bytes.NewBufferString(`{"name":"demo"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Idempotency-Key", "12345678-1234-1234-1234-123456789012")
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("read second response failed: %v", err)
	}

	if string(body1) != string(body2) {
		t.Fatalf("expected cached response body, got %q want %q", string(body2), string(body1))
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("expected handler to run once, got %d", got)
	}
}

func TestAPIKeyAuth_ValidatesBearerAndHeaderTokens(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.APIKeyAuth("secret-token"))
	app.Get("/secure", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	missingResp, err := app.Test(httptest.NewRequest("GET", "/secure", nil))
	if err != nil {
		t.Fatalf("missing token request failed: %v", err)
	}
	if missingResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("missing token status incorrect: got=%d want=%d", missingResp.StatusCode, fiber.StatusUnauthorized)
	}

	bearerReq := httptest.NewRequest("GET", "/secure", nil)
	bearerReq.Header.Set("Authorization", "Bearer secret-token")
	bearerResp, err := app.Test(bearerReq)
	if err != nil {
		t.Fatalf("bearer token request failed: %v", err)
	}
	if bearerResp.StatusCode != fiber.StatusOK {
		t.Fatalf("bearer token status incorrect: got=%d want=%d", bearerResp.StatusCode, fiber.StatusOK)
	}

	headerReq := httptest.NewRequest("GET", "/secure", nil)
	headerReq.Header.Set("X-API-Key", "secret-token")
	headerResp, err := app.Test(headerReq)
	if err != nil {
		t.Fatalf("header token request failed: %v", err)
	}
	if headerResp.StatusCode != fiber.StatusOK {
		t.Fatalf("header token status incorrect: got=%d want=%d", headerResp.StatusCode, fiber.StatusOK)
	}

	wrongReq := httptest.NewRequest("GET", "/secure", nil)
	wrongReq.Header.Set("Authorization", "Bearer wrong-token")
	wrongResp, err := app.Test(wrongReq)
	if err != nil {
		t.Fatalf("wrong token request failed: %v", err)
	}
	if wrongResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("wrong token status incorrect: got=%d want=%d", wrongResp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestIdempotencyMiddleware_AllowsSafeMethods(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.IdempotencyMiddleware())
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatalf("safe method request failed: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read safe method response failed: %v", err)
	}
	if strings.TrimSpace(string(body)) != "ok" {
		t.Fatalf("safe method response incorrect: got=%q want=%q", string(body), "ok")
	}
}
