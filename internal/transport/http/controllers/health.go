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
	cfg   *config.Config
	conn  *database.Connection
	cache helpers.CacheService
}

// NewHealthController 创建健康检查控制器
func NewHealthController(cfg *config.Config, conn *database.Connection, cache helpers.CacheService) *HealthController {
	return &HealthController{cfg: cfg, conn: conn, cache: cache}
}

// Health 返回基础健康状态
func (h *HealthController) Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Ready 返回依赖就绪状态
func (h *HealthController) Ready(c fiber.Ctx) error {
	code := fiber.StatusOK
	checks := fiber.Map{}

	db, err := h.conn.GetDB()
	if err != nil {
		code = fiber.StatusServiceUnavailable
		checks["database"] = fiber.Map{"status": "fail", "error": err.Error()}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			code = fiber.StatusServiceUnavailable
			checks["database"] = fiber.Map{"status": "fail", "error": err.Error()}
		} else {
			checks["database"] = fiber.Map{"status": "ok"}
		}
	}

	redisRequired := h.cfg.Cache.Driver == "redis" || h.cfg.Queue.Concurrency > 0
	if redisRequired {
		addr := fmt.Sprintf("%s:%s", h.cfg.Redis.Host, h.cfg.Redis.Port)
		rdb := redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     h.cfg.Redis.Password,
			DB:           h.cfg.Redis.DB,
			DialTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			code = fiber.StatusServiceUnavailable
			checks["redis"] = fiber.Map{"status": "fail", "error": err.Error()}
		} else {
			checks["redis"] = fiber.Map{"status": "ok"}
		}
		_ = rdb.Close()
	} else {
		checks["redis"] = fiber.Map{"status": "skip"}
	}

	if code == fiber.StatusOK {
		return c.Status(code).JSON(fiber.Map{"status": "ok", "checks": checks})
	}
	return c.Status(code).JSON(fiber.Map{"status": "fail", "checks": checks})
}
