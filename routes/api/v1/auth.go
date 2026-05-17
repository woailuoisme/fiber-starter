package v1

import (
	"strings"
	"time"

	controllers "fiber-starter/app/Http/Controllers"
	middleware "fiber-starter/app/Http/Middleware"
	providers "fiber-starter/app/Providers"

	"github.com/gofiber/fiber/v3"
)

const authRouteTimeout = 30 * time.Second

// AuthRoutePaths keeps the Better Auth style route paths in one place.
type AuthRoutePaths struct {
	BasePath             string
	SignUp               string
	VerifySignUp         string
	SignIn               string
	RefreshSession       string
	SignOut              string
	UpdatePassword       string
	SendPasswordReset    string
	VerifyPasswordReset  string
	ConfirmPasswordReset string
	Session              string
}

// DefaultAuthRoutePaths returns the Better Auth inspired auth route map.
func DefaultAuthRoutePaths() AuthRoutePaths {
	return AuthRoutePaths{
		BasePath:             "/api/v1/auth",
		SignUp:               "/sign-up",
		VerifySignUp:         "/sign-up/verify",
		SignIn:               "/sign-in",
		RefreshSession:       "/refresh",
		SignOut:              "/sign-out",
		UpdatePassword:       "/change-password",
		SendPasswordReset:    "/reset-password",
		VerifyPasswordReset:  "/reset-password/verify",
		ConfirmPasswordReset: "/reset-password/confirm",
		Session:              "/session",
	}
}

// Setup registers Better Auth style auth routes under /api/v1/auth.
func Setup(router fiber.Router, jwtProtected fiber.Handler, authHandler controllers.AuthRouteHandler) {
	rt := providers.App()
	paths := DefaultAuthRoutePaths()
	basePath := strings.TrimPrefix(paths.BasePath, "/api/v1")

	authRouter := middleware.NewTimeoutRouter(
		router.Group(basePath, middleware.Throttle(rt.RateLimiter, "api"), middleware.IdempotencyMiddleware()),
		authRouteTimeout,
	)

	authRouter.Post(paths.SignUp, middleware.Throttle(rt.RateLimiter, "login"), authHandler.SignUp)
	authRouter.Post(paths.VerifySignUp, middleware.Throttle(rt.RateLimiter, "login"), authHandler.VerifySignUp)
	authRouter.Post(paths.SignIn, middleware.Throttle(rt.RateLimiter, "login"), authHandler.SignIn)
	authRouter.Post(paths.RefreshSession, authHandler.RefreshSession)
	authRouter.Post(paths.SignOut, jwtProtected, authHandler.SignOut)
	authRouter.Post(paths.UpdatePassword, jwtProtected, authHandler.UpdatePassword)
	authRouter.Post(paths.SendPasswordReset, authHandler.SendPasswordReset)
	authRouter.Post(paths.VerifyPasswordReset, authHandler.VerifyPasswordReset)
	authRouter.Post(paths.ConfirmPasswordReset, authHandler.ConfirmPasswordReset)
	authRouter.Get(paths.Session, jwtProtected, authHandler.Session)
}
