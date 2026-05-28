package providers_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"lfiber/configs"
	helpers "lfiber/internal/support"
	"lfiber/pkg/logviewer"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogViewer_ReadLogFileBackward(t *testing.T) {
	// Create a temp directory for logs
	tempDir, err := os.MkdirTemp("", "logviewer-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	filePath := filepath.Join(tempDir, "test.log")

	// Write mock logs: newest logs are at the bottom
	mockLogs := []string{
		`{"level":"INFO","timestamp":"2026-05-28 10:00:00","caller":"test.go:10","message":"First log","user_id":123}`,
		`This is a plain text log line`,
		`{"level":"ERROR","timestamp":"2026-05-28 10:02:00","caller":"test.go:20","message":"An error occurred","request_id":"abc"}`,
		`{"level":"WARN","timestamp":"2026-05-28 10:03:00","caller":"test.go:30","message":"A warning log"}`,
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	for _, line := range mockLogs {
		_, err = file.WriteString(line + "\n")
		require.NoError(t, err)
	}
	require.NoError(t, file.Close())

	// 1. Read all logs backward
	entries, total, stats, err := logviewer.Export_readLogFileBackward(filePath, 1, 100, "", "")
	require.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Len(t, entries, 4)
	assert.Equal(t, 4, stats.Total)
	assert.Equal(t, 2, stats.Info) // 1st JSON info, 2nd plain text (defaults to info)
	assert.Equal(t, 1, stats.Warn)
	assert.Equal(t, 1, stats.Error)
	assert.Equal(t, 0, stats.Fatal)
	assert.Equal(t, 0, stats.Debug)

	// Verify reverse order (newest first)
	assert.Equal(t, "WARN", entries[0].Level)
	assert.Equal(t, "A warning log", entries[0].Message)
	assert.Equal(t, "ERROR", entries[1].Level)
	assert.Equal(t, "An error occurred", entries[1].Message)
	assert.Equal(t, "INFO", entries[2].Level)
	assert.Equal(t, "This is a plain text log line", entries[2].Message)
	assert.Equal(t, "INFO", entries[3].Level)
	assert.Equal(t, "First log", entries[3].Message)

	// Verify context parsing in JSON logs
	assert.InDelta(t, 123.0, entries[3].Context["user_id"].(float64), 0.001)
	assert.Equal(t, "abc", entries[1].Context["request_id"])

	// 2. Pagination: Page 2, Limit 2
	entriesPage2, totalPage2, _, err := logviewer.Export_readLogFileBackward(filePath, 2, 2, "", "")
	require.NoError(t, err)
	assert.Equal(t, 4, totalPage2)
	assert.Len(t, entriesPage2, 2)
	assert.Equal(t, "This is a plain text log line", entriesPage2[0].Message)
	assert.Equal(t, "First log", entriesPage2[1].Message)

	// 3. Level filter: ERROR
	entriesError, totalError, _, err := logviewer.Export_readLogFileBackward(filePath, 1, 10, "", "ERROR")
	require.NoError(t, err)
	assert.Equal(t, 1, totalError)
	assert.Len(t, entriesError, 1)
	assert.Equal(t, "An error occurred", entriesError[0].Message)

	// 4. Keyword search: "warning"
	entriesSearch, totalSearch, _, err := logviewer.Export_readLogFileBackward(filePath, 1, 10, "warning", "")
	require.NoError(t, err)
	assert.Equal(t, 1, totalSearch)
	assert.Len(t, entriesSearch, 1)
	assert.Equal(t, "A warning log", entriesSearch[0].Message)
}

func TestLogViewer_BasicAuthFailsClosed(t *testing.T) {
	assert.False(t, logviewer.Export_validBasicAuth("admin", "secret", "", "secret"))
	assert.False(t, logviewer.Export_validBasicAuth("admin", "secret", "admin", ""))
	assert.False(t, logviewer.Export_validBasicAuth("admin", "wrong", "admin", "secret"))
	assert.True(t, logviewer.Export_validBasicAuth("admin", "secret", "admin", "secret"))
}

func TestLogViewer_RoutesUseEnvelopeAndProtectFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logviewer-routes")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "app.log"), []byte(`{"level":"INFO","timestamp":"2026-05-28 10:00:00","message":"hello"}`+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "notes.txt"), []byte("hidden"), 0o644))

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	cfg := configs.LogViewerConfig{
		Enabled:  true,
		Path:     tempDir,
		Username: "admin",
		Password: "secret",
	}
	logviewer.Register(app.Group("/logs"), cfg, "storage/logs")

	unauthorizedReq := httptest.NewRequest("GET", "/logs/api/files", nil)
	unauthorizedResp, err := app.Test(unauthorizedReq)
	require.NoError(t, err)
	assert.Equal(t, 401, unauthorizedResp.StatusCode)

	filesReq := httptest.NewRequest("GET", "/logs/api/files", nil)
	filesReq.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
	filesResp, err := app.Test(filesReq)
	require.NoError(t, err)
	filesPayload := readJSONMap(t, filesResp)
	assert.Equal(t, true, filesPayload["success"])
	data := filesPayload["data"].(map[string]any)
	files := data["files"].([]any)
	require.Len(t, files, 1)
	assert.Equal(t, "app.log", files[0].(map[string]any)["name"])

	entriesReq := httptest.NewRequest("GET", "/logs/api/entries?file=../app.log", nil)
	entriesReq.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
	entriesResp, err := app.Test(entriesReq)
	require.NoError(t, err)
	assert.Equal(t, 400, entriesResp.StatusCode)
}

func TestLogViewer_RedirectsBarePathForRelativeAssets(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	cfg := configs.LogViewerConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
	}
	logviewer.Register(app.Group("/logs"), cfg, "storage/logs")

	req := httptest.NewRequest("GET", "/logs", nil)
	req.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusPermanentRedirect, resp.StatusCode)
	assert.Equal(t, "/logs/", resp.Header.Get("Location"))
}

