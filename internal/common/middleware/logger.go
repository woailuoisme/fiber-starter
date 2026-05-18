package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	logging "fiber-starter/internal/providers/logging"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupLogger 挂载请求日志中间件。
// 作用：在整个中间件链执行完毕后，基于最终响应状态码记录一条结构化访问日志。
// 规则：每个请求只记录一次。普通处理链错误会先交给 ErrorHandler，再按最终状态记录。
// 场景：访问审计、错误回溯、慢请求分析。
func SetupLogger(app *fiber.App) {
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		c.Locals("start_time", start)

		err := c.Next()
		if err != nil {
			if isSyntheticMethodNotAllowed(c, err) {
				return err
			}
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
			logMethodNotAllowedDiagnostic(c, reqID, method, url)
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

func logMethodNotAllowedDiagnostic(c fiber.Ctx, reqID, method, url string) {
	logging.Facade().Warn(
		"405_diagnostic_details",
		zap.String("request_id", reqID),
		zap.String("method", method),
		zap.String("url", url),
		zap.String("headers", requestHeaders(c)),
		zap.String("uri", requestURI(c, url)),
		zap.String("allow", c.GetRespHeader(fiber.HeaderAllow)),
		zap.String("matched_routes", matchingRouteLabels(c, url)),
	)
}

func requestHeaders(c fiber.Ctx) string {
	headers := make([]string, 0, len(c.GetReqHeaders()))
	for key, values := range c.GetReqHeaders() {
		for _, val := range values {
			headers = append(headers, fmt.Sprintf("%s: %s", key, val))
		}
	}

	return strings.Join(headers, " | ")
}

func requestURI(c fiber.Ctx, url string) string {
	host := c.Hostname()
	if host == "" {
		host = "127.0.0.1"
	}

	return fmt.Sprintf("%s://%s%s", requestScheme(c), host, url)
}

func requestScheme(c fiber.Ctx) string {
	if strings.Contains(strings.ToLower(c.Protocol()), "https") {
		return "https"
	}

	return "http"
}

func matchingRouteLabels(c fiber.Ctx, url string) string {
	labels := make([]string, 0)
	for _, r := range c.App().GetRoutes(true) {
		if r.Path == url {
			labels = append(labels, fmt.Sprintf("[%s] %s", r.Method, r.Path))
		}
	}

	return strings.Join(labels, " | ")
}

func isSyntheticMethodNotAllowed(c fiber.Ctx, err error) bool {
	if !errors.Is(err, fiber.ErrMethodNotAllowed) {
		return false
	}
	if c.Method() != fiber.MethodGet {
		return false
	}
	if c.GetRespHeader(fiber.HeaderAllow) != fiber.MethodHead {
		return false
	}

	url := c.OriginalURL()
	for _, r := range c.App().GetRoutes(true) {
		if r.Method == fiber.MethodGet && r.Path == url {
			return true
		}
	}

	return false
}
