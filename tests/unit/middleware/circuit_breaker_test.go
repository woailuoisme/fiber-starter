package tests

import (
	"errors"
	"net/http/httptest"
	"testing"

	"lfiber/configs"
	middleware "lfiber/internal/common/middleware"
	support "lfiber/internal/support"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_DisabledPassesThrough(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: support.HandleHTTPError,
	})

	cfg := &configs.Config{}
	cfg.Security.CircuitBreaker.Enabled = false

	app.Get("/test", middleware.CircuitBreaker(cfg), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCircuitBreaker_TripsAndReturns503(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: support.HandleHTTPError,
	})

	cfg := &configs.Config{}
	cfg.Security.CircuitBreaker.Enabled = true
	cfg.Security.CircuitBreaker.FailureThreshold = 2
	cfg.Security.CircuitBreaker.Timeout = 10
	cfg.Security.CircuitBreaker.SuccessThreshold = 2
	cfg.Security.CircuitBreaker.HalfOpenMaxConcurrent = 1

	// Handler that returns error
	app.Get("/test", middleware.CircuitBreaker(cfg), func(c fiber.Ctx) error {
		return errors.New("something went wrong")
	})

	// 1st request: returns 500 (something went wrong)
	req1 := httptest.NewRequest("GET", "/test", nil)
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp1.StatusCode)

	// 2nd request: returns 500 (something went wrong)
	req2 := httptest.NewRequest("GET", "/test", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp2.StatusCode)

	// 3rd request: since failure threshold is 2, circuit is now OPEN.
	// It should return 503 Service Unavailable directly without hitting the handler.
	req3 := httptest.NewRequest("GET", "/test", nil)
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp3.StatusCode)
}
