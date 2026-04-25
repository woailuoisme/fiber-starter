package bootstrap

import (
	"encoding/json"

	"fiber-starter/app/Http/Middleware"
	"fiber-starter/config"

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

	return fiber.New(fiber.Config{
		ServerHeader:   fiberCfg.ServerHeader,
		BodyLimit:      bodyLimit,
		Concurrency:    concurrency,
		ReadBufferSize: readBufferSize,
		JSONEncoder:    json.Marshal,
		JSONDecoder:    json.Unmarshal,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return middleware.HandleError(c, err)
		},
	})
}
