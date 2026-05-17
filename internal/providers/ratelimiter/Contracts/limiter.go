package contracts

import (
	"fiber-starter/configs"

	fiber "github.com/gofiber/fiber/v3"
)

// Limiter defines the interface for the rate limiting service
type Limiter interface {
	// Strategy returns the configuration for a named strategy
	Strategy(name string) (configs.RateLimitConfig, bool)

	// Middleware returns a fiber middleware for a named strategy
	Middleware(name string) fiber.Handler
}
