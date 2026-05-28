package middleware

import (
	"time"

	"lfiber/configs"
	exceptions "lfiber/internal/common/exceptions"

	"github.com/gofiber/contrib/v3/circuitbreaker"
	"github.com/gofiber/fiber/v3"
)

// CircuitBreakerOption defines the functional option pattern for customizing the circuit breaker.
type CircuitBreakerOption func(*circuitbreaker.Config)

// WithFailureThreshold overrides the failure threshold.
func WithFailureThreshold(threshold int) CircuitBreakerOption {
	return func(cfg *circuitbreaker.Config) {
		cfg.FailureThreshold = threshold
	}
}

// WithTimeout overrides the timeout before moving from Open to Half-Open.
func WithTimeout(d time.Duration) CircuitBreakerOption {
	return func(cfg *circuitbreaker.Config) {
		cfg.Timeout = d
	}
}

// WithSuccessThreshold overrides the success threshold.
func WithSuccessThreshold(threshold int) CircuitBreakerOption {
	return func(cfg *circuitbreaker.Config) {
		cfg.SuccessThreshold = threshold
	}
}

// WithHalfOpenMaxConcurrent overrides the max concurrent requests allowed in half-open state.
func WithHalfOpenMaxConcurrent(max int) CircuitBreakerOption {
	return func(cfg *circuitbreaker.Config) {
		cfg.HalfOpenMaxConcurrent = max
	}
}

// CircuitBreaker returns a route-specific circuit breaker middleware.
// If the circuit breaker is disabled in configs, this is a pass-through middleware.
func CircuitBreaker(cfg *configs.Config, opts ...CircuitBreakerOption) fiber.Handler {
	if cfg == nil || !cfg.Security.CircuitBreaker.Enabled {
		// Pass-through when disabled
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}

	// 1. Initialize default circuitbreaker.Config from application config
	cbCfg := circuitbreaker.Config{
		FailureThreshold:      cfg.Security.CircuitBreaker.FailureThreshold,
		Timeout:               time.Duration(cfg.Security.CircuitBreaker.Timeout) * time.Second,
		SuccessThreshold:      cfg.Security.CircuitBreaker.SuccessThreshold,
		HalfOpenMaxConcurrent: cfg.Security.CircuitBreaker.HalfOpenMaxConcurrent,
		OnOpen: func(c fiber.Ctx) error {
			return exceptions.NewServiceUnavailableException("Service unavailable due to open circuit breaker")
		},
		OnHalfOpen: func(c fiber.Ctx) error {
			return exceptions.NewServiceUnavailableException("Service under recovery, too many requests")
		},
	}

	// 2. Apply code-driven overrides
	for _, opt := range opts {
		opt(&cbCfg)
	}

	// 3. Create the underlying handler
	cb := circuitbreaker.New(cbCfg)

	// 4. Return the middleware handler
	return circuitbreaker.Middleware(cb)
}
