package monitoring

import (
	"os"
	"path/filepath"
	"runtime"

	helpers "fiber-starter/internal/support"

	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers public, non-versioned routes like home, docs, health, ready, and monitor.
func RegisterRoutes(app *fiber.App, healthController *HealthController) {
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

	app.Get("/openapi.json", func(c fiber.Ctx) error {
		return c.SendFile(openAPISpecPath())
	})
	app.Get("/docs", scalarDocs)

	app.Get("/health", healthController.Health)
	app.Get("/ready", healthController.Ready)
	app.Get("/monitor", monitor.New())
}

func openAPISpecPath() string {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		baseDir := filepath.Dir(file)
		repoRoot := filepath.Clean(filepath.Join(baseDir, "..", "..", ".."))
		path := filepath.Join(repoRoot, "docs", "openapi.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	path := filepath.Join("docs", "openapi.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	parentPath := filepath.Join("..", "docs", "openapi.json")
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath
	}

	return path
}

func scalarDocs(c fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(`<!doctype html>
<html lang="en">
  <head>
    <title>Fiber Starter API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body { margin: 0; padding: 0; }
      #app { height: 100vh; width: 100vw; }
      .loading {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100vh;
        font-family: sans-serif;
        background: #0f172a;
        color: white;
      }
    </style>
  </head>
  <body>
    <div id="app"><div class="loading">Loading API Reference...</div></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      (function() {
        const init = () => {
          if (typeof Scalar === 'undefined') {
            document.getElementById('app').innerHTML = '<div class="loading">Error: Scalar library failed to load. Please check your internet connection or CDN availability.</div>';
            return;
          }
          Scalar.createApiReference('#app', {
            url: '/openapi.json?t=' + new Date().getTime(),
            layout: 'modern',
            theme: 'deepSpace',
            hideModels: true,
            hideDownloadButton: true,
            documentDownloadType: 'none',
            showDeveloperTools: false,
            showToolbar: false,
            telemetry: false,
            searchHotKey: 'k',
            defaultOpenFirstTag: true,
            darkMode: true,
          })
        };
        if (document.readyState === 'complete') {
          init();
        } else {
          window.addEventListener('load', init);
        }
      })();
    </script>
  </body>
</html>`)
}
