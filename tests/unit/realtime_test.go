package tests

import (
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // Pusher REST compatibility signs body_md5.
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	models "lfiber/internal/features/user"
	"lfiber/pkg/realtime"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealtime_ParseChannelKinds(t *testing.T) {
	publicChannel, err := realtime.ParseChannel("orders")
	require.NoError(t, err)
	assert.True(t, publicChannel.IsPublic())

	privateChannel, err := realtime.ParseChannel("private-user.123")
	require.NoError(t, err)
	assert.True(t, privateChannel.IsPrivate())
	assert.False(t, privateChannel.IsPresence())

	presenceChannel, err := realtime.ParseChannel("presence-room.1")
	require.NoError(t, err)
	assert.True(t, presenceChannel.IsPresence())
}

func TestRealtime_BuildAuthResponse_PrivateAndPresence(t *testing.T) {
	cfg := &realtime.Config{
		AppKey:    "app-key",
		AppSecret: "secret",
	}
	user := realtime.User{
		ID: "123",
		Info: map[string]any{
			"id":    int64(123),
			"email": "user@example.com",
			"name":  "User",
		},
	}

	resp, err := realtime.BuildAuthResponse(cfg, "socket-1", "private-user.123", user)
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write([]byte("socket-1:private-user.123"))
	expected := "app-key:" + hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, resp.Auth)
	assert.Empty(t, resp.ChannelData)

	presenceResp, err := realtime.BuildAuthResponse(cfg, "socket-1", "presence-room.1", user)
	require.NoError(t, err)
	mac = hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write([]byte("socket-1:presence-room.1:" + presenceResp.ChannelData))
	expectedPresence := "app-key:" + hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expectedPresence, presenceResp.Auth)
	assert.NotEmpty(t, presenceResp.ChannelData)
	assert.Contains(t, presenceResp.ChannelData, "123")
}

func TestRealtime_AuthHandler_UsesCurrentUser(t *testing.T) {
	cfg := &realtime.Config{
		AppKey:    "app-key",
		AppSecret: "secret",
	}
	mgr := realtime.NewManager(cfg, realtime.NewNoopLogger())
	t.Cleanup(func() {
		require.NoError(t, mgr.Close())
	})

	// 注入 AuthUserResolver
	mgr.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		u, ok := c.Locals("user").(*models.User)
		if !ok || u == nil {
			return realtime.User{}, fiber.ErrUnauthorized
		}
		return realtime.User{
			ID: strconv.FormatInt(u.ID, 10),
			Info: map[string]any{
				"id":    u.ID,
				"email": u.Email,
				"name":  u.Name,
			},
		}, nil
	})

	app := fiber.New()
	app.Post("/broadcasting/auth", func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: 99, Email: "me@example.com", Name: "Me"})
		return mgr.AuthHandler()(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/broadcasting/auth", strings.NewReader(`{"socket_id":"socket-2","channel_name":"presence-room.1"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body := testkit.JSONBody(t, resp)
	assert.Contains(t, body, "auth")
	assert.Contains(t, body, "channel_data")
}

func TestRealtime_APIHandler_ValidatesPusherSignature(t *testing.T) {
	cfg := &realtime.Config{
		AppID:     "app-id",
		AppKey:    "app-key",
		AppSecret: "secret",
		BusMode:   "memory",
	}
	mgr := realtime.NewManager(cfg, realtime.NewNoopLogger())
	t.Cleanup(func() {
		require.NoError(t, mgr.Close())
	})

	app := fiber.New()
	api := mgr.APIHandler()
	app.Post("/apps/:appID/events", api)

	body := []byte(`{"name":"orders.updated","channels":["private-orders.1"],"data":"{\"id\":1}"}`)
	req := httptest.NewRequest(http.MethodPost, signedURL(t, http.MethodPost, "/apps/app-id/events", "app-key", "secret", body), strings.NewReader(string(body)))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	invalidReq := httptest.NewRequest(http.MethodPost, "/apps/app-id/events?auth_key=app-key&auth_version=1.0&auth_timestamp=1893456000&auth_signature=bad", strings.NewReader(string(body)))
	invalidReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	invalidResp, err := app.Test(invalidReq)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, invalidResp.StatusCode)
}

func TestRealtime_WebSocketHandler_AliasesLegacyHandler(t *testing.T) {
	mgr := realtime.NewManager(&realtime.Config{AppKey: "app-key"}, realtime.NewNoopLogger())
	t.Cleanup(func() {
		require.NoError(t, mgr.Close())
	})

	assert.NotNil(t, mgr.WebSocketHandler())
	assert.NotNil(t, mgr.Handler())
}

func TestRealtime_SSEFrame_UsesBusinessEventAndEnvelopeData(t *testing.T) {
	frame, err := realtime.NewSSEFrame(realtime.Envelope{
		Event:          "orders.updated",
		Channel:        "private-orders.1",
		Data:           json.RawMessage(`{"id":1}`),
		OriginSocketID: "socket-1",
	})
	require.NoError(t, err)

	assert.Equal(t, "orders.updated", frame.Name)
	assert.NotEmpty(t, frame.ID)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(frame.Data, &payload))
	assert.Equal(t, "private-orders.1", payload["channel"])
	assert.Equal(t, "orders.updated", payload["event"])
	assert.Equal(t, "socket-1", payload["socket_id"])
	assert.Equal(t, map[string]any{"id": float64(1)}, payload["data"])
}

func TestRealtime_SSEHandler_RejectsPrivateChannelWithoutSignature(t *testing.T) {
	cfg := &realtime.Config{
		AppKey:    "app-key",
		AppSecret: "secret",
		BusMode:   "memory",
	}
	mgr := realtime.NewManager(cfg, realtime.NewNoopLogger())
	t.Cleanup(func() {
		require.NoError(t, mgr.Close())
	})

	app := fiber.New()
	app.Get("/sse/app/:appKey", mgr.SSEHandler())

	req := httptest.NewRequest(http.MethodGet, "/sse/app/app-key?channels=private-orders.1", nil)
	req.Header.Set(fiber.HeaderAccept, "text/event-stream")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func signedURL(t *testing.T, method, path, key, secret string, body []byte) string {
	t.Helper()

	sum := md5.Sum(body)
	params := map[string]string{
		"auth_key":       key,
		"auth_timestamp": "1893456000",
		"auth_version":   "1.0",
		"body_md5":       hex.EncodeToString(sum[:]),
	}
	params["auth_signature"] = signPusherRequest(method, path, secret, params)

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return path + "?" + values.Encode()
}

func signPusherRequest(method, path, secret string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key != "auth_signature" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	values := url.Values{}
	for _, key := range keys {
		values.Set(key, params[key])
	}

	stringToSign := strings.ToUpper(method) + "\n" + path + "\n" + values.Encode()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprint(mac, stringToSign)
	return hex.EncodeToString(mac.Sum(nil))
}
