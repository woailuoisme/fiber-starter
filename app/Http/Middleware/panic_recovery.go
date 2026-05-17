package middleware

import (
	"fmt"
	"runtime/debug"

	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// RecoveryMiddleware 捕获 panic 并转成可控错误。
// 作用：兜底异常，避免单个 panic 直接打崩请求流程。
// 场景：如果你需要单独插入恢复逻辑，可手工挂载；默认全局链路已经有 Fiber Recover。
// 使用方式：通常不需要额外挂载，保留为兼容/替代实现。
func RecoveryMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				helpers.LogError("PANIC: "+fmt.Sprint(r), zap.String("stack", string(debug.Stack())))
				_ = helpers.HandleHTTPError(c, fiber.NewError(fiber.StatusInternalServerError, "Internal server error"))
			}
		}()

		return c.Next()
	}
}
