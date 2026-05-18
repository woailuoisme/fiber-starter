package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fiber-starter/internal/bootstrap"
	providers "fiber-starter/internal/providers"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestHostHeaderMiddleware(t *testing.T) {
	app := fiber.New()

	// Register the middleware manually
	app.Use(func(c fiber.Ctx) error {
		if len(c.Request().Header.Host()) == 0 {
			c.Request().Header.SetHost("127.0.0.1")
		}
		return c.Next()
	})

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(c.Hostname())
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = ""
	req.Header = make(http.Header)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body := testkit.ReadBody(t, resp)
	assert.Contains(t, []string{"127.0.0.1", "localhost"}, body)
}
