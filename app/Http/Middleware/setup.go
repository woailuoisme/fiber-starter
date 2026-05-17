package middleware

import (
	"fiber-starter/configs"

	"github.com/gofiber/fiber/v3"
)

// SetupMiddleware 挂载全局默认中间件。
// 作用：按固定顺序装配请求标识、恢复、日志、安全和错误入口。
// 场景：应用启动阶段统一注册，避免在路由层重复拼装。
// 使用方式：由 bootstrap 调用，一次性挂载到 app 上。
func SetupMiddleware(app *fiber.App, cfg *configs.Config) {
	if cfg == nil {
		cfg = &configs.Config{}
	}
	// 0. 分布式追踪：OpenTelemetry (最外层，确保捕获所有逻辑)
	SetupOTEL(app, cfg)
	// 1. 本地计时器：RequestTimer (确保覆盖后续中间件的耗时)
	SetupRequestTimer(app)
	// 2. 快速短路：Favicon（对于图标请求，没必要跑后面的复杂逻辑）
	SetupFavicon(app)
	// 3. 标识注入：RequestID（确保后续的 Recover、Logger、CORS 都能拿到 ID）
	SetupRequestID(app)
	SetupLoadShed(app, cfg)
	// 4. 恐慌捕获：Recover（放在 Logger 之上，这样发生 Panic 时，Logger 依然能记录到 500 状态）
	SetupRecover(app)
	// 5. 跨域处理：CORS（处理 OPTIONS 预检请求，建议放在 Logger 之前或之后，取决于你是否想记录预检日志）
	SetupCORS(app, cfg)
	// 6. 访问日志：Logger（现在它能拿到完整的 Request ID，也能记录被 Recover 捕获后的错误状态）
	SetupLogger(app)
	// 7. 安全与优化：Helmet & ETag
	SetupHelmet(app)
	SetupETag(app)
}

// SetupAuthMiddleware 保留认证中间件的统一装配入口。
// 作用：预留全局认证相关扩展点。
// 场景：未来如果需要增加全局认证前置逻辑，可集中放在这里。
// 使用方式：由应用启动时调用；当前版本不额外挂载任何内容。
func SetupAuthMiddleware(_ *fiber.App) {}
