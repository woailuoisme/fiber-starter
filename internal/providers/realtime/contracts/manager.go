package contracts

import (
	"lfiber/pkg/realtime"

	"github.com/gofiber/fiber/v3"
)

// Manager defines the interface for the realtime manager.
type Manager interface {
	// Handler returns the websocket handler.
	Handler() fiber.Handler

	// AuthHandler returns the authentication handler.
	AuthHandler() fiber.Handler

	// APIHandler returns the Pusher-compatible REST API handler.
	APIHandler() fiber.Handler

	// AuthorizeChannel registers a Laravel-style channel authorization callback.
	AuthorizeChannel(pattern string, auth realtime.ChannelAuthorization)

	// Dispatch sends an event to a channel.
	Dispatch(channel, event string, data any) error

	// Close shuts down the manager.
	Close() error
}
