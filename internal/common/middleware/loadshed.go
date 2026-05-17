package middleware

import (
	"time"

	"fiber-starter/configs"
	exceptions "fiber-starter/internal/common/exceptions"

	loadshed "github.com/gofiber/contrib/v3/loadshed"
	"github.com/gofiber/fiber/v3"
)

// SetupLoadShed 挂载负载保护中间件。
// 作用：根据系统负载（如 CPU）自动拒绝部分请求，保护系统不被压垮。
// 场景：突发流量高峰、系统资源告急。
// 使用方式：全局注册，建议放在较靠前的位置。
func SetupLoadShed(app *fiber.App, cfg *configs.Config) {
	if cfg == nil || !cfg.Security.LoadShed.Enabled {
		return
	}

	app.Use(loadshed.New(loadshed.Config{
		Criteria: &loadshed.CPULoadCriteria{
			LowerThreshold: cfg.Security.LoadShed.LowerThreshold,
			UpperThreshold: cfg.Security.LoadShed.UpperThreshold,
			Interval:       10 * time.Second,
			Getter:         &loadshed.DefaultCPUPercentGetter{},
		},
		OnShed: func(c fiber.Ctx) error {
			// 返回 503 Service Unavailable 异常
			return exceptions.NewServiceUnavailableException("Server is under heavy load, please try again later")
		},
	}))
}
