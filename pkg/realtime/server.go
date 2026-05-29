package realtime

import (
	"strings"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

// Server 用于将实时通信 Manager 挂载接入 Fiber WebSocket 通道。
type Server struct {
	manager *ManagerImpl
	handler fiber.Handler
}

func NewServer(manager *ManagerImpl) *Server {
	return &Server{manager: manager}
}

func (s *Server) Handler() fiber.Handler {
	if s.handler != nil {
		return s.handler
	}

	s.handler = websocket.New(func(conn *websocket.Conn) {
		if s.manager == nil {
			_ = conn.Close()
			return
		}
		s.manager.handleConnect(conn)
	}, websocket.Config{
		HandshakeTimeout: 5 * time.Second,
		Origins:          websocketOrigins(s.manager),
		AllowEmptyOrigin: true,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
	})
	return s.handler
}

func websocketOrigins(manager *ManagerImpl) []string {
	if manager == nil || manager.cfg == nil {
		return []string{"*"}
	}

	return []string{"*"}
}

func (s *Server) Config() *Config {
	if s == nil || s.manager == nil {
		return nil
	}
	return s.manager.cfg
}

func configuredPath(cfg *Config) string {
	if cfg == nil || strings.TrimSpace(cfg.Path) == "" {
		return "/app/:appKey"
	}
	return strings.TrimSpace(cfg.Path)
}
