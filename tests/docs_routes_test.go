package tests

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"fiber-starter/app/Http/Controllers"
	"fiber-starter/routes"

	"github.com/gofiber/fiber/v3"
)

func TestDocsRoutes_ExposeScalarAndOpenAPISpec(t *testing.T) {
	app := fiber.New()
	routes.SetupRoutes(app, nil, func(c fiber.Ctx) error {
		return c.Next()
	}, &controllers.AuthController{}, &controllers.UserController{}, &controllers.HealthController{})

	rootResp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("Request / failed: %v", err)
	}
	if rootResp.StatusCode != fiber.StatusOK {
		t.Fatalf("/ status code incorrect: got=%d want=%d", rootResp.StatusCode, fiber.StatusOK)
	}
	rootBody, err := io.ReadAll(rootResp.Body)
	if err != nil {
		t.Fatalf("Read / body failed: %v", err)
	}
	rootJSON := string(rootBody)
	if !strings.Contains(rootJSON, `"success":true`) || !strings.Contains(rootJSON, `"message":"Welcome to Fiber Starter API"`) {
		t.Fatalf("/ did not return unified success response")
	}
	if !strings.Contains(rootJSON, `"api":"/api/v1"`) {
		t.Fatalf("/ response did not include api version link")
	}

	docsResp, err := app.Test(httptest.NewRequest("GET", "/docs", nil))
	if err != nil {
		t.Fatalf("Request /docs failed: %v", err)
	}
	if docsResp.StatusCode != fiber.StatusOK {
		t.Fatalf("/docs status code incorrect: got=%d want=%d", docsResp.StatusCode, fiber.StatusOK)
	}
	docsBody, err := io.ReadAll(docsResp.Body)
	if err != nil {
		t.Fatalf("Read /docs body failed: %v", err)
	}
	docsHTML := string(docsBody)
	if !strings.Contains(docsHTML, "@scalar/api-reference") || !strings.Contains(docsHTML, "/openapi.json") {
		t.Fatalf("/docs did not render Scalar document shell")
	}

	specResp, err := app.Test(httptest.NewRequest("GET", "/openapi.json", nil))
	if err != nil {
		t.Fatalf("Request /openapi.json failed: %v", err)
	}
	if specResp.StatusCode != fiber.StatusOK {
		t.Fatalf("/openapi.json status code incorrect: got=%d want=%d", specResp.StatusCode, fiber.StatusOK)
	}
	specBody, err := io.ReadAll(specResp.Body)
	if err != nil {
		t.Fatalf("Read /openapi.json body failed: %v", err)
	}
	if !strings.Contains(string(specBody), `"openapi": "3.1.0"`) {
		t.Fatalf("/openapi.json did not return an OpenAPI 3.1 document")
	}

	monitorResp, err := app.Test(httptest.NewRequest("GET", "/monitor", nil))
	if err != nil {
		t.Fatalf("Request /monitor failed: %v", err)
	}
	if monitorResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("/monitor should not be exposed: got=%d want=%d", monitorResp.StatusCode, fiber.StatusNotFound)
	}
}
