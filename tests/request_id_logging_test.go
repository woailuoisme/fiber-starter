package tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	middleware "fiber-starter/app/Http/Middleware"
	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var stdoutMu sync.Mutex

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = old
		_ = w.Close()
	}()

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outCh <- buf.String()
	}()

	fn()
	os.Stdout = old
	_ = w.Close()

	return <-outCh
}

func TestRequestID_GeneratedAndLogged(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := helpers.Logger
	helpers.Logger = zap.New(core)
	t.Cleanup(func() {
		helpers.Logger = prevLogger
	})

	var resp *http.Response
	out := captureStdout(t, func() {
		app := fiber.New()
		middleware.SetupMiddleware(app)
		app.Get("/boom", func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusBadRequest, "boom")
		})

		r, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		resp = r
	})

	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Fatalf("Expected response to contain X-Request-ID")
	}

	var foundError bool
	for _, entry := range observed.All() {
		m := entry.ContextMap()
		if entry.Message == "http_error" {
			foundError = true
			if m["request_id"] != requestID {
				t.Fatalf("error log request_id mismatch: got=%v want=%v", m["request_id"], requestID)
			}
		}
	}

	if !foundError {
		t.Fatalf("Did not capture http_error log")
	}

	if !strings.Contains(out, requestID) {
		t.Fatalf("Access log does not contain request_id: %q", requestID)
	}
}

func TestRequestID_PreservedAndLogged(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	prevLogger := helpers.Logger
	helpers.Logger = zap.New(core)
	t.Cleanup(func() {
		helpers.Logger = prevLogger
	})

	var resp *http.Response
	out := captureStdout(t, func() {
		app := fiber.New()
		middleware.SetupMiddleware(app)
		app.Get("/boom", func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusBadRequest, "boom")
		})

		req := httptest.NewRequest("GET", "/boom", nil)
		req.Header.Set("X-Request-ID", "rid-123")
		r, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		resp = r
	})

	requestID := resp.Header.Get("X-Request-ID")
	if requestID != "rid-123" {
		t.Fatalf("Expected to preserve X-Request-ID: got=%q want=%q", requestID, "rid-123")
	}

	var foundError bool
	for _, entry := range observed.All() {
		m := entry.ContextMap()
		if entry.Message == "http_error" {
			foundError = true
			if m["request_id"] != requestID {
				t.Fatalf("error log request_id mismatch: got=%v want=%v", m["request_id"], requestID)
			}
		}
	}

	if !foundError {
		t.Fatalf("Did not capture http_error log")
	}

	if !strings.Contains(out, requestID) {
		t.Fatalf("Access log does not contain request_id: %q", requestID)
	}
}
