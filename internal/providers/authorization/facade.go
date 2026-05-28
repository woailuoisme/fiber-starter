package authorization

import (
	exceptions "lfiber/internal/common/exceptions"
	authorizationContracts "lfiber/internal/providers/authorization/contracts"
	"lfiber/internal/support/appctx"

	"github.com/gofiber/fiber/v3"
)

func resolveAuthorizer() authorizationContracts.Authorizer {
	if app := appctx.App(); app != nil {
		return app.AuthorizationService()
	}
	return nil
}

// RequirePermissions returns middleware requiring every named permission.
func RequirePermissions(permissions ...string) fiber.Handler {
	if authorizer := resolveAuthorizer(); authorizer != nil {
		return authorizer.RequirePermissions(permissions...)
	}
	return func(_ fiber.Ctx) error {
		return exceptions.NewAuthorizationException("Authorization service unavailable")
	}
}

// Subject returns the current Casbin subject for a Fiber request.
func Subject(c fiber.Ctx) string {
	if authorizer := resolveAuthorizer(); authorizer != nil {
		return authorizer.Subject(c)
	}
	return ""
}
