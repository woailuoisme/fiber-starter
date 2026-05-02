package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

// RequestTimerMiddleware 记录请求开始时间。
// 作用：给后续耗时统计或超时诊断提供起始时间。
// 场景：需要在后续处理流程里计算总耗时的调试或监控扩展。
// 使用方式：按需挂载；默认链路未使用。
func RequestTimerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("start_time", time.Now())
		return c.Next()
	}
}
