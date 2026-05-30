package realtime

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

// User 抽象业务层用户实体，仅包含实时通信需要的最小集
type User struct {
	ID   string         `json:"id"`
	Info map[string]any `json:"info,omitempty"`
}

// Logger 接口解耦了具体的日志库，允许从外部注入任意结构化日志引擎
type Logger interface {
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// ChannelAuthorization 定义了私有/Presence 频道授权回调。
type ChannelAuthorization func(ctx context.Context, user User, channel string, params map[string]string) error

// Manager 定义了实时通信主控的通用接口。
// 仅保留 WebSocket、SSE、订阅鉴权以及主动推送的核心 API。
// 移除旧 Pusher 模拟的 APIHandler、用于内联握手管理的 Close 以及 Handler 别名，因为 Centrifugo 作为独立进程完全接管了连接生命周期与兼容端点。
type Manager interface {
	// WebSocketHandler 返回获取 WebSocket 连接凭证的 Fiber Handler (返回 url 和 token)
	WebSocketHandler() fiber.Handler

	// SSEHandler 返回获取 SSE 连接凭证的 Fiber Handler
	SSEHandler() fiber.Handler

	// AuthHandler 返回处理私有/Presence频道订阅授权的 Fiber Handler
	AuthHandler() fiber.Handler

	// Dispatch 往 Centrifugo 推送事件数据
	Dispatch(channel, event string, data any) error

	// GetNodeID 获取当前节点在集群中的唯一标识
	GetNodeID() string
}
