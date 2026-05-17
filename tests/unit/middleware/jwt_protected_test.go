package tests

import (
	"net/http/httptest"
	"testing"

	"fiber-starter/configs"
	middleware "fiber-starter/internal/common/middleware"
	support "fiber-starter/internal/common/support"
	models "fiber-starter/internal/features/user"
	"fiber-starter/tests/internal/testkit/mocks"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestJWTProtected_RejectsBlacklistedToken(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: support.HandleHTTPError,
	})
	middleware.SetupMiddleware(app, nil)

	cfg := &configs.Config{JWT: configs.JWTConfig{Secret: "secret"}}
	token, err := middleware.GenerateToken(&models.User{ID: 1, Email: "user@example.com", Name: "User"}, cfg)
	require.NoError(t, err)

	cache := new(mocks.CacheService)
	cache.On("Exists", "blacklist:"+token).Return(true, nil).Once()

	app.Get("/secure", middleware.JWTProtected(cfg, cache), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	cache.AssertExpectations(t)
}

func TestJWTProtected_ReturnsUnavailableWhenCacheFails(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: support.HandleHTTPError,
	})
	middleware.SetupMiddleware(app, nil)

	cfg := &configs.Config{JWT: configs.JWTConfig{Secret: "secret"}}
	token, err := middleware.GenerateToken(&models.User{ID: 1, Email: "user@example.com", Name: "User"}, cfg)
	require.NoError(t, err)

	cache := new(mocks.CacheService)
	cache.On("Exists", mock.Anything).Return(false, assert.AnError).Once()

	app.Get("/secure", middleware.JWTProtected(cfg, cache), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	cache.AssertExpectations(t)
}
