package testkit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// DoRequest builds and executes an HTTP request against a Fiber app.
func DoRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}

	return resp
}

// ReadBody returns the full response body as a string.
func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body failed: %v", err)
	}

	return string(body)
}

// JSONBody decodes a JSON response into a generic map.
func JSONBody(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response body failed: %v", err)
	}

	return payload
}

// AssertEnvelope verifies the public response envelope without constraining data shape.
func AssertEnvelope(t *testing.T, payload map[string]any, success bool, code int) {
	t.Helper()

	assert.Equal(t, success, payload["success"])
	assert.EqualValues(t, code, payload["code"])
	require.Contains(t, payload, "message")
}

// AssertErrorEnvelope verifies an error response envelope.
func AssertErrorEnvelope(t *testing.T, resp *http.Response, code int) map[string]any {
	t.Helper()

	payload := JSONBody(t, resp)
	AssertEnvelope(t, payload, false, code)
	return payload
}

// AssertSuccessEnvelope verifies a success response envelope.
func AssertSuccessEnvelope(t *testing.T, resp *http.Response, code int) map[string]any {
	t.Helper()

	payload := JSONBody(t, resp)
	AssertEnvelope(t, payload, true, code)
	require.Contains(t, payload, "data")
	return payload
}

// HasRoutePath reports whether any route matches the requested path.
func HasRoutePath(routes []fiber.Route, want string) bool {
	for _, route := range routes {
		if route.Path == want {
			return true
		}
	}
	return false
}

// ContainsBody reports whether the response body contains the expected text.
func ContainsBody(t *testing.T, resp *http.Response, want string) bool {
	t.Helper()

	body := ReadBody(t, resp)
	return strings.Contains(body, want)
}

// BytesBody reads and returns the response body as raw bytes.
func BytesBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body failed: %v", err)
	}

	return body
}

// BufferBody reads and returns the response body using a bytes.Buffer.
func BufferBody(t *testing.T, resp *http.Response) *bytes.Buffer {
	t.Helper()

	return bytes.NewBuffer(BytesBody(t, resp))
}

// AcquireCtx returns a valid fiber.Ctx for testing.
func AcquireCtx(t *testing.T) (fiber.Ctx, *fiber.App) {
	t.Helper()
	app := fiber.New()
	return app.AcquireCtx(&fasthttp.RequestCtx{}), app
}
