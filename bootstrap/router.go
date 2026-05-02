package bootstrap

import (
	providers "fiber-starter/app/Providers"
	"fiber-starter/routes"

	"github.com/gofiber/fiber/v3"
)

func setupAppRoutes(app *fiber.App, runtime *providers.Runtime) error {
	return routes.SetupApplicationRoutes(
		app,
		runtime.Config,
		runtime.Cache,
		runtime.AuthController,
		runtime.UserController,
		runtime.HealthController,
	)
}
