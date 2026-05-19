package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fiber-starter/internal/bootstrap"
	middleware "fiber-starter/internal/common/middleware"
	providers "fiber-starter/internal/providers"
	logging "fiber-starter/internal/providers/logging"
	helpers "fiber-starter/internal/support"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger_405EmptyHost(t *testing.T) {
	t.Setenv("I18N_LANGUAGE_DIR", testkit.RepoRoot(t)+"/lang")

	runtime, err := providers.Build()
	require.NoError(t, err)
	defer func() {
		_ = runtime.Close()
	}()

	app := bootstrap.NewHTTPApp(runtime.Config)
	err = bootstrap.SetupApplicationRoutes(app)
	require.NoError(t, err)

	t.Log("--- REGISTERED ROUTES ---")
	for _, route := range app.GetRoutes() {
		t.Logf("[%s] %s -> %s", route.Method, route.Path, route.Name)
	}
	t.Log("-------------------------")

	// 1. Standard GET / with Host header
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("Host", "localhost:3000")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp1.StatusCode)

	// 1b. GET / with Host: 127.0.0.1
	req1b := httptest.NewRequest("GET", "/", nil)
	req1b.Header.Set("Host", "127.0.0.1")
	resp1b, err := app.Test(req1b)
	require.NoError(t, err)
	defer resp1b.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp1b.StatusCode)

	// 2. GET / with empty Host header
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Host = ""
	req2.Header = make(http.Header) // Completely empty headers!
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	t.Logf("Response code for empty Host: %d", resp2.StatusCode)
	body := testkit.ReadBody(t, resp2)
	t.Logf("Response body for empty Host: %s", body)
	assert.Equal(t, fiber.StatusOK, resp2.StatusCode)

	// 3. POST / to intentionally trigger 405 Method Not Allowed
	req3 := httptest.NewRequest("POST", "/", nil)
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()

	t.Logf("Response code for POST / (should be 405): %d", resp3.StatusCode)
	t.Logf("Allow header for POST /: %s", resp3.Header.Get("Allow"))
	body3 := testkit.ReadBody(t, resp3)
	t.Logf("Response body for POST /: %s", body3)
}

func TestLogger_405DiagnosticFiltersMiddlewareRoutes(t *testing.T) {
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
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)

	var diagnostics map[string]interface{}
	for _, entry := range observed.All() {
		if entry.Message == "405_diagnostic_details" {
			diagnostics = entry.ContextMap()
			break
		}
	}

	require.NotNil(t, diagnostics)
	assert.Equal(t, "GET, HEAD", diagnostics["allow"])
	assert.Equal(t, "[GET] / | [HEAD] /", diagnostics["matched_routes"])
	assert.NotContains(t, diagnostics, "Authorization")
}

func TestLogger_SkipsSyntheticGet405FromMiddlewareTraversal(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := logging.DefaultLogger
	logging.DefaultLogger = zap.New(core)
	t.Cleanup(func() {
		logging.DefaultLogger = prevLogger
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: helpers.HandleHTTPError,
	})
	middleware.SetupLogger(app)
	app.Use(func(c fiber.Ctx) error {
		c.Set(fiber.HeaderAllow, fiber.MethodHead)
		return fiber.ErrMethodNotAllowed
	})
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)

	for _, entry := range observed.All() {
		assert.NotEqual(t, "405_diagnostic_details", entry.Message)
		assert.NotEqual(t, "client_error", entry.Message)
	}
}
