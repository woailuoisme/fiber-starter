package middleware

import (
	"strings"

	"fiber-starter/configs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// SetupCORS 挂载跨域中间件。
// 作用：控制哪些前端源、方法和请求头可以跨域访问服务。
// 场景：前后端分离、浏览器直连 API、WebView 调用。
// 使用方式：全局注册，通过配置项调整白名单。
func SetupCORS(app *fiber.App, cfg *configs.Config) {
	app.Use(cors.New(corsConfig(cfg)))
}

func corsConfig(cfg *configs.Config) cors.Config {
	var origins, methods, headers string
	if cfg != nil {
		origins = cfg.Security.CORS.AllowedOrigins
		methods = cfg.Security.CORS.AllowedMethods
		headers = cfg.Security.CORS.AllowedHeaders
	}

	allowedOrigins := splitList(origins, []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"https://localhost:3000",
	})
	allowedMethods := splitList(methods, []string{
		fiber.MethodGet,
		fiber.MethodPost,
		fiber.MethodHead,
		fiber.MethodPut,
		fiber.MethodDelete,
		fiber.MethodPatch,
		fiber.MethodOptions,
	})
	allowedHeaders := splitList(headers, []string{
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
	if raw == "" {
		return defaults
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
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
