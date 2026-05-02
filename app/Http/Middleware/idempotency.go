package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/idempotency"
)

// IdempotencyMiddleware 返回幂等写入保护中间件。
// 作用：对带幂等键的重复提交返回一致结果，避免重复创建或重复扣款。
// 场景：支付、订单、Webhook、重复点击提交按钮等写操作。
// 使用方式：客户端在请求头中传入 X-Idempotency-Key，服务端按路由挂载。
func IdempotencyMiddleware() fiber.Handler {
	return idempotency.New(idempotency.Config{
		Lifetime: 30 * time.Minute,
	})
}
