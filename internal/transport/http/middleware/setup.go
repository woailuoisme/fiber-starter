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

// setupCoreMiddleware 配置核心中间件
func setupCoreMiddleware(app *fiber.App) {
	if _, err := os.Stat("./public/favicon.ico"); err == nil {
		app.Use(favicon.New(favicon.Config{
			File: "./public/favicon.ico",
			URL:  "/favicon.ico",
		}))
	}

	// 备用SVG favicon - 如果.ico文件不存在
	app.Get("/favicon.svg", func(c fiber.Ctx) error {
		return c.SendFile("./public/favicon.svg")
	})

	// 请求ID中间件 - 为每个请求生成唯一ID
	app.Use(requestid.New())

	// 恢复中间件 - 从panic中恢复
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 日志中间件 - 记录HTTP请求
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

// setupSecurityMiddleware 配置安全中间件
func setupSecurityMiddleware(app *fiber.App) {
	// CORS中间件 - 处理跨域请求
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
	}))

	// 安全头中间件 - 添加安全相关的HTTP头
	app.Use(helmet.New())
}

// setupTimeoutMiddleware 配置请求超时中间件
func setupTimeoutMiddleware(app *fiber.App) {
	// 超时中间件 - 设置请求超时
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
func SetupAuthMiddleware(_ *fiber.App) {
	// 这里可以添加认证相关的全局中间件
	// 例如：API密钥验证、黑名单检查等
}

// SetupTimeoutRedirect 配置超时重定向中间件
func SetupTimeoutRedirect(app *fiber.App) {
	// 自定义重定向中间件
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

// isMobile 检查是否为移动设备
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

// isMaintenanceMode 检查是否为维护模式
func isMaintenanceMode() bool {
	// 这里可以从配置文件或环境变量读取维护模式状态
	// 示例：返回false表示不在维护模式
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
