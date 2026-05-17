package monitoring

import (
	"fiber-starter/configs"
	health "fiber-starter/internal/common/support/Health"
	providers "fiber-starter/internal/providers"

	"github.com/gofiber/fiber/v3"
)

// HealthController 提供健康检查与就绪检查接口
type HealthController struct {
	cfg *configs.Config
	rt  *providers.Runtime
}

// NewHealthController 创建健康检查控制器
func NewHealthController(cfg *configs.Config) *HealthController {
	return &HealthController{
		cfg: cfg,
		rt:  providers.App(),
	}
}

// Health 返回基础健康状态
func (h *HealthController) Health(c fiber.Ctx) error {
	return c.SendString("ok")
}

// Ready 返回依赖就绪状态
func (h *HealthController) Ready(c fiber.Ctx) error {
	agg := health.NewAggregator(h.rt)
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
