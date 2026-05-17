package auth

import (
	"time"

	middleware "fiber-starter/internal/common/middleware"
	"fiber-starter/internal/support/appctx"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// RegisterRoutes registers auth routes under the provided router group.
func RegisterRoutes(router fiber.Router) {
	rt := appctx.App()

	authService := NewAuthService(
		rt.ConnectionValue(),
		rt.AppConfig(),
		rt.CacheStore(),
		rt.EmailServiceValue(),
	)
	authController := NewAuthController(authService)
	jwtProtected := middleware.JWTProtected(rt.AppConfig(), rt.CacheStore())

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
