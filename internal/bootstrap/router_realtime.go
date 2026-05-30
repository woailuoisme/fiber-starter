package bootstrap

import (
	"strings"

	"lfiber/internal/providers"

	"github.com/gofiber/fiber/v3"
)

// registerRealtimeRoutes 注册系统实时消息的基础端点。
// 物理隔离在独立文件内，使主 router.go 仅聚焦于底座核心及诊断端点。
func registerRealtimeRoutes(app *fiber.App, jwtProtected fiber.Handler) {
	rt := providers.App()
	if rt == nil || rt.Realtime == nil || rt.Config == nil {
		return
	}

	if path := strings.TrimSpace(rt.Config.WebSocket.AuthPath); path != "" {
		app.Post(path, jwtProtected, rt.Realtime.AuthHandler())
	}

	if path := strings.TrimSpace(rt.Config.WebSocket.Path); path != "" {
		app.Get(path, rt.Realtime.WebSocketHandler())
	}
	if strings.TrimSpace(rt.Config.WebSocket.Path) != "/app/:appKey" {
		app.Get("/app/:appKey", rt.Realtime.WebSocketHandler())
	}
	if path := strings.TrimSpace(rt.Config.WebSocket.SSEPath); path != "" {
		app.Get(path, rt.Realtime.SSEHandler())
	}
	if strings.TrimSpace(rt.Config.WebSocket.SSEPath) != "/sse/app/:appKey" {
		app.Get("/sse/app/:appKey", rt.Realtime.SSEHandler())
	}
}
