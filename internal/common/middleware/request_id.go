package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

const requestIDHeader = fiber.HeaderXRequestID

// SetupRequestID 挂载请求标识中间件。
// 作用：为每个请求生成唯一 ID，便于日志关联和链路追踪。
// 场景：排查线上问题、串联网关/下游日志、对接 APM。
// 使用方式：全局注册即可，客户端可选传入 X-Request-ID 参与链路。
func SetupRequestID(app *fiber.App) {
	app.Use(requestid.New(requestid.Config{
		Header: requestIDHeader,
	}))
}

func getRequestID(c fiber.Ctx) string {
	if v := requestid.FromContext(c); v != "" {
		return v
	}

	if v := c.Get(requestIDHeader); v != "" {
		return v
	}

	if v := c.Get("X-Request-ID"); v != "" {
		return v
	}

	if v := c.Locals(requestIDHeader); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}
