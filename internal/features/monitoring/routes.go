package monitoring

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	helpers "lfiber/internal/support"

	"github.com/gofiber/contrib/v3/monitor"
	swaggerui "github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers public, non-versioned routes like home, docs, health, ready, and monitor.
func RegisterRoutes(app *fiber.App, h *HealthController) {
	app.Get("/", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "Welcome to lfiber API", fiber.Map{
			"version": "1.0.0",
			"docs":    "/docs",
			"scalar":  "/docs/scalar",
			"openapi": "/openapi.json",
			"health":  "/health",
			"ready":   "/ready",
			"monitor": "/monitor",
			"api":     "/api/v1",
		})
	})
	swaggerSpec := mustReadSwaggerSpec()
	swaggerHandler := swaggerui.New(swaggerui.Config{
		BasePath:    "/",
		FilePath:    "openapi.json",
		FileContent: swaggerSpec,
		Path:        "docs",
		Title:       "lfiber API Reference",
	})

	app.Get("/docs", swaggerHandler)

	app.Get("/docs/scalar", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(`<!doctype html>
<html>
  <head>
    <title>Scalar API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body { margin: 0; }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/openapi.json">
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference" crossorigin></script>
  </body>
</html>`)
	})
	app.Get("/openapi.json", swaggerHandler)
	app.Get("/health", h.Health)
	app.Get("/ready", h.Ready)
	app.Get("/monitor", monitor.New())
}

func openAPISpecPath() string {
	var candidates []string
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
		candidates = append(candidates, filepath.Join(root, "docs", "openapi.json"))
	}
	candidates = append(
		candidates,
		filepath.Join("docs", "openapi.json"),
		filepath.Join("..", "docs", "openapi.json"),
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join("docs", "openapi.json")
}

func mustReadSwaggerSpec() []byte {
	specPath := openAPISpecPath()
	//nolint:gosec // specPath points to repo-generated Swagger output, not user input.
	spec, err := os.ReadFile(specPath)
	if err != nil {
		panic(fmt.Errorf("failed to read swagger spec %q: %w", specPath, err))
	}
	return spec
}
