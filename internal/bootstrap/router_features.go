package bootstrap

import (
	"lfiber/internal/features/auth"
	"lfiber/internal/features/user"

	"github.com/gofiber/fiber/v3"
)

// registerFeatureRoutes 统一在此聚合注册所有的业务领域（Features）路由。
// 解耦主 router.go，使其无需 import 任何具体的业务包。
func registerFeatureRoutes(router fiber.Router) {
	auth.RegisterRoutes(router)
	user.RegisterRoutes(router)
}
