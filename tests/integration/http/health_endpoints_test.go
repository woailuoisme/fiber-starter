package tests

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"lfiber/configs"
	monitoring "lfiber/internal/features/monitoring"
	providers "lfiber/internal/providers"
	cache "lfiber/internal/providers/cache"
	database "lfiber/internal/providers/database"
	"lfiber/internal/support/health"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoints_DoNotRegress(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	conn := database.NewConnection(cfg)
	t.Cleanup(func() {
		_ = conn.Close()
	})

	_, err := conn.GetDB()
	if err != nil {
		t.Skipf("Skip test: cannot initialize database - %v", err)
	}

	// Explicitly disable optional services so the health check passes with just the DB
	cfg.Mail.Enabled = false
	cfg.Queue.Enabled = false
	cfg.Search.Enabled = false
	cfg.Redis.Enabled = false
	cfg.WebSocket.Enabled = false
	cfg.Cache.Driver = "memory"

	cacheManager := cache.NewManager(cfg)
	rt := &providers.Runtime{
		Config:     cfg,
		Connection: conn,
		Cache:      cacheManager.Store(),
	}
	providers.SetInstance(rt)

	hc := monitoring.NewHealthController(cfg)

	app := fiber.New()
	app.Get("/health", hc.Health)
	app.Get("/ready", hc.Ready)

	reqHealth := httptest.NewRequest("GET", "/health", nil)
	healthResp, err := app.Test(reqHealth)
	require.NoError(t, err)
	if healthResp.StatusCode != 200 {
		body, _ := io.ReadAll(healthResp.Body)
		t.Logf("Response body: %s", string(body))
	}
	assert.Equal(t, fiber.StatusOK, healthResp.StatusCode)

	reqReady := httptest.NewRequest("GET", "/ready", nil)
	readyResp, err := app.Test(reqReady)
	require.NoError(t, err)
	if readyResp.StatusCode != 200 {
		body, _ := io.ReadAll(readyResp.Body)
		t.Logf("Response body: %s", string(body))
	}
	assert.Equal(t, fiber.StatusOK, readyResp.StatusCode)

	var payload struct {
		Status   string `json:"status"`
		Services map[string]struct {
			Status string `json:"status"`
		} `json:"services"`
	}
	require.NoError(t, json.NewDecoder(readyResp.Body).Decode(&payload))
	assert.Equal(t, "ok", payload.Services["database"].Status)
}

func TestHealthEndpoints_HealthIsLightweightReadyProbesDependencies(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	cfg.Database.Default = "pgsql"
	cfg.Database.Connections = map[string]configs.DBConnection{
		"pgsql": {
			Driver:   "postgres",
			Host:     "127.0.0.1",
			Port:     "1",
			Database: "missing",
			Username: "postgres",
			SSLMode:  "disable",
		},
	}
	cfg.Cache.Enabled = false
	cfg.Mail.Enabled = false
	cfg.Queue.Enabled = false
	cfg.Search.Enabled = false
	cfg.Storage.Enabled = false

	rt := &providers.Runtime{
		Config:     cfg,
		Connection: database.NewConnection(cfg),
	}
	providers.SetInstance(rt)
	t.Cleanup(func() {
		_ = rt.Close()
	})

	hc := monitoring.NewHealthController(cfg)

	app := fiber.New()
	app.Get("/health", hc.Health)
	app.Get("/ready", hc.Ready)

	healthResp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, healthResp.StatusCode)

	readyResp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, readyResp.StatusCode)
}

func TestHealthEndpoints_ReadyReportsDegradedForNonCriticalDependency(t *testing.T) {
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
		Degraded: map[string]string{
			"cache": "redis://default:secret@localhost:6379 unavailable",
		},
	}
	providers.SetInstance(rt)
	t.Cleanup(func() {
		_ = rt.Close()
	})

	hc := monitoring.NewHealthController(cfg)

	app := fiber.New()
	app.Get("/health", hc.Health)
	app.Get("/ready", hc.Ready)

	healthResp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, healthResp.StatusCode)

	readyResp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, readyResp.StatusCode)

	var payload struct {
		Status   string `json:"status"`
		Services map[string]struct {
			Status   string `json:"status"`
			Error    string `json:"error"`
			Critical bool   `json:"critical"`
		} `json:"services"`
	}
	require.NoError(t, json.NewDecoder(readyResp.Body).Decode(&payload))
	assert.Equal(t, health.OverallDegraded, payload.Status)
	assert.Equal(t, health.StatusDegraded, payload.Services["cache"].Status)
	assert.False(t, payload.Services["cache"].Critical)
	assert.NotContains(t, payload.Services["cache"].Error, "secret@")
}

func TestHealthEndpoints_ReadyFailsForCriticalDependency(t *testing.T) {
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
		Connection: &testkit.StubConnection{HealthErr: errors.New("password=topsecret unavailable")},
	}
	providers.SetInstance(rt)
	t.Cleanup(func() {
		_ = rt.Close()
	})

	hc := monitoring.NewHealthController(cfg)

	app := fiber.New()
	app.Get("/ready", hc.Ready)

	readyResp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, readyResp.StatusCode)

	var payload struct {
		Status   string `json:"status"`
		Services map[string]struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"services"`
	}
	require.NoError(t, json.NewDecoder(readyResp.Body).Decode(&payload))
	assert.Equal(t, health.OverallFail, payload.Status)
	assert.Equal(t, health.StatusFail, payload.Services["database"].Status)
	assert.NotContains(t, payload.Services["database"].Error, "topsecret")
}
