package monitoring

import (
	"fiber-starter/configs"
	"fiber-starter/internal/support/appctx"
	health "fiber-starter/internal/support/health"

	"github.com/gofiber/fiber/v3"
)

// HealthController 提供健康检查与就绪检查接口
type HealthController struct {
	cfg *configs.Config
	app appctx.Application
}

// NewHealthController 创建健康检查控制器
func NewHealthController(cfg *configs.Config) *HealthController {
	return &HealthController{
		cfg: cfg,
		app: appctx.App(),
	}
}

// Health 返回基础健康状态
//
//	@Summary		基础健康检查
//	@Description	获取当前 HTTP 服务的存活状态。
//	@Tags			系统监控
//	@Produce		plain
//	@Success		200	{string}	string	"ok"
//	@Router			/health [get]
func (h *HealthController) Health(c fiber.Ctx) error {
	return c.SendString("ok")
}

// Ready 返回依赖就绪状态
//
//	@Summary		就绪性检查
//	@Description	获取当前服务各项依赖（数据库、缓存等）的健康状态。
//	@Tags			系统监控
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"服务正常"
//	@Failure		503	{object}	map[string]interface{}	"服务不可用"
//	@Router			/ready [get]
func (h *HealthController) Ready(c fiber.Ctx) error {
	agg := health.NewAggregator(h.app)
	results, allHealthy := agg.CheckAll()

	status := fiber.StatusOK
	if !allHealthy {
		status = fiber.StatusServiceUnavailable
	}

	return c.Status(status).JSON(fiber.Map{
		"status":   map[bool]string{true: "ok", false: "fail"}[allHealthy],
		"services": results,
	})
}
