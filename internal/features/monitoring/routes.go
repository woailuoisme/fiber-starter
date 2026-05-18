package monitoring

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"fiber-starter/internal/support/appctx"
	helpers "fiber-starter/internal/support"

	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers public, non-versioned routes like home, docs, health, ready, and monitor.
func RegisterRoutes(app *fiber.App, h *HealthController) {
	app.Get("/", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "Welcome to Fiber Starter API", fiber.Map{
			"version": "1.0.0",
			"docs":    "/docs",
			"openapi": "/openapi.json",
			"health":  "/health",
			"ready":   "/ready",
			"monitor": "/monitor",
			"api":     "/api/v1",
		})
	})
	app.Get("/openapi.json", func(c fiber.Ctx) error { return c.SendFile(openAPISpecPath()) })
	app.Get("/docs", redocDocs)
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
	candidates = append(candidates,
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

func redocDocs(c fiber.Ctx) error {
	specURL := "/openapi.json"
	if rt := appctx.App(); rt != nil {
		if cfg := rt.AppConfig(); cfg != nil && cfg.App.URL != "" {
			specURL = strings.TrimSuffix(cfg.App.URL, "/") + "/openapi.json"
		}
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	html := strings.ReplaceAll(redocHTMLTemplate, "{{.SpecURL}}", specURL)
	return c.SendString(html)
}

const redocHTMLTemplate = `<!doctype html>
<html lang="en">
<head>
  <title>Fiber Starter API Reference</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <style>
    body { margin: 0; padding: 0; }
    /* Hide the "API docs by Redocly" watermark */
    a[href^="https://github.com/Redocly/redoc"],
    a[href^="https://redocly.com/redoc"],
    a[href*="redocly.com"] {
      display: none !important;
    }
  </style>
</head>
<body>
  <redoc id="redoc-container"></redoc>
  <script src="https://cdn.jsdelivr.net/npm/redoc@2.5.2/bundles/redoc.standalone.js"></script>
  <script>
    Redoc.init("{{.SpecURL}}", {
      sortRequiredPropsFirst: true,
      expandResponses: "200,201",
      jsonSamplesExpandLevel: 3,
      pathInMiddlePanel: true,
      nativeScrollbars: true,
      hideHostname: true,
      lazyRendering: true,
      hideDownloadButtons: true,
      sanitize: true
    }, document.getElementById('redoc-container'));
  </script>
</body>
</html>`
