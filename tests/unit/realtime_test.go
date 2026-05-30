package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	models "lfiber/internal/features/user"
	"lfiber/pkg/realtime"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealtime_ParseChannelKinds(t *testing.T) {
	publicChannel, err := realtime.ParseChannel("orders")
	require.NoError(t, err)
	assert.True(t, publicChannel.IsPublic())

	privateChannel, err := realtime.ParseChannel("private:user.123")
	require.NoError(t, err)
	assert.True(t, privateChannel.IsPrivate())
	assert.False(t, privateChannel.IsPresence())

	presenceChannel, err := realtime.ParseChannel("presence:room.1")
	require.NoError(t, err)
	assert.True(t, presenceChannel.IsPresence())

	emptyChannel, err := realtime.ParseChannel("")
	require.Error(t, err)
	assert.Empty(t, emptyChannel.Name)
}

func TestRealtime_JWTGeneration_ConnectionAndSubscription(t *testing.T) {
	secret := "my-secret-key"
	userID := "123"
	info := map[string]any{"name": "John Doe", "email": "john@example.com"}

	// 1. Connection Token
	connToken, err := realtime.GenerateConnectionToken(secret, userID, 3600, info)
	require.NoError(t, err)
	require.NotEmpty(t, connToken)

	// Verify Connection Token
	var connClaims realtime.ConnectionClaims
	parsedConnToken, err := jwt.ParseWithClaims(connToken, &connClaims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, parsedConnToken.Valid)
	assert.Equal(t, userID, connClaims.Subject)
	assert.Equal(t, "John Doe", connClaims.Info["name"])

	// 2. Subscription Token
	subToken, err := realtime.GenerateSubscriptionToken(secret, userID, "private:chat.1", 3600, info)
	require.NoError(t, err)
	require.NotEmpty(t, subToken)

	// Verify Subscription Token
	var subClaims realtime.SubscriptionClaims
	parsedSubToken, err := jwt.ParseWithClaims(subToken, &subClaims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, parsedSubToken.Valid)
	assert.Equal(t, userID, subClaims.Subject)
	assert.Equal(t, "private:chat.1", subClaims.Channel)
	assert.Equal(t, "John Doe", subClaims.Info["name"])

	// 3. Error cases
	_, err = realtime.GenerateConnectionToken("", userID, 3600, info)
	require.ErrorContains(t, err, "secret is empty")

	_, err = realtime.GenerateSubscriptionToken("", userID, "private:chat.1", 3600, info)
	require.ErrorContains(t, err, "secret is empty")

	_, err = realtime.GenerateSubscriptionToken(secret, userID, "", 3600, info)
	require.ErrorContains(t, err, "channel is required")
}

func TestRealtime_Registry_Authorization(t *testing.T) {
	mgr := realtime.NewManager(&realtime.Config{Secret: "secret-key"}, nil)

	var authorizedParams map[string]string
	var authorizedUser realtime.User
	var called bool

	// 注册带点号、冒号、斜线占位符以及通配符的多种频道路由
	mgr.AuthorizeChannel("private:users.{id}", func(ctx context.Context, user realtime.User, channel string, params map[string]string) error {
		called = true
		authorizedUser = user
		authorizedParams = params
		if params["id"] == "block" {
			return errors.New("denied")
		}
		return nil
	})

	mgr.AuthorizeChannel("private:rooms:{room_id}", func(ctx context.Context, user realtime.User, channel string, params map[string]string) error {
		called = true
		authorizedUser = user
		authorizedParams = params
		return nil
	})

	mgr.AuthorizeChannel("private/tasks/{task_id}", func(ctx context.Context, user realtime.User, channel string, params map[string]string) error {
		called = true
		authorizedParams = params
		return nil
	})

	appUser := realtime.User{ID: "42"}

	// 1. 点号分隔变量抽取
	called = false
	app := fiber.New()
	app.Post("/auth", mgr.AuthHandler())
	mgr.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		return appUser, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"channel":"private:users.123"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.True(t, called)
	assert.Equal(t, "42", authorizedUser.ID)
	assert.Equal(t, "123", authorizedParams["id"])

	// 2. 冒号分隔变量抽取
	called = false
	req = httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"channel":"private:rooms:abc-def"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.True(t, called)
	assert.Equal(t, "abc-def", authorizedParams["room_id"])

	// 3. 斜线分隔变量抽取
	called = false
	req = httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"channel":"private/tasks/999"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.True(t, called)
	assert.Equal(t, "999", authorizedParams["task_id"])

	// 4. 被拒绝的权限测试
	called = false
	req = httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"channel":"private:users.block"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.True(t, called)
}

