package channels

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"fiber-starter/configs"
	contracts "fiber-starter/internal/providers/notification/contracts"

	"github.com/go-resty/resty/v2"
)

// GotifyChannel sends notifications to a Gotify instance.
type GotifyChannel struct {
	client *resty.Client
	token  string
}

// NewGotifyChannel creates a new GotifyChannel instance.
func NewGotifyChannel(cfg configs.GotifyNotificationConfig) (*GotifyChannel, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	baseURL := strings.TrimSpace(cfg.URL)
	token := strings.TrimSpace(cfg.Token)
	if baseURL == "" {
		return nil, fmt.Errorf("gotify url is required when notification.gotify.enabled is true")
	}
	if token == "" {
		return nil, fmt.Errorf("gotify token is required when notification.gotify.enabled is true")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse gotify url: %w", err)
	}

	client := resty.New().
		SetBaseURL(strings.TrimRight(parsed.String(), "/")).
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json")

	return &GotifyChannel{client: client, token: token}, nil
}

// Send sends the notification via Gotify.
func (c *GotifyChannel) Send(notifiable interface{}, notification contracts.Notification) error {
	gotifyNotification, ok := notification.(contracts.GotifyNotification)
	if !ok {
		return nil
	}

	message := gotifyNotification.ToGotify(notifiable)
	payload := contracts.GotifyMessage{
		Title:    strings.TrimSpace(message.Title),
		Message:  message.Message,
		Priority: message.Priority,
	}

	if payload.Title == "" {
		payload.Title = "Fiber Starter"
	}

	resp, err := c.client.R().
		SetQueryParam("token", c.token).
		SetBody(payload).
		Post("/message")
	if err != nil {
		return fmt.Errorf("send gotify notification: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("gotify request failed: status=%s body=%s", resp.Status(), strings.TrimSpace(resp.String()))
	}

	return nil
}
