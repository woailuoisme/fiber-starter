package websocket

import (
	"time"

	"lfiber/pkg/realtime/internal/pusher"

	fiberws "github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

type Controller interface {
	AppKey() string
	ClientEventsEnabled() bool
	HeartbeatIntervalSeconds() int
	Info(msg string, fields ...any)
	MaxMessageSize() int
	PublishClientEvent(pusher.Envelope)
	RegisterWebSocketSession(*Session)
	RemoveWebSocketSession(string)
	ResolveUser(*fiberws.Conn) (pusher.User, error)
	SubscribeWebSocketSession(*Session, pusher.SubscribePayload) error
	UnsubscribeWebSocketSession(*Session, string)
	Warn(msg string, fields ...any)
	WriteQueueSize() int
}

// Server 用于将实时通信 Manager 挂载接入 Fiber WebSocket 通道。
type Server struct {
	controller Controller
	handler    fiber.Handler
}

func NewServer(controller Controller) *Server {
	return &Server{controller: controller}
}

func (s *Server) Handler() fiber.Handler {
	if s.handler != nil {
		return s.handler
	}

	s.handler = fiberws.New(func(conn *fiberws.Conn) {
		if s.controller == nil {
			_ = conn.Close()
			return
		}
		s.handleConnect(conn)
	}, fiberws.Config{
		HandshakeTimeout: 5 * time.Second,
		Origins:          []string{"*"},
		AllowEmptyOrigin: true,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
	})
	return s.handler
}

func (s *Server) handleConnect(conn *fiberws.Conn) {
	if conn == nil {
		return
	}

	if appKey := s.controller.AppKey(); appKey != "" {
		if requestKey := conn.Params("appKey"); requestKey != "" && requestKey != appKey {
			_ = conn.WriteMessage(fiberws.CloseMessage, fiberws.FormatCloseMessage(fiberws.ClosePolicyViolation, "invalid app key"))
			_ = conn.Close()
			return
		}
	}

	if maxMessageSize := s.controller.MaxMessageSize(); maxMessageSize > 0 {
		conn.SetReadLimit(int64(maxMessageSize))
	}

	session := NewSession(s.controller, conn)
	s.controller.RegisterWebSocketSession(session)

	_ = session.SendMessage(pusher.Message{
		Event: pusher.EventConnectEstablished,
		Data: pusher.EncodeData(pusher.ConnectionEstablishedData{
			SocketID:        session.ID(),
			ActivityTimeout: s.controller.HeartbeatIntervalSeconds(),
		}),
	})
	s.controller.Info("realtime_connected", "socket_id", session.ID())

	session.Start()
}
