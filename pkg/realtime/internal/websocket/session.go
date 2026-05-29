package websocket

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"lfiber/pkg/realtime/internal/pusher"

	fiberws "github.com/gofiber/contrib/v3/websocket"
	"github.com/google/uuid"
)

// Session 代表一个连接到当前节点的活跃 WebSocket 连接会话。
type Session struct {
	controller Controller
	conn       *fiberws.Conn
	id         string

	mu        sync.RWMutex
	writeMu   sync.Mutex
	user      pusher.User
	channels  map[string]struct{}
	presence  map[string]pusher.PresenceMember
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once

	heartbeatInterval time.Duration
	lastPong          time.Time
}

func NewSession(controller Controller, conn *fiberws.Conn) *Session {
	heartbeat := time.Duration(controller.HeartbeatIntervalSeconds()) * time.Second
	if heartbeat <= 0 {
		heartbeat = 30 * time.Second
	}

	queueSize := controller.WriteQueueSize()
	if queueSize <= 0 {
		queueSize = 128
	}

	session := &Session{
		controller:        controller,
		conn:              conn,
		id:                uuid.NewString(),
		channels:          make(map[string]struct{}),
		presence:          make(map[string]pusher.PresenceMember),
		send:              make(chan []byte, queueSize),
		done:              make(chan struct{}),
		heartbeatInterval: heartbeat,
		lastPong:          time.Now(),
	}

	if u, err := controller.ResolveUser(conn); err == nil {
		session.user = u
	}

	return session
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) User() pusher.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.user
}

func (s *Session) SetUser(user pusher.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.user = user
}

func (s *Session) ChannelNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channels := make([]string, 0, len(s.channels))
	for channel := range s.channels {
		channels = append(channels, channel)
	}
	return channels
}

func (s *Session) AddChannel(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[channel] = struct{}{}
}

func (s *Session) RemoveChannel(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.channels, channel)
	delete(s.presence, channel)
}

func (s *Session) HasChannel(channel string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.channels[channel]
	return ok
}

func (s *Session) SetPresenceMember(channel string, member pusher.PresenceMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presence[channel] = member
}

func (s *Session) PresenceMember(channel string) (pusher.PresenceMember, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.presence[channel]
	return member, ok
}

func (s *Session) Start() {
	go s.writePump()
	go s.heartbeatLoop()
	s.readLoop()
}

func (s *Session) EnqueueFrame(frame []byte) bool {
	select {
	case <-s.done:
		return false
	default:
	}
	select {
	case s.send <- frame:
		return true
	default:
		s.controller.Warn("realtime_send_queue_full", "socket_id", s.id)
		s.controller.RemoveWebSocketSession(s.id)
		return false
	}
}

func (s *Session) TouchPong() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPong = time.Now()
}

func (s *Session) Shutdown() {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.conn != nil {
			_ = s.conn.Close()
		}
	})
}

func (s *Session) readLoop() {
	defer s.controller.RemoveWebSocketSession(s.id)
	for {
		messageType, data, err := s.conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != fiberws.TextMessage {
			continue
		}
		s.handleInbound(data)
	}
}

func (s *Session) writePump() {
	for {
		select {
		case frame, ok := <-s.send:
			if !ok {
				return
			}
			s.writeMu.Lock()
			err := s.conn.WriteMessage(fiberws.TextMessage, frame)
			s.writeMu.Unlock()
			if err != nil {
				s.controller.RemoveWebSocketSession(s.id)
				return
			}
		case <-s.done:
			return
		}
	}
}

func (s *Session) heartbeatLoop() {
	ticker := time.NewTicker(s.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			lastPong := s.lastPong
			s.mu.RUnlock()

			if time.Since(lastPong) > 2*s.heartbeatInterval {
				s.controller.Warn("realtime_heartbeat_timeout", "socket_id", s.id)
				s.controller.RemoveWebSocketSession(s.id)
				return
			}
			_ = s.SendMessage(pusher.Message{Event: pusher.EventPing})
		case <-s.done:
			return
		}
	}
}

func (s *Session) handleInbound(data []byte) {
	var msg pusher.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		s.sendError("invalid realtime payload")
		return
	}

	switch msg.Event {
	case pusher.EventSubscribe:
		s.handleSubscribe(msg)
	case pusher.EventUnsubscribe:
		s.handleUnsubscribe(msg)
	case pusher.EventPong:
		s.TouchPong()
	case pusher.EventPing:
		_ = s.SendMessage(pusher.Message{Event: pusher.EventPong})
	default:
		s.handleBroadcast(msg)
	}
}

func (s *Session) handleSubscribe(msg pusher.Message) {
	payload, err := pusher.DecodeSubscribePayload(msg.Data)
	if err != nil {
		s.sendError(err.Error())
		return
	}

	if err := s.controller.SubscribeWebSocketSession(s, payload); err != nil {
		s.sendError(err.Error())
		return
	}
}

func (s *Session) handleUnsubscribe(msg pusher.Message) {
	payload, err := pusher.DecodeSubscribePayload(msg.Data)
	if err != nil {
		payload.Channel = msg.Channel
	}
	if payload.Channel == "" {
		s.sendError("missing channel")
		return
	}
	s.controller.UnsubscribeWebSocketSession(s, payload.Channel)
}

func (s *Session) handleBroadcast(msg pusher.Message) {
	if !s.controller.ClientEventsEnabled() {
		s.sendError("client events are disabled")
		return
	}
	if msg.Channel == "" || msg.Event == "" || len(msg.Data) == 0 {
		return
	}
	if !strings.HasPrefix(msg.Event, "client-") || !s.HasChannel(msg.Channel) {
		s.sendError("client event is not allowed")
		return
	}
	s.controller.PublishClientEvent(pusher.Envelope{
		Event:          msg.Event,
		Channel:        msg.Channel,
		Data:           msg.Data,
		OriginSocketID: s.id,
	})
}

func (s *Session) sendError(message string) {
	_ = s.SendMessage(pusher.Message{
		Event: pusher.EventError,
		Data:  pusher.EncodeData(pusher.ErrorPayload{Message: message}),
	})
}

func (s *Session) SendMessage(msg pusher.Message) error {
	frame, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if ok := s.EnqueueFrame(frame); !ok {
		return nil
	}
	return nil
}
