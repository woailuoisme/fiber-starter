package tests

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"fiber-starter/internal/transport/http/routers"

	"github.com/gofiber/fiber/v3"
)

func TestDocsRoutes_ExposeScalarAndOpenAPISpec(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	routers.SetupRoutes(app, func(c fiber.Ctx) error {
		return c.Next()
	}, nil, nil, nil, nil)

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
	if !strings.Contains(string(specBody), `"swagger"`) {
		t.Fatalf("/openapi.json did not return generated swagger spec")
	}

	monitorResp, err := app.Test(httptest.NewRequest("GET", "/monitor", nil))
	if err != nil {
		t.Fatalf("Request /monitor failed: %v", err)
	}
	if monitorResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("/monitor should not be exposed: got=%d want=%d", monitorResp.StatusCode, fiber.StatusNotFound)
	}
}
