package tests

import (
	"net/http/httptest"
	"testing"

	"fiber-starter/internal/bootstrap"
	providers "fiber-starter/internal/providers"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsRoutes_ExposeRedocAndOpenAPISpec(t *testing.T) {
	t.Setenv("I18N_LANGUAGE_DIR", testkit.RepoRoot(t)+"/lang")

	runtime, err := providers.Build()
	require.NoError(t, err)
	defer func() {
		_ = runtime.Close()
	}()

	app := fiber.New()
	err = bootstrap.SetupApplicationRoutes(app)
	require.NoError(t, err)

	requiredPaths := []string{
		"/api/v1/auth/sign-up",
		"/api/v1/auth/sign-up/verify",
		"/api/v1/auth/sign-in",
		"/api/v1/auth/reset-password",
		"/api/v1/auth/reset-password/verify",
		"/api/v1/auth/reset-password/confirm",
		"/api/v1/auth/session",
		"/api/v1/users/",
		"/api/v1/users/profile",
	}
	registeredRoutes := app.GetRoutes(false)
	for _, want := range requiredPaths {
		assert.Truef(t, testkit.HasRoutePath(registeredRoutes, want), "route %s was not registered", want)
	}

	rootResp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)
	defer rootResp.Body.Close()
	require.Equal(t, fiber.StatusOK, rootResp.StatusCode)
	rootJSON := testkit.ReadBody(t, rootResp)
	assert.Contains(t, rootJSON, `"success":true`)
	assert.Contains(t, rootJSON, `"message":"Welcome to Fiber Starter API"`)
	assert.Contains(t, rootJSON, `"api":"/api/v1"`)

	docsResp, err := app.Test(httptest.NewRequest("GET", "/docs", nil))
	require.NoError(t, err)
	defer docsResp.Body.Close()
	require.Equal(t, fiber.StatusOK, docsResp.StatusCode)
	docsHTML := testkit.ReadBody(t, docsResp)
	assert.Contains(t, docsHTML, "<redoc")
	assert.Contains(t, docsHTML, "redoc.standalone.js")
	assert.Contains(t, docsHTML, "/openapi.json")

	specResp, err := app.Test(httptest.NewRequest("GET", "/openapi.json", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, specResp.StatusCode)
	defer specResp.Body.Close()

	specJSON := testkit.ReadBody(t, specResp)
	assert.Contains(t, specJSON, `"swagger": "2.0"`)
}
