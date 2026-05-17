package Contracts

// Notification defines the interface for a notification.
type Notification interface {
	// Via returns the channels the notification should be sent on.
	Via(notifiable interface{}) []string
}

// MailNotification defines the interface for notifications that can be sent via mail.
type MailNotification interface {
	Notification
	ToMail(notifiable interface{}) interface{} // Returns a mail message or data
}

// GotifyNotification defines the interface for notifications that can be sent via Gotify.
type GotifyNotification interface {
	Notification
	ToGotify(notifiable interface{}) GotifyMessage
}

// TelegramNotification defines the interface for notifications that can be sent via Telegram.
type TelegramNotification interface {
	Notification
	ToTelegram(notifiable interface{}) TelegramMessage
}

// DatabaseNotification defines the interface for notifications that can be stored in the database.
type DatabaseNotification interface {
	Notification
	ToDatabase(notifiable interface{}) map[string]interface{}
}

// Dispatcher defines the interface for sending notifications.
type Dispatcher interface {
	// Send sends the given notification to the given notifiable entities.
	Send(notifiables interface{}, notification Notification) error
}

// Channel defines the interface for a notification channel.
type Channel interface {
	// Send sends the given notification.
	Send(notifiable interface{}, notification Notification) error
}

// GotifyMessage represents a Gotify push payload.
type GotifyMessage struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// TelegramMessage represents a Telegram sendMessage payload.
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}
