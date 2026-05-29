package realtime

import (
	"context"

	"lfiber/pkg/realtime/internal/pusher"

	"github.com/gofiber/fiber/v3"
)

// User 抽象业务层用户实体，仅包含实时通信需要的最小集
type User = pusher.User

// Logger 接口解耦了具体的日志库，允许从外部注入任意结构化日志引擎
type Logger interface {
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// ChannelAuthorization 定义了私有/Presence 频道授权回调。
type ChannelAuthorization func(ctx context.Context, user User, channel string, params map[string]string) error

// Manager 定义了实时通信主控的通用接口
type Manager interface {
	// WebSocketHandler 返回对接客户端 WebSocket 连接的 Fiber Handler
	WebSocketHandler() fiber.Handler

	// Handler 返回对接客户端 WebSocket 连接的 Fiber Handler
	Handler() fiber.Handler

	// SSEHandler 返回对接客户端 Server-Sent Events 连接的 Fiber Handler
	SSEHandler() fiber.Handler

	// AuthHandler 返回处理广播验证授权的 Fiber Handler
	AuthHandler() fiber.Handler

	// APIHandler 返回 Pusher-compatible REST API Handler
	APIHandler() fiber.Handler

	// Dispatch 推送事件到指定的频道
	Dispatch(channel, event string, data any) error

	// Close 关闭并释放所有相关连接和通道订阅
	Close() error

	// GetNodeID 获取当前节点在集群中的唯一标识
	GetNodeID() string
}
