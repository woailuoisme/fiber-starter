package tests

import (
	"errors"
	"net/http/httptest"
	"testing"

	"fiber-starter/configs"
	monitoring "fiber-starter/internal/features/monitoring"
	providers "fiber-starter/internal/providers"
	"fiber-starter/internal/support/health"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadinessContract_DegradedNonCriticalDependencyReturns200(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	cfg.Cache.Enabled = true
	cfg.Mail.Enabled = false
	cfg.Queue.Enabled = false
	cfg.Search.Enabled = false
	cfg.Storage.Enabled = false
	cfg.Services.Dependencies = map[string]configs.ServiceDependencyConfig{
		"database": {Critical: true},
		"cache":    {Critical: false},
	}

	rt := &providers.Runtime{
		Config:     cfg,
		Connection: &testkit.StubConnection{},
		Cache:      nil,
	}
	providers.SetInstance(rt)

	app := fiber.New()
	app.Get("/ready", monitoring.NewHealthController(cfg).Ready)

	resp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	payload := testkit.JSONBody(t, resp)
	assert.Equal(t, health.OverallDegraded, payload["status"])
	services := payload["services"].(map[string]any)
	cache := services["cache"].(map[string]any)
	assert.Equal(t, health.StatusDegraded, cache["status"])
	assert.Equal(t, false, cache["critical"])
}

func TestReadinessContract_CriticalDependencyFailureReturns503(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	cfg.Cache.Enabled = false
	cfg.Mail.Enabled = false
	cfg.Queue.Enabled = false
	cfg.Search.Enabled = false
	cfg.Storage.Enabled = false
	cfg.Services.Dependencies = map[string]configs.ServiceDependencyConfig{
		"database": {Critical: true},
	}

	rt := &providers.Runtime{
		Config:     cfg,
		Connection: &testkit.StubConnection{HealthErr: errors.New("dsn=postgres://user:pass@localhost/db unavailable")},
	}
	providers.SetInstance(rt)

	app := fiber.New()
	app.Get("/ready", monitoring.NewHealthController(cfg).Ready)

	resp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	payload := testkit.JSONBody(t, resp)
	assert.Equal(t, health.OverallFail, payload["status"])
	services := payload["services"].(map[string]any)
	database := services["database"].(map[string]any)
	assert.Equal(t, health.StatusFail, database["status"])
	assert.Equal(t, true, database["critical"])
	assert.NotContains(t, database["error"], "pass@")
}
