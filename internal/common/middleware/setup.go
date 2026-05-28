package middleware

import (
	"lfiber/configs"

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
	// 1. 访问日志：Logger（最外层，基于最终响应状态码记录，每个请求仅记录一次）
	SetupLogger(app)
	// 2. 标识注入：RequestID（确保后续短路、Recover、CORS 都能拿到 ID）
	SetupRequestID(app)
	// 3. 链路追踪：OTEL（在短路中间件前挂载，确保业务请求可观测）
	SetupOTEL(app, cfg)
	// 4. 快速短路：Favicon（对于图标请求，没必要跑后面的复杂逻辑）
	SetupFavicon(app)
	SetupLoadShed(app, cfg)
	// 5. 恐慌捕获：Recover（放在 Logger 内层，panic 转换为 500 后由 Logger 记录）
	SetupRecover(app)
	// 6. 跨域处理：CORS
	SetupCORS(app, cfg)
	// 7. 安全与优化：Helmet & ETag
	SetupHelmet(app)
	SetupETag(app)
}
