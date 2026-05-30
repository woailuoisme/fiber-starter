package contracts

import (
	"lfiber/pkg/realtime"

	"github.com/gofiber/fiber/v3"
)

// Manager defines the interface for the realtime manager.
type Manager interface {
	// WebSocketHandler returns the websocket handler.
	WebSocketHandler() fiber.Handler

	// SSEHandler returns the Server-Sent Events handler.
	SSEHandler() fiber.Handler

	// AuthHandler returns the authentication handler.
	AuthHandler() fiber.Handler

	// AuthorizeChannel registers a Laravel-style channel authorization callback.
	AuthorizeChannel(pattern string, auth realtime.ChannelAuthorization)

	// Dispatch sends an event to a channel.
	Dispatch(channel, event string, data any) error
}
