package middleware

import (
	"strings"

	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// SetupCORS 挂载跨域中间件。
// 作用：控制哪些前端源、方法和请求头可以跨域访问服务。
// 场景：前后端分离、浏览器直连 API、WebView 调用。
// 使用方式：全局注册，通过配置项调整白名单。
func SetupCORS(app *fiber.App, cfg *config.Config) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	app.Use(cors.New(corsConfig(cfg)))
}

func corsConfig(cfg *config.Config) cors.Config {
	allowedOrigins := splitList(cfg.Security.CORS.AllowedOrigins, []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"https://localhost:3000",
	})
	allowedMethods := splitList(cfg.Security.CORS.AllowedMethods, []string{
		fiber.MethodGet,
		fiber.MethodPost,
		fiber.MethodHead,
		fiber.MethodPut,
		fiber.MethodDelete,
		fiber.MethodPatch,
		fiber.MethodOptions,
	})
	allowedHeaders := splitList(cfg.Security.CORS.AllowedHeaders, []string{
		fiber.HeaderOrigin,
		fiber.HeaderContentType,
		fiber.HeaderAccept,
		fiber.HeaderAuthorization,
		fiber.HeaderCacheControl,
		fiber.HeaderXRequestedWith,
		requestIDHeader,
		"X-Idempotency-Key",
		"X-API-Key",
		fiber.HeaderIfNoneMatch,
	})

	allowCredentials := !containsWildcard(allowedOrigins)

	return cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     allowedMethods,
		AllowHeaders:     allowedHeaders,
		ExposeHeaders:    []string{fiber.HeaderContentLength, requestIDHeader, "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", fiber.HeaderRetryAfter},
		AllowCredentials: allowCredentials,
		MaxAge:           86400,
	}
}

func splitList(raw string, defaults []string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	if len(values) == 0 {
		return defaults
	}
	return values
}

func containsWildcard(values []string) bool {
	for _, value := range values {
		if value == "*" {
			return true
		}
	}
	return false
}
