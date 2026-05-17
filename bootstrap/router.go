package bootstrap

import (
	"fiber-starter/routes"

	"github.com/gofiber/fiber/v3"
)

func setupAppRoutes(app *fiber.App) error {
	// Directly delegate to routes package, which handles its own dependencies via providers.App()
	return routes.SetupApplicationRoutes(app)
}
