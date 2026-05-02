package bootstrap

import (
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	json "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func createFiberApp(cfg *config.Config) *fiber.App {
	fiberCfg := cfg.App.Fiber

	concurrency := fiberCfg.Concurrency
	if concurrency == 0 {
		concurrency = 256 * 1024
	}

	bodyLimit := fiberCfg.BodyLimit
	if bodyLimit == 0 {
		bodyLimit = 4 * 1024 * 1024
	}

	readBufferSize := fiberCfg.ReadBufferSize
	if readBufferSize == 0 {
		readBufferSize = 16 * 1024
	}

	readTimeout := fiberCfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 30
	}

	writeTimeout := fiberCfg.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 30
	}

	idleTimeout := fiberCfg.IdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = 120
	}

	proxyHeader := fiberCfg.ProxyHeader
	if proxyHeader == "" {
		proxyHeader = fiber.HeaderXForwardedFor
	}

	return fiber.New(fiber.Config{
		ServerHeader:      fiberCfg.ServerHeader,
		BodyLimit:         bodyLimit,
		Concurrency:       concurrency,
		ReadBufferSize:    readBufferSize,
		ReadTimeout:       time.Duration(readTimeout) * time.Second,
		WriteTimeout:      time.Duration(writeTimeout) * time.Second,
		IdleTimeout:       time.Duration(idleTimeout) * time.Second,
		TrustProxy:        fiberCfg.TrustProxy,
		ProxyHeader:       proxyHeader,
		StreamRequestBody: fiberCfg.StreamRequestBody,
		Immutable:         fiberCfg.Immutable,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
		ErrorHandler:      helpers.HandleHTTPError,
	})
}
