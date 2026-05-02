package tests

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	controllers "fiber-starter/app/Http/Controllers"
	database "fiber-starter/database"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoints_DoNotRegress(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)

	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	_, err = conn.GetDB()
	if err != nil {
		t.Skipf("Skip test: cannot initialize database - %v", err)
	}

	hc := controllers.NewHealthController(cfg, conn, nil)

	app := fiber.New()
	app.Get("/health", hc.Health)
	app.Get("/ready", hc.Ready)

	healthResp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, healthResp.StatusCode)

	readyResp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, readyResp.StatusCode)

	var payload map[string]any
	require.NoError(t, json.NewDecoder(readyResp.Body).Decode(&payload))
	assert.Equal(t, "ok", payload["status"])
}
