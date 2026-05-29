package sse

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	fibersse "github.com/gofiber/fiber/v3/middleware/sse"
)

type Controller interface {
	AppKey() string
	HeartbeatIntervalSeconds() int
	RegisterSSESession(*Session)
	RemoveSSESession(string)
	SubscribeSSESession(fiber.Ctx, *Session, SubscribeRequest) error
	ValidateSSESubscribeRequest(SubscribeRequest) error
	WriteQueueSize() int
	Warn(msg string, fields ...any)
}

type Server struct {
	controller Controller
	handler    fiber.Handler
}

type SubscribeRequest struct {
	SocketID             string
	Channels             []string
	Auth                 string
	ChannelData          string
	Auths                map[string]string
	ChannelDataByChannel map[string]string
}

func NewServer(controller Controller) *Server {
	return &Server{controller: controller}
}

func (s *Server) Handler() fiber.Handler {
	if s.handler != nil {
		return s.handler
	}

	streamHandler := fibersse.New(fibersse.Config{
		DisableHeartbeat: true,
		Handler: func(c fiber.Ctx, stream *fibersse.Stream) error {
			return s.handle(c, stream)
		},
	})

	s.handler = func(c fiber.Ctx) error {
		if appKey := strings.TrimSpace(s.controller.AppKey()); appKey != "" {
			if requestKey := strings.TrimSpace(c.Params("appKey")); requestKey != "" && requestKey != appKey {
				return fiber.NewError(fiber.StatusForbidden, "invalid app key")
			}
		}

		req, err := ParseSubscribeRequest(c)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if err := s.controller.ValidateSSESubscribeRequest(req); err != nil {
			return err
		}
		c.Locals("realtime.sse_request", req)

		return streamHandler(c)
	}
	return s.handler
}

func (s *Server) handle(c fiber.Ctx, stream *fibersse.Stream) error {
	req, _ := c.Locals("realtime.sse_request").(SubscribeRequest)

	session := NewSession(s.controller, req.SocketID)
	s.controller.RegisterSSESession(session)
	if err := s.controller.SubscribeSSESession(c, session, req); err != nil {
		s.controller.RemoveSSESession(session.ID())
		return err
	}
	defer s.controller.RemoveSSESession(session.ID())

	if err := stream.Comment("connected"); err != nil {
		return err
	}

	heartbeat := time.Duration(s.controller.HeartbeatIntervalSeconds()) * time.Second
	if heartbeat <= 0 {
		heartbeat = 30 * time.Second
	}
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case frame := <-session.Frames():
			if err := stream.Event(fibersse.Event{
				ID:   frame.ID,
				Name: frame.Name,
				Data: frame.Data,
			}); err != nil {
				return err
			}
		case <-ticker.C:
			if err := stream.Comment("ping"); err != nil {
				return err
			}
		case <-session.Done():
			return nil
		case <-stream.Context().Done():
			return nil
		}
	}
}

func ParseSubscribeRequest(c fiber.Ctx) (SubscribeRequest, error) {
	req := SubscribeRequest{
		SocketID:    strings.TrimSpace(c.Query("socket_id")),
		Auth:        strings.TrimSpace(c.Query("auth")),
		ChannelData: strings.TrimSpace(c.Query("channel_data")),
	}

	for _, item := range strings.Split(c.Query("channels"), ",") {
		channel := strings.TrimSpace(item)
		if channel != "" {
			req.Channels = append(req.Channels, channel)
		}
	}
	if len(req.Channels) == 0 {
		if channel := strings.TrimSpace(c.Query("channel")); channel != "" {
			req.Channels = append(req.Channels, channel)
		}
	}
	if len(req.Channels) == 0 {
		return SubscribeRequest{}, errors.New("missing channels")
	}

	if raw := strings.TrimSpace(c.Query("auths")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &req.Auths); err != nil {
			return SubscribeRequest{}, errors.New("invalid auths")
		}
	}
	if raw := strings.TrimSpace(c.Query("channel_data_by_channel")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &req.ChannelDataByChannel); err != nil {
			return SubscribeRequest{}, errors.New("invalid channel_data_by_channel")
		}
	}

	return req, nil
}

func (r SubscribeRequest) AuthFor(channel string) string {
	if len(r.Auths) > 0 {
		if auth := strings.TrimSpace(r.Auths[channel]); auth != "" {
			return auth
		}
	}
	return r.Auth
}

func (r SubscribeRequest) ChannelDataFor(channel string) string {
	if len(r.ChannelDataByChannel) > 0 {
		if data := strings.TrimSpace(r.ChannelDataByChannel[channel]); data != "" {
			return data
		}
	}
	return r.ChannelData
}
