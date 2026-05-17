package middleware

import (
	"fiber-starter/app/Providers/RateLimiter/Contracts"

	"github.com/gofiber/fiber/v3"
)

// Throttle returns a rate limiting middleware for the specified strategy
func Throttle(limiter Contracts.Limiter, strategy string) fiber.Handler {
	return limiter.Middleware(strategy)
}
