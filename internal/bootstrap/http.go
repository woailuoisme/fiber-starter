package bootstrap

import (
	"time"

	"lfiber/configs"
	requests "lfiber/internal/common/requests"
	helpers "lfiber/internal/support"

	json "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

const (
	defaultBodyLimit      = 4 * 1024 * 1024
	defaultConcurrency    = 256 * 1024
	defaultReadBufferSize = 16 * 1024
	defaultReadTimeout    = 30
	defaultWriteTimeout   = 30
	defaultIdleTimeout    = 120
)

// NewHTTPApp creates and configures a new Fiber application instance
func NewHTTPApp(cfg *configs.Config) *fiber.App {
	fiberCfg := cfg.App.Fiber

	return fiber.New(fiber.Config{
		ServerHeader:      fiberCfg.ServerHeader,
		BodyLimit:         defaultInt(fiberCfg.BodyLimit, defaultBodyLimit),
		Concurrency:       defaultInt(fiberCfg.Concurrency, defaultConcurrency),
		ReadBufferSize:    defaultInt(fiberCfg.ReadBufferSize, defaultReadBufferSize),
		ReadTimeout:       time.Duration(defaultInt(fiberCfg.ReadTimeout, defaultReadTimeout)) * time.Second,
		WriteTimeout:      time.Duration(defaultInt(fiberCfg.WriteTimeout, defaultWriteTimeout)) * time.Second,
		IdleTimeout:       time.Duration(defaultInt(fiberCfg.IdleTimeout, defaultIdleTimeout)) * time.Second,
		TrustProxy:        fiberCfg.TrustProxy,
		ProxyHeader:       defaultStr(fiberCfg.ProxyHeader, fiber.HeaderXForwardedFor),
		StreamRequestBody: fiberCfg.StreamRequestBody,
		Immutable:         fiberCfg.Immutable,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
		StructValidator:   requests.NewStructValidator(),
		ErrorHandler:      helpers.HandleHTTPError,
	})
}

func defaultInt(val, def int) int {
	if val <= 0 {
		return def
	}
	return val
}

func defaultStr(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
