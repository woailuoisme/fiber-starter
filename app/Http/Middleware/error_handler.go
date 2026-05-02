package middleware

import (
	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
)

// ErrorHandler 是全局 HTTP 错误入口。
// 作用：拦截路由链路中的错误，并统一交给 HTTP 错误映射层处理。
// 场景：必须挂在 app.Use(...) 的最外层，保证控制器和下游中间件抛出的错误都能被捕获。
func ErrorHandler(c fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}
	return helpers.HandleHTTPError(c, err)
}
