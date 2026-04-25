package controllers

import (
	"context"
	"fmt"
	"time"

	"fiber-starter/internal/config"
	database "fiber-starter/internal/db"
	"fiber-starter/internal/platform/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// HealthController 提供健康检查与就绪检查接口
type HealthController struct {
	cfg  *config.Config
	conn *database.Connection
}

// NewHealthController 创建健康检查控制器
func NewHealthController(cfg *config.Config, conn *database.Connection, _ helpers.CacheService) *HealthController {
	return &HealthController{cfg: cfg, conn: conn}
}

// Health 返回基础健康状态
func (h *HealthController) Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Ready 返回依赖就绪状态
func (h *HealthController) Ready(c fiber.Ctx) error {
	checks := fiber.Map{
		"database": h.checkDatabase(),
		"redis":    h.checkRedis(),
	}

	status := fiber.StatusOK
	responseStatus := "ok"
	if checks["database"].(fiber.Map)["status"] != "ok" || checks["redis"].(fiber.Map)["status"] == "fail" {
		status = fiber.StatusServiceUnavailable
		responseStatus = "fail"
	}

	return c.Status(status).JSON(fiber.Map{
		"status": responseStatus,
		"checks": checks,
	})
}

func (h *HealthController) checkDatabase() fiber.Map {
	db, err := h.conn.GetDB()
	if err != nil {
		return fiber.Map{"status": "fail", "error": err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fiber.Map{"status": "fail", "error": err.Error()}
	}

	return fiber.Map{"status": "ok"}
}

func (h *HealthController) checkRedis() fiber.Map {
	if h.cfg.Cache.Driver != "redis" && h.cfg.Queue.Concurrency <= 0 {
		return fiber.Map{"status": "skip"}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", h.cfg.Redis.Host, h.cfg.Redis.Port),
		Password:     h.cfg.Redis.Password,
		DB:           h.cfg.Redis.DB,
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})
	defer func() { _ = rdb.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fiber.Map{"status": "fail", "error": err.Error()}
	}

	return fiber.Map{"status": "ok"}
}
