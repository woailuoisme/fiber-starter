package realtime

import (
	"errors"

	"lfiber/pkg/realtime/internal/pusher"

	"github.com/gofiber/fiber/v3"
)

type (
	AuthRequest  = pusher.AuthRequest
	AuthResponse = pusher.AuthResponse
)

func BuildAuthSignature(appKey, appSecret, socketID, channel string, channelData ...string) string {
	return pusher.BuildAuthSignature(appKey, appSecret, socketID, channel, channelData...)
}

func BuildPresenceChannelData(user User) string {
	return pusher.BuildPresenceChannelData(user)
}

func ValidateChannelAuth(appKey, appSecret, socketID, channelName, auth string, channelData ...string) error {
	return pusher.ValidateChannelAuth(appKey, appSecret, socketID, channelName, auth, channelData...)
}

func BuildAuthResponse(cfg *Config, socketID, channelName string, user User) (AuthResponse, error) {
	if cfg == nil {
		return AuthResponse{}, errors.New("realtime config unavailable")
	}
	return pusher.BuildAuthResponse(pusher.AuthConfig{
		AppKey:    cfg.AppKey,
		AppSecret: cfg.AppSecret,
	}, socketID, channelName, user)
}

func parseAuthRequest(c fiber.Ctx) (AuthRequest, error) {
	return pusher.ParseAuthRequest(c)
}
