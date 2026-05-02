package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

// SetupHelmet 挂载安全响应头中间件。
// 作用：补充常见安全响应头，降低浏览器侧攻击面。
// 场景：公开 API、管理后台、需要安全默认值的 Web 服务。
// 使用方式：全局注册，通常不需要路由级重复设置。
func SetupHelmet(app *fiber.App) {
	app.Use(helmet.New())
}
