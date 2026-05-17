package controllers

import (
	providers "fiber-starter/app/Providers"
	health "fiber-starter/app/Support/Health"
	"fiber-starter/configs"

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
//
//	@Summary		基础健康检查
//	@Description	检查应用服务是否存活，用于 Kubernetes Liveness Probe。
//	@Tags			系统监控
//	@Produce		plain
//	@Success		200	{string}	string	"ok"
//	@Router			/health [get]
func (h *HealthController) Health(c fiber.Ctx) error {
	return c.SendString("ok")
}

// Ready 返回依赖就绪状态
//
//	@Summary		服务就绪检查
//	@Description	检查数据库、Redis 等核心依赖服务是否连接正常，用于 Kubernetes Readiness Probe。
//	@Tags			系统监控
//	@Produce		json
//	@Success		200	{object}	map[string]Health.Status	"ok"
//	@Router			/ready [get]
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
