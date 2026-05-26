package monitoring

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	helpers "lfiber/internal/support"
	"lfiber/internal/support/otel"

	"github.com/gofiber/contrib/v3/monitor"
	swaggerui "github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		Next: func(c fiber.Ctx) bool {
			// 非 Debug 模式下，禁用并隐藏文档路由
			return !h.cfg.App.Debug
		},
		CacheAge: func() int {
			if h.cfg.App.Debug {
				return 0 // 开发调试时禁用缓存
			}
			return 3600 // 生产环境下允许浏览器缓存 1 小时
		}(),
	})

	app.Get("/docs", swaggerHandler)

	app.Get("/docs/scalar", func(c fiber.Ctx) error {
		if !h.cfg.App.Debug {
			return c.Next()
		}
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
      data-url="/openapi.json"
      data-configuration='{
        "theme": "purple",
        "layout": "modern",
        "searchHotKey": "k",
        "showSidebar": true,
        "persistCredentials": true,
        "hideModels": true
      }'>
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference" crossorigin></script>
  </body>
</html>`)
	})
	app.Get("/openapi.json", swaggerHandler)
	app.Get("/health", h.Health)
	app.Get("/ready", h.Ready)
	app.Get("/monitor", monitor.New())

	if h.cfg.OTEL.MetricsEnabled && otel.GlobalPrometheusExporter != nil {
		metricsPath := h.cfg.OTEL.MetricsPath
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		// OTel prometheus exporter registers to the default prometheus registry;
		// promhttp.Handler() serves that registry over HTTP.
		app.Get(metricsPath, adaptor.HTTPHandler(promhttp.Handler()))
	}
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
