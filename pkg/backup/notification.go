package backup

import (
	"fmt"
	"strings"

	mailProvider "lfiber/internal/providers/mail"
	notificationContracts "lfiber/internal/providers/notification/contracts"
)

type Notifiable struct {
	MailTo string
}

func (n Notifiable) RouteNotificationForMail() string {
	return n.MailTo
}

type Notification struct {
	Channels []string
	Title    string
	Message  string
	Priority int
}

func (n Notification) Via(_ interface{}) []string {
	return notificationChannels(n.Channels)
}

func (n Notification) ToMail(_ interface{}) interface{} {
	return mailProvider.NewMessage().
		Subject(n.Title).
		Plain(n.Message)
}

func (n Notification) ToGotify(_ interface{}) notificationContracts.GotifyMessage {
	return notificationContracts.GotifyMessage{Title: n.Title, Message: n.Message, Priority: n.Priority}
}

func (n Notification) ToTelegram(_ interface{}) notificationContracts.TelegramMessage {
	return notificationContracts.TelegramMessage{Text: fmt.Sprintf("*%s*\n%s", n.Title, n.Message), ParseMode: "Markdown"}
}

func notificationChannels(channels []string) []string {
	out := make([]string, 0, len(channels))
	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if channel != "" {
			out = append(out, channel)
		}
	}
	if len(out) == 0 {
		return []string{"mail"}
	}
	return out
}
