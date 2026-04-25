package bootstrap

import (
	"fiber-starter/app/Providers"
	"fiber-starter/routes"

	"github.com/gofiber/fiber/v3"
)

func setupAppRoutes(app *fiber.App, container *providers.Container) error {
	return routes.SetupApplicationRoutes(app, container)
}
