package middleware

import (
	"time"

	"fiber-starter/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// AuthLimiter 返回用于公开写接口的限流中间件。
// 作用：按时间窗口限制单个来源请求频率，降低刷接口和暴力尝试风险。
// 场景：注册、登录、短信验证码、刷新令牌等公开写接口。
// 使用方式：按路由组挂载，例如 auth 组；最大次数和窗口秒数从配置读取。
func AuthLimiter(cfg *config.Config) fiber.Handler {
	if cfg == nil {
		cfg = &config.Config{}
	}

	max := cfg.Security.RateLimit.Max
	if max <= 0 {
		max = 100
	}

	window := cfg.Security.RateLimit.Window
	if window <= 0 {
		window = 60
	}

	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: time.Duration(window) * time.Second,
	})
}
