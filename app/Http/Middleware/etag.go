package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/etag"
)

// SetupETag 挂载 ETag 中间件。
// 作用：为响应生成缓存校验标识，减少重复传输。
// 场景：GET 列表、静态内容、版本变化不频繁的响应。
// 使用方式：全局注册，客户端可用 If-None-Match 触发 304。
func SetupETag(app *fiber.App) {
	app.Use(etag.New())
}
