package monitoring

import (
	"encoding/json"
	"net/url"

	"lfiber/docs"
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
	swaggerSpec := injectSwaggerHost(mustReadSwaggerSpec(), h.cfg.App.URL)
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

func mustReadSwaggerSpec() []byte {
	return docs.OpenAPISpec
}

// injectSwaggerHost 解析 appURL 并动态修改 OpenAPI Spec 的 host 和 schemes 属性，
// 这样在 Swagger UI 等组件渲染时可以正确显示并调用当前环境的基准 URL 接口。
func injectSwaggerHost(spec []byte, appURL string) []byte {
	if appURL == "" {
		return spec
	}
	u, err := url.Parse(appURL)
	if err != nil {
		return spec
	}

	var data map[string]any
	if err := json.Unmarshal(spec, &data); err != nil {
		return spec
	}

	data["host"] = u.Host

	if u.Scheme != "" {
		data["schemes"] = []string{u.Scheme}
	}

	newSpec, err := json.Marshal(data)
	if err != nil {
		return spec
	}
	return newSpec
}
