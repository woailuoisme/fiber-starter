package tests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"lfiber/internal/bootstrap"
	providers "lfiber/internal/providers"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsRoutes_ExposeSwaggerUIAndOpenAPISpec(t *testing.T) {
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
	assert.Contains(t, rootJSON, `"message":"Welcome to lfiber API"`)
	assert.Contains(t, rootJSON, `"api":"/api/v1"`)
	assert.Contains(t, rootJSON, `"scalar":"/docs/scalar"`)

	docsResp, err := app.Test(httptest.NewRequest("GET", "/docs", nil))
	require.NoError(t, err)
	defer docsResp.Body.Close()
	require.Equal(t, fiber.StatusOK, docsResp.StatusCode)
	docsHTML := testkit.ReadBody(t, docsResp)
	assert.Contains(t, strings.ToLower(docsHTML), "swagger-ui")
	assert.Contains(t, docsHTML, "SwaggerUIBundle")
	assert.Contains(t, docsHTML, "/openapi.json")

	scalarResp, err := app.Test(httptest.NewRequest("GET", "/docs/scalar", nil))
	require.NoError(t, err)
	defer scalarResp.Body.Close()
	require.Equal(t, fiber.StatusOK, scalarResp.StatusCode)
	scalarHTML := testkit.ReadBody(t, scalarResp)
	assert.Contains(t, scalarHTML, `src="https://cdn.jsdelivr.net/npm/@scalar/api-reference" crossorigin`)
	assert.Contains(t, scalarHTML, "data-url=\"/openapi.json\"")

	specResp, err := app.Test(httptest.NewRequest("GET", "/openapi.json", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, specResp.StatusCode)
	defer specResp.Body.Close()

	specJSON := testkit.ReadBody(t, specResp)
	assert.Contains(t, specJSON, `"swagger": "2.0"`)
	assertSuccessSchemasHideErrorFields(t, specJSON)
	assertNoDataSuccessSchemasHideDataField(t, specJSON)
}

func assertSuccessSchemasHideErrorFields(t *testing.T, specJSON string) {
	t.Helper()

	var spec map[string]any
	require.NoError(t, json.Unmarshal([]byte(specJSON), &spec))

	definitions, ok := spec["definitions"].(map[string]any)
	require.True(t, ok)
	successDefinition, ok := definitions["APISuccessResponse"].(map[string]any)
	require.True(t, ok)
	successProperties, ok := successDefinition["properties"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, successProperties, "errors")
	assert.NotContains(t, successProperties, "exception")

	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok)
	for _, rawPath := range paths {
		methods, ok := rawPath.(map[string]any)
		require.True(t, ok)
		for _, rawOperation := range methods {
			operation, ok := rawOperation.(map[string]any)
			require.True(t, ok)
			responses, ok := operation["responses"].(map[string]any)
			require.True(t, ok)
			for status, rawResponse := range responses {
				if !strings.HasPrefix(status, "2") {
					continue
				}
				responseBytes, err := json.Marshal(rawResponse)
				require.NoError(t, err)
				responseJSON := string(responseBytes)
				assert.NotContains(t, responseJSON, "APIResponse")
				assert.NotContains(t, responseJSON, "errors")
				assert.NotContains(t, responseJSON, "exception")
			}
		}
	}
}

func assertNoDataSuccessSchemasHideDataField(t *testing.T, specJSON string) {
	t.Helper()

	var spec map[string]any
	require.NoError(t, json.Unmarshal([]byte(specJSON), &spec))

	definitions, ok := spec["definitions"].(map[string]any)
	require.True(t, ok)
	noDataDefinition, ok := definitions["APISuccessNoDataResponse"].(map[string]any)
	require.True(t, ok)
	noDataProperties, ok := noDataDefinition["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, noDataProperties, "success")
	assert.Contains(t, noDataProperties, "code")
	assert.Contains(t, noDataProperties, "message")
	assert.NotContains(t, noDataProperties, "data")
	assert.NotContains(t, noDataProperties, "errors")
	assert.NotContains(t, noDataProperties, "exception")

	for _, path := range []string{
		"post /api/v1/auth/sign-out",
		"post /api/v1/auth/change-password",
		"post /api/v1/auth/reset-password",
		"post /api/v1/auth/reset-password/confirm",
		"delete /api/v1/users/{id}",
	} {
		schema := successResponseSchema(t, spec, path)
		ref, ok := schema["$ref"].(string)
		require.Truef(t, ok, "success response for %s should be a direct $ref", path)
		assert.Equal(t, "#/definitions/APISuccessNoDataResponse", ref)
	}
}

func successResponseSchema(t *testing.T, spec map[string]any, path string) map[string]any {
	t.Helper()

	parts := strings.SplitN(path, " ", 2)
	require.Len(t, parts, 2)
	method := strings.ToLower(parts[0])
	pathKey := parts[1]

	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok)
	pathItem, ok := paths[pathKey].(map[string]any)
	require.Truef(t, ok, "path %s missing from spec", pathKey)

	operation, ok := pathItem[method].(map[string]any)
	require.Truef(t, ok, "method %s missing from spec for path %s", method, pathKey)
	responses, ok := operation["responses"].(map[string]any)
	require.True(t, ok)
	response, ok := responses["200"].(map[string]any)
	require.True(t, ok)
	schema, ok := response["schema"].(map[string]any)
	require.True(t, ok)
	return schema
}
