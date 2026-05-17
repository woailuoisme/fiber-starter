package middleware

import (
	"fiber-starter/internal/providers/ratelimiter/Contracts"

	"github.com/gofiber/fiber/v3"
)

// Throttle returns a rate limiting middleware for the specified strategy
func Throttle(limiter Contracts.Limiter, strategy string) fiber.Handler {
	return limiter.Middleware(strategy)
}
