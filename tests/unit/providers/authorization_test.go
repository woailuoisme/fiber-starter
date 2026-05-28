package providers_test

import (
	"net/http/httptest"
	"path/filepath"
	"testing"

	"lfiber/configs"
	middleware "lfiber/internal/common/middleware"
	authorization "lfiber/internal/providers/authorization"
	helpers "lfiber/internal/support"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationProvider_RegisterAndRequirePermissions(t *testing.T) {
	cfg := configs.AuthorizationConfig{
		ModelFile:  filepath.Join("..", "..", "configs", "casbin", "model.conf"),
		PolicyFile: filepath.Join("..", "..", "configs", "casbin", "policy.csv"),
	}

	service, err := authorization.Register(cfg)
	require.NoError(t, err)
	require.NotNil(t, service)

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Get("/allowed", withJWTUser(1), service.RequirePermissions("users:list"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/forbidden", withJWTUser(2), service.RequirePermissions("users:list"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/unauthorized", service.RequirePermissions("users:list"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	allowedResp, err := app.Test(httptest.NewRequest("GET", "/allowed", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, allowedResp.StatusCode)

	forbiddenResp, err := app.Test(httptest.NewRequest("GET", "/forbidden", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, forbiddenResp.StatusCode)

	unauthorizedResp, err := app.Test(httptest.NewRequest("GET", "/unauthorized", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, unauthorizedResp.StatusCode)
}

func TestAuthorizationProvider_FailsClosedForMissingPolicyFiles(t *testing.T) {
	_, err := authorization.Register(configs.AuthorizationConfig{
		ModelFile:  filepath.Join(t.TempDir(), "missing-model.conf"),
		PolicyFile: filepath.Join(t.TempDir(), "missing-policy.csv"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve authorization model file")
}

func withJWTUser(id int64) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("user_claims", &middleware.JWTClaims{UserID: id})
		return c.Next()
	}
}

func TestAuthorizationFacade_FailsClosedWithoutApp(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Get("/secure", authorization.RequirePermissions("users:list"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/secure", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
