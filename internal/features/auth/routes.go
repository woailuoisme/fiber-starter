package auth

import (
	"time"

	"fiber-starter/configs"
	middleware "fiber-starter/internal/common/middleware"
	cache "fiber-starter/internal/providers/cache/contracts"
	database "fiber-starter/internal/providers/database/contracts"
	hashContracts "fiber-starter/internal/providers/hash/contracts"
	mailContracts "fiber-starter/internal/providers/mail/contracts"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// RegisterRoutes registers auth routes under the provided router group.
func RegisterRoutes(
	router fiber.Router,
	db database.Connection,
	cfg *configs.Config,
	cacheStore cache.Store,
	mailer mailContracts.Mailer,
	hasher hashContracts.Hasher,
) {
	authService := NewAuthService(
		db,
		cfg,
		cacheStore,
		mailer,
		hasher,
	)
	authController := NewAuthController(authService)
	jwtProtected := middleware.JWTProtected(cfg, cacheStore)

	authRouter := middleware.NewTimeoutRouter(
		router.Group("/auth"),
		routeTimeout,
	)

	authRouter.Post("/sign-up", authController.SignUp)
	authRouter.Post("/sign-up/verify", authController.VerifySignUp)
	authRouter.Post("/sign-in", authController.SignIn)
	authRouter.Post("/refresh", authController.RefreshSession)
	authRouter.Post("/sign-out", jwtProtected, authController.SignOut)
	authRouter.Post("/change-password", jwtProtected, authController.UpdatePassword)
	authRouter.Post("/reset-password", authController.SendPasswordReset)
	authRouter.Post("/reset-password/verify", authController.VerifyPasswordReset)
	authRouter.Post("/reset-password/confirm", authController.ConfirmPasswordReset)
	authRouter.Get("/session", jwtProtected, authController.Session)
}
