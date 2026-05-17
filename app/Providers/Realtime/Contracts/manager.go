package contracts

import (
	"github.com/gofiber/fiber/v3"
)

// Manager defines the interface for the realtime manager.
type Manager interface {
	// Handler returns the websocket handler.
	Handler() fiber.Handler

	// AuthHandler returns the authentication handler.
	AuthHandler() fiber.Handler

	// Dispatch sends an event to a channel.
	Dispatch(channel, event string, data any) error

	// Close shuts down the manager.
	Close() error
}
