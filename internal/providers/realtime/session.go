package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"fiber-starter/internal/features/user"

	"github.com/gofiber/contrib/v3/socketio"
	"go.uber.org/zap"
)

type Session struct {
	manager *Manager
	kws     *socketio.Websocket
	id      string

	mu        sync.RWMutex
	user      *user.User
	channels  map[string]struct{}
	presence  map[string]PresenceMember
	recv      chan []byte
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once

	heartbeatInterval time.Duration
	lastPong          time.Time
}

func newSession(manager *Manager, kws *socketio.Websocket) *Session {
	heartbeat := time.Duration(manager.heartbeatIntervalSeconds()) * time.Second
	if heartbeat <= 0 {
		heartbeat = 30 * time.Second
	}

	queueSize := manager.writeQueueSize()
	if queueSize <= 0 {
		queueSize = 128
	}

	session := &Session{
		manager:           manager,
		kws:               kws,
		id:                kws.GetUUID(),
		channels:          make(map[string]struct{}),
		presence:          make(map[string]PresenceMember),
		recv:              make(chan []byte, queueSize),
		send:              make(chan []byte, queueSize),
		done:              make(chan struct{}),
		heartbeatInterval: heartbeat,
		lastPong:          time.Now(),
	}

	if user, ok := kws.Locals("user").(*user.User); ok && user != nil {
		session.user = user
	}

	return session
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) User() *user.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.user
}

func (s *Session) SetUser(user *user.User) {
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

func (s *Session) addChannel(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[channel] = struct{}{}
}

func (s *Session) removeChannel(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.channels, channel)
	delete(s.presence, channel)
}

func (s *Session) hasChannel(channel string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.channels[channel]
	return ok
}

func (s *Session) setPresenceMember(channel string, member PresenceMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presence[channel] = member
}

func (s *Session) presenceMember(channel string) (PresenceMember, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.presence[channel]
	return member, ok
}

func (s *Session) Start() {
	go s.readPump()
	go s.writePump()
	go s.heartbeatLoop()
}

func (s *Session) Inbound(data []byte) bool {
	select {
	case <-s.done:
		return false
	default:
	}
	select {
	case s.recv <- data:
		return true
	default:
		s.manager.logWarn("realtime_recv_queue_full", zap.String("socket_id", s.id))
		s.manager.removeSession(s.id)
		return false
	}
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
		s.manager.logWarn("realtime_send_queue_full", zap.String("socket_id", s.id))
		s.manager.removeSession(s.id)
		return false
	}
}

func (s *Session) TouchPong() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPong = time.Now()
}

func (s *Session) shutdown() {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.kws != nil && s.kws.IsAlive() {
			s.kws.Close()
		}
	})
}

func (s *Session) readPump() {
	for {
		select {
		case data, ok := <-s.recv:
			if !ok {
				return
			}
			s.handleInbound(data)
		case <-s.done:
			return
		}
	}
}

func (s *Session) writePump() {
	for {
		select {
		case frame, ok := <-s.send:
			if !ok {
				return
			}
			if s.kws == nil || !s.kws.IsAlive() {
				s.manager.removeSession(s.id)
				return
			}
			s.kws.Emit(frame)
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
				s.manager.logWarn("realtime_heartbeat_timeout", zap.String("socket_id", s.id))
				s.manager.removeSession(s.id)
				return
			}
			_ = s.SendMessage(Message{
				Event: EventPing,
			})
		case <-s.done:
			return
		}
	}
}

func (s *Session) handleInbound(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		s.sendError("invalid realtime payload")
		return
	}

	switch msg.Event {
	case EventSubscribe:
		s.handleSubscribe(msg)
	case EventUnsubscribe:
		s.handleUnsubscribe(msg)
	case EventPong:
		s.TouchPong()
	case EventPing:
		_ = s.SendMessage(Message{Event: EventPong})
	default:
		s.handleBroadcast(msg)
	}
}

func (s *Session) handleSubscribe(msg Message) {
	payload, err := decodeSubscribePayload(msg.Data)
	if err != nil {
		s.sendError(err.Error())
		return
	}

	if err := s.manager.subscribeSession(s, payload); err != nil {
		s.sendError(err.Error())
		return
	}
}

func (s *Session) handleUnsubscribe(msg Message) {
	payload, err := decodeSubscribePayload(msg.Data)
	if err != nil {
		// fallback to channel from top-level field for simple clients
		payload.Channel = msg.Channel
	}
	if payload.Channel == "" {
		s.sendError("missing channel")
		return
	}
	s.manager.unsubscribeSession(s, payload.Channel)
}

func (s *Session) handleBroadcast(msg Message) {
	if msg.Channel == "" || msg.Event == "" || len(msg.Data) == 0 {
		return
	}
	s.manager.publishEnvelope(Envelope{
		NodeID:         s.manager.nodeID,
		Event:          msg.Event,
		Channel:        msg.Channel,
		Data:           msg.Data,
		OriginSocketID: s.id,
	})
}

func (s *Session) sendError(message string) {
	_ = s.SendMessage(Message{
		Event: EventError,
		Data:  encodeJSON(ErrorPayload{Message: message}),
	})
}

func (s *Session) SendMessage(msg Message) error {
	frame, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if ok := s.EnqueueFrame(frame); !ok {
		return nil
	}
	return nil
}
