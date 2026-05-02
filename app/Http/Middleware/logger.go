package middleware

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

// SetupLogger 挂载请求日志中间件。
// 作用：输出结构化访问日志，并带上 request id。
// 场景：访问审计、错误回溯、慢请求分析。
// 使用方式：全局注册，默认输出到标准输出。
func SetupLogger(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"request_id": func(output logger.Buffer, c fiber.Ctx, _ *logger.Data, _ string) (int, error) {
				return output.WriteString(getRequestID(c))
			},
		},
		Stream: os.Stdout,
		Format: "${time} ${ip} ${request_id} ${status} - ${latency} ${method} ${path} ${error}\n",
	}))
}