func TestRealtime_AuthHandler_ReturnsSubscriptionToken(t *testing.T) {
	cfg := &realtime.Config{
		Enabled:   true,
		Secret:    "secret-key",
		ClientURL: "ws://localhost:8000/connection/websocket",
	}
	mgr := realtime.NewManager(cfg, realtime.NewNoopLogger())

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

	req := httptest.NewRequest(http.MethodPost, "/broadcasting/auth", strings.NewReader(`{"channel":"private:room.1"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body := testkit.JSONBody(t, resp)
	assert.Contains(t, body, "token")

	// 验证生成的订阅 Token
	tokenStr := body["token"].(string)
	var claims realtime.SubscriptionClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		return []byte("secret-key"), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, "99", claims.Subject)
	assert.Equal(t, "private:room.1", claims.Channel)
	assert.Equal(t, "Me", claims.Info["name"])
}

func TestRealtime_AuthHandler_EdgeCases(t *testing.T) {
	cfg := &realtime.Config{
		Enabled: true,
		Secret:  "secret-key",
	}

	// 1. 未配置 authUserResolver -> 500
	mgrNoResolver := realtime.NewManager(cfg, nil)
	app := fiber.New()
	app.Post("/auth", mgrNoResolver.AuthHandler())

	req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"channel":"private:room.1"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// 2. resolver 报错 -> 401
	mgrErrResolver := realtime.NewManager(cfg, nil)
	mgrErrResolver.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		return realtime.User{}, errors.New("auth failed")
	})
	app2 := fiber.New()
	app2.Post("/auth", mgrErrResolver.AuthHandler())

	req = httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"channel":"private:room.1"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err = app2.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// 3. 缺少 channel 参数 -> 400
	req = httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err = app2.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRealtime_WebSocketHandler_ReturnsConnectionConfig(t *testing.T) {
	cfg := &realtime.Config{
		Enabled:   true,
		Secret:    "secret-key",
		ClientURL: "ws://localhost:8000/connection/websocket",
	}
	mgr := realtime.NewManager(cfg, realtime.NewNoopLogger())

	// 注入 AuthUserResolver
	mgr.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		return realtime.User{
			ID:   "100",
			Info: map[string]any{"name": "Tester"},
		}, nil
	})

	app := fiber.New()
	app.Get("/app/:appKey", mgr.WebSocketHandler())

	req := httptest.NewRequest(http.MethodGet, "/app/test-key", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body := testkit.JSONBody(t, resp)
	assert.Equal(t, "ws://localhost:8000/connection/websocket", body["url"])
	assert.Equal(t, "100", body["user_id"])
	assert.Contains(t, body, "token")

	// 验证生成的连接 Token
	tokenStr := body["token"].(string)
	var claims realtime.ConnectionClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		return []byte("secret-key"), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, "100", claims.Subject)
	assert.Equal(t, "Tester", claims.Info["name"])
}

func TestRealtime_SSEHandler_ReturnsConnectionConfig(t *testing.T) {
	cfg := &realtime.Config{
		Enabled:      true,
		Secret:       "secret-key",
		ClientSSEURL: "http://localhost:8000/connection/sse",
	}
	mgr := realtime.NewManager(cfg, nil)
	mgr.SetAuthUserResolver(func(c fiber.Ctx) (realtime.User, error) {
		return realtime.User{ID: "200"}, nil
	})

	app := fiber.New()
	app.Get("/sse", mgr.SSEHandler())

	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body := testkit.JSONBody(t, resp)
	assert.Equal(t, "http://localhost:8000/connection/sse", body["url"])
	assert.Equal(t, "200", body["user_id"])
	assert.Contains(t, body, "token")
}
