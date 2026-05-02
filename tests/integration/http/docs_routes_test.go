package tests

import (
	"net/http/httptest"
	"testing"

	controllers "fiber-starter/app/Http/Controllers"
	"fiber-starter/routes"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsRoutes_ExposeScalarAndOpenAPISpec(t *testing.T) {
	app := fiber.New()
	routes.SetupRoutes(app, nil, func(c fiber.Ctx) error {
		return c.Next()
	}, &controllers.AuthController{}, &controllers.UserController{}, &controllers.HealthController{})

	requiredPaths := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/users/",
		"/api/v1/users/profile",
	}
	registeredRoutes := app.GetRoutes(false)
	for _, want := range requiredPaths {
		assert.Truef(t, testkit.HasRoutePath(registeredRoutes, want), "route %s was not registered", want)
	}

	rootResp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, rootResp.StatusCode)
	rootJSON := testkit.ReadBody(t, rootResp)
	assert.Contains(t, rootJSON, `"success":true`)
	assert.Contains(t, rootJSON, `"message":"Welcome to Fiber Starter API"`)
	assert.Contains(t, rootJSON, `"api":"/api/v1"`)

	docsResp, err := app.Test(httptest.NewRequest("GET", "/docs", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, docsResp.StatusCode)
	docsHTML := testkit.ReadBody(t, docsResp)
	assert.Contains(t, docsHTML, "@scalar/api-reference")
	assert.Contains(t, docsHTML, "/openapi.json")

	specResp, err := app.Test(httptest.NewRequest("GET", "/openapi.json", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, specResp.StatusCode)
	assert.Contains(t, testkit.ReadBody(t, specResp), `"openapi": "3.1.0"`)
}
