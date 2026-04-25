package tests

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	controllers "fiber-starter/app/Http/Controllers"
	database "fiber-starter/database"

	"github.com/gofiber/fiber/v3"
)

func TestHealthEndpoints_DoNotRegress(t *testing.T) {
	cfg := newTestConfigSQLite(t)

	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Skip test: cannot connect to database - %v", err)
		return
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	if _, err := conn.GetDB(); err != nil {
		t.Skipf("Skip test: cannot initialize database - %v", err)
		return
	}

	hc := controllers.NewHealthController(cfg, conn, nil)

	app := fiber.New()
	app.Get("/health", hc.Health)
	app.Get("/ready", hc.Ready)

	healthResp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatalf("Request /health failed: %v", err)
	}
	if healthResp.StatusCode != fiber.StatusOK {
		t.Fatalf("/health status code incorrect: got=%d want=%d", healthResp.StatusCode, fiber.StatusOK)
	}

	readyResp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
	if err != nil {
		t.Fatalf("Request /ready failed: %v", err)
	}
	if readyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("/ready status code incorrect: got=%d want=%d", readyResp.StatusCode, fiber.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(readyResp.Body).Decode(&payload); err != nil {
		t.Fatalf("Parse /ready response failed: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("/ready status incorrect: got=%v want=%v", payload["status"], "ok")
	}
}
