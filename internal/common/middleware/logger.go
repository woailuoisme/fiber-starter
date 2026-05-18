package middleware

import (
	"fmt"
	"strings"
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
		c.Locals("start_time", start)

		err := c.Next()
		if err != nil {
			_ = c.App().ErrorHandler(c, err)
			err = nil
		}

		latency := time.Since(start)
		status := c.Response().StatusCode()

		reqID := getRequestID(c)
		method := c.Method()
		url := c.OriginalURL()

		fields := []zap.Field{
			zap.String("request_id", reqID),
			zap.String("ip", c.IP()),
			zap.String("method", method),
			zap.String("url", url),
			zap.String("latency", latency.String()),
			zap.Int("status", status),
		}

		if status == fiber.StatusMethodNotAllowed {
			var headers []string
			for key, values := range c.GetReqHeaders() {
				for _, val := range values {
					headers = append(headers, fmt.Sprintf("%s: %s", key, val))
				}
			}
			host := c.Hostname()
			if host == "" {
				host = "127.0.0.1"
			}
			scheme := strings.ToLower(c.Protocol())
			if strings.Contains(scheme, "https") {
				scheme = "https"
			} else {
				scheme = "http"
			}
			requestURI := fmt.Sprintf("%s://%s%s", scheme, host, url)

			var matchedRoutes []string
			for _, r := range c.App().GetRoutes() {
				if r.Path == "/" || r.Path == url {
					matchedRoutes = append(matchedRoutes, fmt.Sprintf("[%s] %s", r.Method, r.Path))
				}
			}

			logging.Facade().Warn(
				"405_diagnostic_details",
				zap.String("request_id", reqID),
				zap.String("method", method),
				zap.String("url", url),
				zap.String("headers", strings.Join(headers, " | ")),
				zap.String("uri", requestURI),
				zap.String("matched_routes", strings.Join(matchedRoutes, " | ")),
			)
		}

		switch {
		case status >= fiber.StatusInternalServerError:
			logging.Facade().Error("server_error", fields...)
		case status >= fiber.StatusBadRequest:
			logging.Facade().Warn("client_error", fields...)
		default:
			logging.Facade().Info("access", fields...)
		}

		return err
	})
}
