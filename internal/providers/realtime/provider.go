package realtime

import (
	"errors"
	"fmt"
	"strings"

	"lfiber/configs"
	"lfiber/internal/features/user"
	realtimeContracts "lfiber/internal/providers/realtime/contracts"
	helpers "lfiber/internal/support"
	"lfiber/pkg/realtime"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// zapLoggerBridge 把应用 zap logger 适配为 realtime.Logger 接口
type zapLoggerBridge struct{}

func (zapLoggerBridge) Info(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			zapFields = append(zapFields, zap.Any(fmt.Sprint(fields[i]), fields[i+1]))
		}
	}
	helpers.Info(msg, zapFields...)
}

func (zapLoggerBridge) Warn(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			zapFields = append(zapFields, zap.Any(fmt.Sprint(fields[i]), fields[i+1]))
		}
	}
	helpers.Warn(msg, zapFields...)
}

func (zapLoggerBridge) Error(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			zapFields = append(zapFields, zap.Any(fmt.Sprint(fields[i]), fields[i+1]))
		}
	}
	helpers.Error(msg, zapFields...)
}

// RegisterRealtime 将全局配置、默认连接及鉴权适配，实例化并装载到应用容器中
func RegisterRealtime(cfg *configs.Config) (realtimeContracts.Manager, error) {
	if !cfg.WebSocket.Enabled {
		return nil, nil
	}

	pkgCfg := &realtime.Config{
		Enabled:           cfg.WebSocket.Enabled,
		AppID:             cfg.WebSocket.AppID,
		AppKey:            cfg.WebSocket.AppKey,
		AppSecret:         cfg.WebSocket.AppSecret,
		Path:              cfg.WebSocket.Path,
		SSEPath:           cfg.WebSocket.SSEPath,
		BusMode:           cfg.WebSocket.BusMode,
		RedisPrefix:       cfg.WebSocket.RedisPrefix,
		WriteQueueSize:    cfg.WebSocket.WriteQueueSize,
		MaxMessageSize:    cfg.WebSocket.MaxMessageSize,
		HeartbeatInterval: cfg.WebSocket.HeartbeatInterval,
		PresenceTTL:       cfg.WebSocket.PresenceTTL,
	}

	// 共享/复用全局 Redis 客户端连接
	if strings.EqualFold(strings.TrimSpace(cfg.WebSocket.BusMode), "redis") {
		addr := strings.TrimSpace(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port))
		pkgCfg.RedisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}

	logger := zapLoggerBridge{}
	manager := realtime.NewManager(pkgCfg, logger)

	// 注入 WebSocket 连接建立时的用户解析规则 (从 Locals 读取已认证的业务用户)
	manager.SetUserResolver(func(conn *websocket.Conn) (realtime.User, error) {
		u, ok := conn.Locals("user").(*user.User)
		if !ok || u == nil {
			return realtime.User{}, errors.New("unauthorized")
		}
		return realtime.User{
			ID: fmt.Sprintf("%d", u.ID),
			Info: map[string]any{
				"id":    u.ID,
				"email": u.Email,
				"name":  u.Name,
			},
		}, nil
	})

	// 注入通道订阅鉴权 Hook 闭包
	manager.SetOnSubscribe(func(sessionID string, channel string, u realtime.User) error {
		if u.ID == "" {
			return errors.New("authentication required for private channel")
		}
		// 预留位置：后续可通过 Redis 或数据库查询，细粒度判定用户对 channelName 的拥有权限
		return nil
	})

	// 注入 HTTP Broadcasting Auth 广播路由的鉴权提取逻辑
	manager.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		u, ok := c.Locals("user").(*user.User)
		if !ok || u == nil {
			return realtime.User{}, errors.New("unauthorized")
		}
		return realtime.User{
			ID: fmt.Sprintf("%d", u.ID),
			Info: map[string]any{
				"id":    u.ID,
				"email": u.Email,
				"name":  u.Name,
			},
		}, nil
	})

	return manager, nil
}
