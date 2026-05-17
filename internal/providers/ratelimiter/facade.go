package ratelimiter

import (
	"fiber-starter/configs"
	"fiber-starter/internal/providers/ratelimiter/Contracts"
	"fiber-starter/internal/support/appctx"

	"github.com/gofiber/fiber/v3"
)

// resolveLimiter returns the rate limiter instance from the container.
func resolveLimiter() Contracts.Limiter {
	if app := appctx.App(); app != nil {
		return app.RateLimiterService()
	}
	return nil
}

// Strategy returns the configuration for a named strategy
func Strategy(name string) (configs.RateLimitConfig, bool) {
	if l := resolveLimiter(); l != nil {
		return l.Strategy(name)
	}
	return configs.RateLimitConfig{}, false
}

// Middleware returns a fiber middleware for a named strategy
func Middleware(name string) fiber.Handler {
	if l := resolveLimiter(); l != nil {
		return l.Middleware(name)
	}
	// Return a no-op middleware if limiter is not available
	return func(c fiber.Ctx) error {
		return c.Next()
	}
}
