package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// SetupHostHeader 挂载请求 Host 默认值填充中间件。
// 作用：为缺少 Host 头的请求（如某些健康检查、原始 TCP 探测）自动填充默认 Host (127.0.0.1)。
// 场景：防止 Fiber 路由在空 Host 头下发生 405 Method Not Allowed 错误。
// 使用方式：必须作为全局最外层中间件挂载。
func SetupHostHeader(app *fiber.App) {
	app.Use(func(c fiber.Ctx) error {
		if len(c.Request().Header.Host()) == 0 {
			c.Request().Header.SetHost("127.0.0.1")
			c.Request().URI().SetHost("127.0.0.1")
		}
		return c.Next()
	})
}
