package middleware

import (
	support "fiber-starter/app/Support"

	contribzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SetupLogger 挂载请求日志中间件。
// 作用：使用 zap 记录结构化访问日志。
// 场景：访问审计、错误回溯、慢请求分析。
// 使用方式：全局注册，默认使用 support.Logger。
func SetupLogger(app *fiber.App) {
	app.Use(contribzap.New(contribzap.Config{
		Logger: support.Logger,
		FieldsFunc: func(c fiber.Ctx) []zap.Field {
			return []zap.Field{
				zap.String("request_id", getRequestID(c)),
			}
		},
	}))
}
