package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// SetupRecover 挂载 panic 恢复中间件。
// 作用：捕获 panic 并防止请求链路崩溃。
// 场景：所有对外 API 请求都应启用，兜底未处理异常。
// 使用方式：全局注册，建议尽早挂载。
func SetupRecover(app *fiber.App) {
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
}