func TestLogViewer_ServesVendorAssetsWithExecutableMimeTypes(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	cfg := configs.LogViewerConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
	}
	logviewer.Register(app.Group("/logs"), cfg, "storage/logs")

	tests := []struct {
		path        string
		contentType string
	}{
		{path: "/logs/assets/vendor/tw-animate.1.4.0.css", contentType: "text/css; charset=utf-8"},
		{path: "/logs/assets/vendor/fontawesome.7.2.0.min.css", contentType: "text/css; charset=utf-8"},
		{path: "/logs/assets/vendor/fontawesome-solid.7.2.0.min.css", contentType: "text/css; charset=utf-8"},
		{path: "/logs/assets/vendor/fa-solid-900.7.2.0.woff2", contentType: "font/woff2"},
		{path: "/logs/assets/vendor/tailwind-browser.4.3.0.global.js", contentType: "application/javascript; charset=utf-8"},
		{path: "/logs/assets/vendor/alpine.3.15.12.min.js", contentType: "application/javascript; charset=utf-8"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, tt.contentType, resp.Header.Get("Content-Type"))
	}
}

func TestLogViewer_IndexUsesFontAwesomeWithoutInlineSVG(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	cfg := configs.LogViewerConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
	}
	logviewer.Register(app.Group("/logs"), cfg, "storage/logs")

	req := httptest.NewRequest("GET", "/logs/", nil)
	req.SetBasicAuth("admin", "secret")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	html := string(body)

	assert.NotContains(t, html, "<svg")
	assert.NotContains(t, html, "</svg>")
	assert.Contains(t, html, "assets/vendor/fontawesome.7.2.0.min.css")
	assert.Contains(t, html, "assets/vendor/fontawesome-solid.7.2.0.min.css")
	assert.Contains(t, html, "fa-solid")
}

func TestLogViewer_DeleteDisabledByDefault(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logviewer-delete")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	logPath := filepath.Join(tempDir, "app.log")
	require.NoError(t, os.WriteFile(logPath, []byte("hello\n"), 0o644))

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	cfg := configs.LogViewerConfig{
		Enabled:  true,
		Path:     tempDir,
		Username: "admin",
		Password: "secret",
	}
	logviewer.Register(app.Group("/logs"), cfg, "storage/logs")

	req := httptest.NewRequest("DELETE", "/logs/api/files/app.log", nil)
	req.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
	assert.FileExists(t, logPath)
}

func basicAuthHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func readJSONMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()

	var payload map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	return payload
}
