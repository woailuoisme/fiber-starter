package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"fiber-starter/configs"
	models "fiber-starter/internal/features/user"
	realtime "fiber-starter/internal/providers/realtime"
	"fiber-starter/tests/internal/testkit"

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
	cfg := &configs.Config{
		WebSocket: configs.WebSocketConfig{
			AppKey:    "app-key",
			AppSecret: "secret",
		},
	}
	user := &models.User{ID: 123, Email: "user@example.com", Name: "User"}

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
	_, _ = mac.Write([]byte("socket-1:presence-room.1"))
	expectedPresence := "app-key:" + hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expectedPresence, presenceResp.Auth)
	assert.NotEmpty(t, presenceResp.ChannelData)
	assert.Contains(t, presenceResp.ChannelData, strconv.FormatInt(user.ID, 10))
}

func TestRealtime_AuthHandler_UsesCurrentUser(t *testing.T) {
	cfg := &configs.Config{
		WebSocket: configs.WebSocketConfig{
			AppKey:    "app-key",
			AppSecret: "secret",
		},
	}
	mgr := realtime.NewManager(cfg)
	t.Cleanup(func() {
		require.NoError(t, mgr.Close())
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
