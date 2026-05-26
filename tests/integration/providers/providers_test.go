package providers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	providers "lfiber/internal/providers"
	notification "lfiber/internal/providers/notification"
	channels "lfiber/internal/providers/notification/channels"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	"lfiber/tests/internal/testkit"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvidersBuild(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	storageRoot := t.TempDir()

	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_SQLITE_DATABASE", cfg.Database.Connections["sqlite"].Database)
	t.Setenv("CACHE_DRIVER", "memory")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("STORAGE_LOCAL_ROOT", storageRoot)
	t.Setenv("STORAGE_LOCAL_URL", "/storage")
	t.Setenv("I18N_LANGUAGE_DIR", testkit.RepoRoot(t)+"/lang")

	t.Run("BuildRuntime", func(t *testing.T) {
		runtime, err := providers.Build()
		require.NoError(t, err)
		require.NotNil(t, runtime)

		// Verify all providers are initialized
		assert.NotNil(t, runtime.Connection)
		assert.NotNil(t, runtime.Cache)
		assert.NotNil(t, runtime.Storage)
		assert.NotNil(t, runtime.Log)
		assert.NotNil(t, runtime.RateLimiter)
		defer func() {
			_ = runtime.Close()
		}()
	})
}

func TestProvidersBuild_DoesNotConnectDependencies(t *testing.T) {
	t.Setenv("DB_CONNECTION", "pgsql")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "1")
	t.Setenv("DB_DATABASE", "missing")
	t.Setenv("DB_USERNAME", "postgres")
	t.Setenv("CACHE_DRIVER", "redis")
	t.Setenv("REDIS_HOST", "127.0.0.1")
	t.Setenv("REDIS_PORT", "1")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("STORAGE_LOCAL_ROOT", t.TempDir())
	t.Setenv("STORAGE_LOCAL_URL", "/storage")
	t.Setenv("I18N_LANGUAGE_DIR", testkit.RepoRoot(t)+"/lang")

	runtime, err := providers.Build()
	require.NoError(t, err)
	require.NotNil(t, runtime)
	defer func() {
		_ = runtime.Close()
	}()

	require.NotNil(t, runtime.Connection)
	require.NotNil(t, runtime.Cache)
}

func TestProvidersBuild_NotificationChannels(t *testing.T) {
	cfg := testkit.NewSQLiteConfig(t)
	storageRoot := t.TempDir()

	var gotifyHits atomic.Int32
	var telegramHits atomic.Int32

	t.Setenv("DB_CONNECTION", "sqlite")
	t.Setenv("DB_SQLITE_DATABASE", cfg.Database.Connections["sqlite"].Database)
	t.Setenv("CACHE_DRIVER", "memory")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("STORAGE_LOCAL_ROOT", storageRoot)
	t.Setenv("STORAGE_LOCAL_URL", "/storage")
	t.Setenv("I18N_LANGUAGE_DIR", testkit.RepoRoot(t)+"/lang")
	t.Setenv("NOTIFICATION_GOTIFY_ENABLED", "true")
	t.Setenv("NOTIFICATION_GOTIFY_URL", "http://gotify.test")
	t.Setenv("NOTIFICATION_GOTIFY_TOKEN", "secret-token")
	t.Setenv("NOTIFICATION_TELEGRAM_ENABLED", "true")
	t.Setenv("NOTIFICATION_TELEGRAM_API_URL", "http://telegram.test")
	t.Setenv("NOTIFICATION_TELEGRAM_BOT_TOKEN", "bot-token")
	t.Setenv("NOTIFICATION_TELEGRAM_CHAT_ID", "chat-id")

	runtime, err := providers.Build()
	require.NoError(t, err)
	require.NotNil(t, runtime)
	defer func() {
		_ = runtime.Close()
	}()

	manager := runtime.Notification.(*notification.Manager)
	gotifyChan := manager.Channel("gotify").(*channels.GotifyChannel)
	telegramChan := manager.Channel("telegram").(*channels.TelegramChannel)

	gotifyClient := resty.New().SetTransport(&mockRoundTripper{
		roundTrip: func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/message", r.URL.Path)
			assert.Equal(t, "secret-token", r.URL.Query().Get("token"))

			var payload notificationContracts.GotifyMessage
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			assert.Equal(t, "Build Failed", payload.Title)
			assert.Equal(t, "deployment failed", payload.Message)
			assert.Equal(t, 7, payload.Priority)

			gotifyHits.Add(1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
				Header:     make(http.Header),
			}, nil
		},
	})
	gotifyChan.SetClient(gotifyClient)

	telegramClient := resty.New().SetTransport(&mockRoundTripper{
		roundTrip: func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/botbot-token/sendMessage", r.URL.Path)

			var payload notificationContracts.TelegramMessage
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			assert.Equal(t, "override-chat", payload.ChatID)
			assert.Equal(t, "deployment failed", payload.Text)
			assert.Equal(t, "Markdown", payload.ParseMode)

			telegramHits.Add(1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true,"result":{}}`)),
				Header:     make(http.Header),
			}, nil
		},
	})
	telegramChan.SetClient(telegramClient)

	err = runtime.Notification.Send(notificationRecipient{}, notificationFanout{})
	require.NoError(t, err)

	assert.Equal(t, int32(1), gotifyHits.Load())
	assert.Equal(t, int32(1), telegramHits.Load())
}

type mockRoundTripper struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

type notificationRecipient struct{}

type notificationFanout struct{}

func (notificationFanout) Via(notifiable interface{}) []string {
	return []string{"gotify", "telegram"}
}

func (notificationFanout) ToGotify(notifiable interface{}) notificationContracts.GotifyMessage {
	return notificationContracts.GotifyMessage{
		Title:    "Build Failed",
		Message:  "deployment failed",
		Priority: 7,
	}
}

func (notificationFanout) ToTelegram(notifiable interface{}) notificationContracts.TelegramMessage {
	return notificationContracts.TelegramMessage{
		ChatID:    "override-chat",
		Text:      "deployment failed",
		ParseMode: "Markdown",
	}
}
