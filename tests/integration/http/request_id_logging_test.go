package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	middleware "fiber-starter/internal/common/middleware"
	logging "fiber-starter/internal/providers/logging"
	helpers "fiber-starter/internal/support"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestID_GeneratedAndLogged(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := logging.DefaultLogger
	logging.DefaultLogger = zap.New(core)
	t.Cleanup(func() {
		logging.DefaultLogger = prevLogger
	})

	var resp *http.Response
	app := fiber.New(fiber.Config{
		ErrorHandler: helpers.HandleHTTPError,
	})
	middleware.SetupMiddleware(app, nil)
	app.Get("/boom", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "boom")
	})

	r, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
	require.NoError(t, err)
	resp = r

	requestID := resp.Header.Get("X-Request-ID")
	require.NotEmpty(t, requestID)

	var foundError bool
	t.Logf("Observed logs: %d", len(observed.All()))
	for _, entry := range observed.All() {
		t.Logf("Entry: msg=%s context=%v", entry.Message, entry.ContextMap())
		m := entry.ContextMap()
		if entry.Message == "client_error" || entry.Message == "server_error" {
			foundError = true
			assert.Equal(t, requestID, m["request_id"])
		}
	}

	require.True(t, foundError)
}

func TestRequestID_PreservedAndLogged(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := logging.DefaultLogger
	logging.DefaultLogger = zap.New(core)
	t.Cleanup(func() {
		logging.DefaultLogger = prevLogger
	})

	var resp *http.Response
	app := fiber.New(fiber.Config{
		ErrorHandler: helpers.HandleHTTPError,
	})
	middleware.SetupMiddleware(app, nil)
	app.Get("/boom", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "boom")
	})

	req := httptest.NewRequest("GET", "/boom", nil)
	req.Header.Set("X-Request-ID", "rid-123")
	r, err := app.Test(req)
	require.NoError(t, err)
	resp = r

	requestID := resp.Header.Get("X-Request-ID")
	assert.Equal(t, "rid-123", requestID)

	var foundError bool
	t.Logf("Observed logs: %d", len(observed.All()))
	for _, entry := range observed.All() {
		t.Logf("Entry: msg=%s context=%v", entry.Message, entry.ContextMap())
		m := entry.ContextMap()
		if entry.Message == "client_error" || entry.Message == "server_error" {
			foundError = true
			assert.Equal(t, requestID, m["request_id"])
		}
	}

	require.True(t, foundError)
}

func TestRequestLogging_RedactsSensitiveURLAndHeaders(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := logging.DefaultLogger
	logging.DefaultLogger = zap.New(core)
	t.Cleanup(func() {
		logging.DefaultLogger = prevLogger
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: helpers.HandleHTTPError,
	})
	middleware.SetupMiddleware(app, nil)
	app.Get("/boom", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "boom")
	})

	req := httptest.NewRequest("GET", "/boom?token=secret-token", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var found bool
	for _, entry := range observed.All() {
		if entry.Message != "client_error" {
			continue
		}
		found = true
		fields := entry.ContextMap()
		assert.Contains(t, fields["url"], helpers.RedactionSentinel())
		assert.NotContains(t, fields["url"], "secret-token")
	}

	require.True(t, found)
}
