package bootstrap

import (
	"time"

	"fiber-starter/configs"
	helpers "fiber-starter/internal/support"

	json "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

// NewHTTPApp creates and configures a new Fiber application instance
func NewHTTPApp(cfg *configs.Config) *fiber.App {
	fiberCfg := cfg.App.Fiber

	return fiber.New(fiber.Config{
		ServerHeader:      fiberCfg.ServerHeader,
		BodyLimit:         defaultInt(fiberCfg.BodyLimit, 4*1024*1024),
		Concurrency:       defaultInt(fiberCfg.Concurrency, 256*1024),
		ReadBufferSize:    defaultInt(fiberCfg.ReadBufferSize, 16*1024),
		ReadTimeout:       time.Duration(defaultInt(fiberCfg.ReadTimeout, 30)) * time.Second,
		WriteTimeout:      time.Duration(defaultInt(fiberCfg.WriteTimeout, 30)) * time.Second,
		IdleTimeout:       time.Duration(defaultInt(fiberCfg.IdleTimeout, 120)) * time.Second,
		TrustProxy:        fiberCfg.TrustProxy,
		ProxyHeader:       defaultStr(fiberCfg.ProxyHeader, fiber.HeaderXForwardedFor),
		StreamRequestBody: fiberCfg.StreamRequestBody,
		Immutable:         fiberCfg.Immutable,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
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
