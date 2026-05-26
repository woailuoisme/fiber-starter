package middleware

import (
	"lfiber/internal/providers/ratelimiter/contracts"

	"github.com/gofiber/fiber/v3"
)

// Throttle returns a rate limiting middleware for the specified strategy
func Throttle(limiter contracts.Limiter, strategy string) fiber.Handler {
	return limiter.Middleware(strategy)
}
