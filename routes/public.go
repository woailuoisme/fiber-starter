package routes

import (
	"os"
	"path/filepath"
	"runtime"

	controllers "fiber-starter/app/Http/Controllers"
	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
)

// registerPublicRoutes registers public, non-versioned routes.
func registerPublicRoutes(app *fiber.App, healthController *controllers.HealthController) {
	app.Get("/", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "Welcome to Fiber Starter API", fiber.Map{
			"version": "1.0.0",
			"docs":    "/docs",
			"openapi": "/openapi.json",
			"health":  "/health",
			"ready":   "/ready",
			"api":     "/api/v1",
		})
	})

	app.Get("/openapi.json", func(c fiber.Ctx) error {
		return c.SendFile(openAPISpecPath())
	})
	app.Get("/docs", scalarDocs)

	// Health Check：给 Docker / K8s / 负载均衡器使用的存活探针。
	// 使用方式：调用 /health 获取基础存活状态，/ready 获取就绪状态。
	app.Get("/health", healthController.Health)
	app.Get("/ready", healthController.Ready)
}

func openAPISpecPath() string {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		baseDir := filepath.Dir(file)
		repoRoot := filepath.Clean(filepath.Join(baseDir, ".."))
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
<html>
  <head>
    <title>Fiber Starter API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>body{margin:0}</style>
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.json',
        layout: 'modern',
        theme: 'default'
      })
    </script>
  </body>
</html>`)
}
