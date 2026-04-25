package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/timeout"
)

// SetupMiddleware 配置所有中间件
func SetupMiddleware(app *fiber.App) {
	setupCoreMiddleware(app)
	setupSecurityMiddleware(app)
	setupTimeoutMiddleware(app)
	app.Use(ErrorHandler)
}

func setupCoreMiddleware(app *fiber.App) {
	if _, err := os.Stat("./public/favicon.ico"); err == nil {
		app.Use(favicon.New(favicon.Config{
			File: "./public/favicon.ico",
			URL:  "/favicon.ico",
		}))
	}

	app.Get("/favicon.svg", func(c fiber.Ctx) error {
		return c.SendFile("./public/favicon.svg")
	})

	app.Use(requestid.New())
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"request_id": func(output logger.Buffer, c fiber.Ctx, _ *logger.Data, _ string) (int, error) {
				return output.WriteString(requestid.FromContext(c))
			},
		},
		Stream: os.Stdout,
		Format: "${time} ${ip} ${request_id} ${status} - ${latency} ${method} ${path} ${error}\n",
	}))
}

func setupSecurityMiddleware(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	app.Use(helmet.New())
}

func setupTimeoutMiddleware(app *fiber.App) {
	app.Use(timeout.New(func(c fiber.Ctx) error {
		return c.Next()
	}, timeout.Config{
		Timeout: 30 * time.Second,
		Next: func(c fiber.Ctx) bool {
			return isTimeoutExemptPath(c.Path())
		},
		OnTimeout: func(c fiber.Ctx) error {
			return HandleError(c, fiber.ErrRequestTimeout)
		},
	}))
}

// SetupAuthMiddleware 配置认证相关的中间件
func SetupAuthMiddleware(_ *fiber.App) {}

// SetupTimeoutRedirect 配置超时重定向中间件
func SetupTimeoutRedirect(app *fiber.App) {
	app.Use(func(c fiber.Ctx) error {
		if redirectURL := resolveRedirectURL(c); redirectURL != "" {
			return c.Redirect().Status(fiber.StatusMovedPermanently).To(redirectURL)
		}
		return c.Next()
	})
}

func isTimeoutExemptPath(path string) bool {
	switch path {
	case "/health", "/ready", "/docs", "/openapi.json":
		return true
	default:
		return false
	}
}

func resolveRedirectURL(c fiber.Ctx) string {
	if isMaintenanceMode() && c.Path() != "/maintenance" {
		return "/maintenance"
	}

	if c.Path() == "/old-path" {
		return "/new-path"
	}

	userAgent := c.Get("User-Agent")
	if userAgent != "" && c.Path() == "/" && isMobile(userAgent) && !fiber.Query[bool](c, "desktop", false) {
		return "/mobile"
	}

	return ""
}

func isMobile(userAgent string) bool {
	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPad", "iPod", "BlackBerry",
		"Windows Phone", "webOS", "Opera Mini", "IEMobile",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgentLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func isMaintenanceMode() bool {
	return false
}

// GetFaviconHTMLTags 返回HTML中使用的favicon标签
func GetFaviconHTMLTags() string {
	return `
	<!-- Favicon -->
	<link rel="icon" type="image/x-icon" href="/favicon.ico">
	<link rel="icon" type="image/svg+xml" href="/favicon.svg">
	<link rel="apple-touch-icon" href="/favicon.svg">
	`
}
