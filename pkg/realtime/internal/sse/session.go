package sse

import (
	"encoding/json"
	"sync"

	"lfiber/pkg/realtime/internal/pusher"

	"github.com/google/uuid"
)

type Envelope struct {
	Channel  string          `json:"channel"`
	Event    string          `json:"event"`
	Data     json.RawMessage `json:"data,omitempty"`
	SocketID string          `json:"socket_id,omitempty"`
}

type Frame struct {
	ID   string
	Name string
	Data json.RawMessage
}

type Session struct {
	controller Controller
	id         string

	mu        sync.RWMutex
	channels  map[string]struct{}
	send      chan Frame
	done      chan struct{}
	closeOnce sync.Once
}

func NewSession(controller Controller, id string) *Session {
	if id == "" {
		id = uuid.NewString()
	}

	queueSize := controller.WriteQueueSize()
	if queueSize <= 0 {
		queueSize = 128
	}

	return &Session{
		controller: controller,
		id:         id,
		channels:   make(map[string]struct{}),
		send:       make(chan Frame, queueSize),
		done:       make(chan struct{}),
	}
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Done() <-chan struct{} {
	return s.done
}

func (s *Session) Frames() <-chan Frame {
	return s.send
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

func (s *Session) Enqueue(frame Frame) bool {
	select {
	case <-s.done:
		return false
	default:
	}
	select {
	case s.send <- frame:
		return true
	default:
		s.controller.Warn("realtime_sse_send_queue_full", "socket_id", s.id)
		s.Shutdown()
		return false
	}
}

func (s *Session) SendEnvelope(env pusher.Envelope) bool {
	frame, err := NewFrame(env)
	if err != nil {
		s.controller.Warn("realtime_sse_frame_encode_failed", "channel", env.Channel, "event", env.Event, "error", err.Error())
		return true
	}
	return s.Enqueue(frame)
}

func (s *Session) Shutdown() {
	s.closeOnce.Do(func() {
		close(s.done)
	})
}

func NewFrame(env pusher.Envelope) (Frame, error) {
	data := env.Data
	if len(data) == 0 {
		data = json.RawMessage(`null`)
	}

	payload := Envelope{
		Channel:  env.Channel,
		Event:    env.Event,
		Data:     data,
		SocketID: env.OriginSocketID,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return Frame{}, err
	}

	return Frame{
		ID:   uuid.NewString(),
		Name: env.Event,
		Data: raw,
	}, nil
}
