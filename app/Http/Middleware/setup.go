package middleware

import (
	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
)

// SetupMiddleware 挂载全局默认中间件。
// 作用：按固定顺序装配请求标识、恢复、日志、安全和错误入口。
// 场景：应用启动阶段统一注册，避免在路由层重复拼装。
// 使用方式：由 bootstrap 调用，一次性挂载到 app 上。
func SetupMiddleware(app *fiber.App, cfg *config.Config) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	SetupFavicon(app)
	SetupRequestID(app)
	SetupRecover(app)
	SetupLogger(app)
	SetupCORS(app, cfg)
	SetupHelmet(app)
	SetupETag(app)
	app.Use(ErrorHandler)
}

// SetupAuthMiddleware 保留认证中间件的统一装配入口。
// 作用：预留全局认证相关扩展点。
// 场景：未来如果需要增加全局认证前置逻辑，可集中放在这里。
// 使用方式：由应用启动时调用；当前版本不额外挂载任何内容。
func SetupAuthMiddleware(_ *fiber.App) {}
