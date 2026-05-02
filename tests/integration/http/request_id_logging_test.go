package tests

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	middleware "fiber-starter/app/Http/Middleware"
	helpers "fiber-starter/app/Support"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var stdoutMu sync.Mutex

func captureStdout(t *testing.T, fn func()) string {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	return testkit.CaptureOutput(t, "stdout", fn)
}

func TestRequestID_GeneratedAndLogged(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := helpers.Logger
	helpers.Logger = zap.New(core)
	t.Cleanup(func() {
		helpers.Logger = prevLogger
	})

	var resp *http.Response
	out := captureStdout(t, func() {
		app := fiber.New()
		middleware.SetupMiddleware(app, nil)
		app.Get("/boom", func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusBadRequest, "boom")
		})

		r, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
		assert.NoError(t, err)
		resp = r
	})

	requestID := resp.Header.Get("X-Request-ID")
	require.NotEmpty(t, requestID)

	var foundError bool
	for _, entry := range observed.All() {
		m := entry.ContextMap()
		if entry.Message == "http_error" {
			foundError = true
			assert.Equal(t, requestID, m["request_id"])
		}
	}

	require.True(t, foundError)

	assert.Contains(t, out, requestID)
}

func TestRequestID_PreservedAndLogged(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := helpers.Logger
	helpers.Logger = zap.New(core)
	t.Cleanup(func() {
		helpers.Logger = prevLogger
	})

	var resp *http.Response
	out := captureStdout(t, func() {
		app := fiber.New()
		middleware.SetupMiddleware(app, nil)
		app.Get("/boom", func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusBadRequest, "boom")
		})

		req := httptest.NewRequest("GET", "/boom", nil)
		req.Header.Set("X-Request-ID", "rid-123")
		r, err := app.Test(req)
		assert.NoError(t, err)
		resp = r
	})

	requestID := resp.Header.Get("X-Request-ID")
	assert.Equal(t, "rid-123", requestID)

	var foundError bool
	for _, entry := range observed.All() {
		m := entry.ContextMap()
		if entry.Message == "http_error" {
			foundError = true
			assert.Equal(t, requestID, m["request_id"])
		}
	}

	require.True(t, foundError)

	assert.Contains(t, out, requestID)
}
