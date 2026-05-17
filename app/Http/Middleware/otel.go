package middleware

import (
	"fiber-starter/configs"

	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
)

// SetupOTEL 挂载 OpenTelemetry 追踪中间件。
// 作用：自动为每个 HTTP 请求创建 Span，并注入追踪上下文。
// 建议：作为最外层或非常靠前的中间件。
func SetupOTEL(app *fiber.App, cfg *configs.Config) {
	if cfg == nil || !cfg.OTEL.Enabled {
		return
	}

	app.Use(fiberotel.Middleware(
		fiberotel.WithSpanNameFormatter(func(c fiber.Ctx) string {
			return c.Method() + " " + c.Route().Path
		}),
	))
}
