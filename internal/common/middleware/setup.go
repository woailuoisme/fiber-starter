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
	// 0. 基础修正：Host 头补全（防止空 Host 导致 405 路由匹配失败，必须最外层）
	SetupHostHeader(app)
	// 1. 访问日志：Logger（次外层，基于最终响应状态码记录，每个请求仅记录一次）
	SetupLogger(app)
	// 2. 快速短路：Favicon（对于图标请求，没必要跑后面的复杂逻辑）
	SetupFavicon(app)
	// 3. 标识注入：RequestID（确保后续的 Recover、CORS 都能拿到 ID）
	SetupRequestID(app)
	SetupLoadShed(app, cfg)
	// 4. 恐慌捕获：Recover（放在 Logger 内层，panic 转换为 500 后由 Logger 记录）
	SetupRecover(app)
	// 5. 跨域处理：CORS
	SetupCORS(app, cfg)
	// 6. 安全与优化：Helmet & ETag
	SetupHelmet(app)
	SetupETag(app)
}

// SetupAuthMiddleware 保留认证中间件的统一装配入口。
// 作用：预留全局认证相关扩展点。
// 场景：未来如果需要增加全局认证前置逻辑，可集中放在这里。
// 使用方式：由应用启动时调用；当前版本不额外挂载任何内容。
func SetupAuthMiddleware(_ *fiber.App) {}
