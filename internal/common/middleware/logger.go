package middleware

import (
	"time"

	logging "fiber-starter/internal/providers/logging"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupLogger 挂载请求日志中间件。
// 作用：在整个中间件链执行完毕后，基于最终响应状态码记录一条结构化访问日志。
// 规则：每个请求只记录一次。不调用 ErrorHandler，不干扰错误处理链路。
// 场景：访问审计、错误回溯、慢请求分析。
func SetupLogger(app *fiber.App) {
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()

		// 执行后续所有中间件和路由处理器；error 原样返回给 Fiber，
		// Fiber 的 app.ErrorHandler 负责统一发送错误响应。
		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		fields := []zap.Field{
			zap.String("request_id", getRequestID(c)),
			zap.String("ip", c.IP()),
			zap.String("method", c.Method()),
			zap.String("url", c.OriginalURL()),
			zap.String("ua", c.Get(fiber.HeaderUserAgent)),
			zap.Duration("latency", latency),
			zap.Int("status", status),
		}

		switch {
		case status >= 500:
			logging.Facade().Error("server_error", fields...)
		case status >= 400:
			logging.Facade().Warn("client_error", fields...)
		default:
			logging.Facade().Info("access", fields...)
		}

		return err
	})
}
