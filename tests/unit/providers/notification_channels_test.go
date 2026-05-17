package providers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"fiber-starter/configs"
	channels "fiber-starter/internal/providers/notification/channels"
	notificationContracts "fiber-starter/internal/providers/notification/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type gotifyNotification struct{}

func (gotifyNotification) Via(notifiable interface{}) []string {
	return []string{"gotify"}
}

func (gotifyNotification) ToGotify(notifiable interface{}) notificationContracts.GotifyMessage {
	return notificationContracts.GotifyMessage{
		Title:    "Build Failed",
		Message:  "deployment failed",
		Priority: 7,
	}
}

type telegramNotification struct{}

func (telegramNotification) Via(notifiable interface{}) []string {
	return []string{"telegram"}
}

func (telegramNotification) ToTelegram(notifiable interface{}) notificationContracts.TelegramMessage {
	return notificationContracts.TelegramMessage{
		ChatID:    "override-chat",
		Text:      "deployment failed",
		ParseMode: "Markdown",
	}
}

type unsupportedNotification struct{}

func (unsupportedNotification) Via(notifiable interface{}) []string {
	return []string{"gotify"}
}

func TestGotifyChannel_Send(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotToken string
	var gotPayload notificationContracts.GotifyMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.URL.Query().Get("token")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	channel, err := channels.NewGotifyChannel(configs.GotifyNotificationConfig{
		Enabled:  true,
		URL:      server.URL,
		Token:    "secret-token",
		Title:    "Fiber Starter",
		Priority: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, channel)

	err = channel.Send(struct{}{}, gotifyNotification{})
	require.NoError(t, err)
	assert.Equal(t, "/message", gotPath)
	assert.Equal(t, "secret-token", gotToken)
	assert.Equal(t, "Build Failed", gotPayload.Title)
	assert.Equal(t, "deployment failed", gotPayload.Message)
	assert.Equal(t, 7, gotPayload.Priority)
}

func TestTelegramChannel_Send(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotPayload notificationContracts.TelegramMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	channel, err := channels.NewTelegramChannel(configs.TelegramNotificationConfig{
		Enabled:   true,
		APIURL:    server.URL,
		BotToken:  "bot-token",
		ChatID:    "chat-id",
		ParseMode: "",
	})
	require.NoError(t, err)
	require.NotNil(t, channel)

	err = channel.Send(struct{}{}, telegramNotification{})
	require.NoError(t, err)
	assert.Equal(t, "/botbot-token/sendMessage", gotPath)
	assert.Equal(t, "override-chat", gotPayload.ChatID)
	assert.Equal(t, "deployment failed", gotPayload.Text)
	assert.Equal(t, "Markdown", gotPayload.ParseMode)
}

func TestTelegramChannel_RequestRejected(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer server.Close()

	channel, err := channels.NewTelegramChannel(configs.TelegramNotificationConfig{
		Enabled:  true,
		APIURL:   server.URL,
		BotToken: "bot-token",
		ChatID:   "chat-id",
	})
	require.NoError(t, err)

	err = channel.Send(struct{}{}, telegramNotification{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chat not found")
}

func TestNotificationChannels_NoOpForUnsupportedNotification(t *testing.T) {
	t.Parallel()

	gotifyChannel, err := channels.NewGotifyChannel(configs.GotifyNotificationConfig{
		Enabled: true,
		URL:     "http://example.invalid",
		Token:   "token",
	})
	require.NoError(t, err)

	telegramChannel, err := channels.NewTelegramChannel(configs.TelegramNotificationConfig{
		Enabled:  true,
		APIURL:   "http://example.invalid",
		BotToken: "token",
		ChatID:   "chat-id",
	})
	require.NoError(t, err)

	assert.NoError(t, gotifyChannel.Send(struct{}{}, unsupportedNotification{}))
	assert.NoError(t, telegramChannel.Send(struct{}{}, unsupportedNotification{}))
}

func TestNotificationChannels_InvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := channels.NewGotifyChannel(configs.GotifyNotificationConfig{
		Enabled: true,
		Token:   "token",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gotify url is required")

	_, err = channels.NewTelegramChannel(configs.TelegramNotificationConfig{
		Enabled:  true,
		APIURL:   "http://example.com",
		BotToken: "",
		ChatID:   "chat-id",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "telegram bot_token is required")
}

func TestGotifyChannel_RequestFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"error":"boom"}`)
	}))
	defer server.Close()

	channel, err := channels.NewGotifyChannel(configs.GotifyNotificationConfig{
		Enabled: true,
		URL:     server.URL,
		Token:   "secret-token",
	})
	require.NoError(t, err)

	err = channel.Send(struct{}{}, gotifyNotification{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gotify request failed")
}
