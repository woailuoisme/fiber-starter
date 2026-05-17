package channels

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"fiber-starter/configs"
	contracts "fiber-starter/internal/providers/notification/contracts"

	"github.com/go-resty/resty/v2"
)

// TelegramChannel sends notifications to Telegram chats.
type TelegramChannel struct {
	client   *resty.Client
	botToken string
	chatID   string
}

// NewTelegramChannel creates a new TelegramChannel instance.
func NewTelegramChannel(cfg configs.TelegramNotificationConfig) (*TelegramChannel, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	apiURL := strings.TrimSpace(cfg.APIURL)
	botToken := strings.TrimSpace(cfg.BotToken)
	chatID := strings.TrimSpace(cfg.ChatID)
	if apiURL == "" {
		return nil, fmt.Errorf("telegram api_url is required when notification.telegram.enabled is true")
	}
	if botToken == "" {
		return nil, fmt.Errorf("telegram bot_token is required when notification.telegram.enabled is true")
	}
	if chatID == "" {
		return nil, fmt.Errorf("telegram chat_id is required when notification.telegram.enabled is true")
	}

	parsed, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse telegram api_url: %w", err)
	}

	client := resty.New().
		SetBaseURL(strings.TrimRight(parsed.String(), "/")).
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json")

	return &TelegramChannel{client: client, botToken: botToken, chatID: chatID}, nil
}

// SetClient replaces the internal resty.Client (primarily for testing).
func (c *TelegramChannel) SetClient(client *resty.Client) {
	c.client = client
}

type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// Send sends the notification via Telegram.
func (c *TelegramChannel) Send(notifiable interface{}, notification contracts.Notification) error {
	telegramNotification, ok := notification.(contracts.TelegramNotification)
	if !ok {
		return nil
	}

	message := telegramNotification.ToTelegram(notifiable)
	payload := contracts.TelegramMessage{
		ChatID:    strings.TrimSpace(message.ChatID),
		Text:      message.Text,
		ParseMode: strings.TrimSpace(message.ParseMode),
	}
	if payload.ChatID == "" {
		payload.ChatID = c.chatID
	}

	resp, err := c.client.R().
		SetBody(payload).
		Post("/bot" + c.botToken + "/sendMessage")
	if err != nil {
		return fmt.Errorf("send telegram notification: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("telegram request failed: status=%s body=%s", resp.Status(), strings.TrimSpace(resp.String()))
	}

	var result telegramResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}
	if !result.OK {
		desc := strings.TrimSpace(result.Description)
		if desc == "" {
			desc = strings.TrimSpace(resp.String())
		}
		return fmt.Errorf("telegram request rejected: %s", desc)
	}

	return nil
}
