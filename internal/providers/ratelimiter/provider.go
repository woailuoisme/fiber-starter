package ratelimiter

import (
	"sync"
	"time"

	"fiber-starter/configs"
	ratelimiterContracts "fiber-starter/internal/providers/ratelimiter/Contracts"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

type Service struct {
	mu       sync.RWMutex
	limiters map[string]fiber.Handler
	config   configs.LimiterConfig
}

// Register initializes and returns the rate limiter service contract.
func Register(cfg configs.LimiterConfig) (ratelimiterContracts.Limiter, error) {
	return &Service{
		limiters: make(map[string]fiber.Handler),
		config:   cfg,
	}, nil
}

// Strategy returns the configuration for a named strategy
func (s *Service) Strategy(name string) (configs.RateLimitConfig, bool) {
	strat, ok := s.config.Strategies[name]
	return strat, ok
}

// Middleware returns a fiber middleware for a named strategy
func (s *Service) Middleware(name string) fiber.Handler {
	s.mu.RLock()
	if handler, ok := s.limiters[name]; ok {
		s.mu.RUnlock()
		return handler
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check
	if handler, ok := s.limiters[name]; ok {
		return handler
	}

	requestedName := name
	cacheName := name

	// Resolve strategy
	strat, ok := s.config.Strategies[name]
	if !ok {
		// Fallback to default if it exists and isn't the same name
		if s.config.Default != "" && s.config.Default != name {
			cacheName = s.config.Default
			if handler, ok := s.limiters[cacheName]; ok {
				s.limiters[requestedName] = handler
				return handler
			}

			strat, ok = s.config.Strategies[cacheName]
		}

		// Final fallback to a very permissive default
		if !ok {
			strat = configs.RateLimitConfig{Max: 100, Window: 60}
			cacheName = requestedName
		}
	}

	handler := limiter.New(limiter.Config{
		Max:        strat.Max,
		Expiration: time.Duration(strat.Window) * time.Second,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests, please try again later.",
			})
		},
	})

	s.limiters[requestedName] = handler
	s.limiters[cacheName] = handler
	return handler
}
