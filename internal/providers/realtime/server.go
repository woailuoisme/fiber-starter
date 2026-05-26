package realtime

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"lfiber/configs"

	"github.com/gofiber/contrib/v3/socketio"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

var (
	activeRealtimeManager atomic.Pointer[Manager]
	socketioOnce          sync.Once
)

// Server wires the realtime manager into the Fiber/socketio transport.
type Server struct {
	manager *Manager
	handler fiber.Handler
	once    sync.Once
}

func NewServer(manager *Manager) *Server {
	return &Server{manager: manager}
}

func setActiveManager(manager *Manager) {
	activeRealtimeManager.Store(manager)
	registerSocketListeners()
}

func clearActiveManager(manager *Manager) {
	if activeRealtimeManager.Load() == manager {
		activeRealtimeManager.Store(nil)
	}
}

func registerSocketListeners() {
	socketioOnce.Do(func() {
		socketio.On(socketio.EventMessage, func(payload *socketio.EventPayload) {
			if manager := activeRealtimeManager.Load(); manager != nil {
				manager.handleMessage(payload)
			}
		})
		socketio.On(socketio.EventDisconnect, func(payload *socketio.EventPayload) {
			if manager := activeRealtimeManager.Load(); manager != nil {
				manager.handleDisconnect(payload)
			}
		})
		socketio.On(socketio.EventPong, func(payload *socketio.EventPayload) {
			if manager := activeRealtimeManager.Load(); manager != nil {
				manager.handlePong(payload)
			}
		})
		socketio.On(socketio.EventError, func(payload *socketio.EventPayload) {
			if manager := activeRealtimeManager.Load(); manager != nil && payload != nil && payload.Error != nil {
				manager.logWarn("realtime_socket_error")
			}
		})
	})
}

func (s *Server) Handler() fiber.Handler {
	s.once.Do(func() {
		if s.manager != nil {
			setActiveManager(s.manager)
		}

		s.handler = socketio.New(func(kws *socketio.Websocket) {
			if s.manager == nil {
				return
			}
			s.manager.handleConnect(kws)
		}, websocket.Config{
			Next:             nil,
			HandshakeTimeout: 5 * time.Second,
			Origins:          websocketOrigins(s.manager),
			AllowEmptyOrigin: true,
			ReadBufferSize:   4096,
			WriteBufferSize:  4096,
		})
	})
	return s.handler
}

func websocketOrigins(manager *Manager) []string {
	if manager == nil || manager.cfg == nil {
		return []string{"*"}
	}

	allowed := strings.TrimSpace(manager.cfg.Security.CORS.AllowedOrigins)
	if allowed == "" {
		return []string{"*"}
	}

	parts := strings.Split(allowed, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			origins = append(origins, part)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

func (s *Server) Config() *configs.Config {
	if s == nil || s.manager == nil {
		return nil
	}
	return s.manager.cfg
}
