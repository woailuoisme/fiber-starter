package realtime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// AuthRequest Pusher 风格的鉴权请求参数结构体
type AuthRequest struct {
	SocketID    string `json:"socket_id" form:"socket_id"`
	ChannelName string `json:"channel_name" form:"channel_name"`
}

// AuthResponse Pusher 风格的广播授权响应结构体
type AuthResponse struct {
	Auth        string `json:"auth"`
	ChannelData string `json:"channel_data,omitempty"`
}

func BuildAuthSignature(appKey, appSecret, socketID, channel string, channelData ...string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	payload := socketID + ":" + channel
	if len(channelData) > 0 && channelData[0] != "" {
		payload += ":" + channelData[0]
	}
	_, _ = mac.Write([]byte(payload))
	return appKey + ":" + hex.EncodeToString(mac.Sum(nil))
}

func BuildPresenceChannelData(user User) string {
	if user.ID == "" {
		return ""
	}

	payload := map[string]any{
		"user_id":   user.ID,
		"user_info": user.Info,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(raw)
}

func ValidateChannelAuth(appKey, appSecret, socketID, channelName, auth string, channelData ...string) error {
	expected := BuildAuthSignature(appKey, appSecret, socketID, channelName, channelData...)
	if !hmac.Equal([]byte(expected), []byte(auth)) {
		return errors.New("invalid channel auth signature")
	}
	return nil
}

func isPrivateLike(channel Channel) bool {
	return channel.Kind == ChannelPrivate || channel.Kind == ChannelPresence
}

func BuildAuthResponse(cfg *Config, socketID, channelName string, user User) (AuthResponse, error) {
	if cfg == nil {
		return AuthResponse{}, errors.New("realtime config unavailable")
	}

	channel, err := ParseChannel(channelName)
	if err != nil {
		return AuthResponse{}, err
	}
	if !isPrivateLike(channel) {
		return AuthResponse{}, errors.New("public channels do not require auth")
	}
	if user.ID == "" {
		return AuthResponse{}, errors.New("authenticated user required")
	}

	resp := AuthResponse{}
	if channel.IsPresence() {
		resp.ChannelData = BuildPresenceChannelData(user)
	}
	resp.Auth = BuildAuthSignature(cfg.AppKey, cfg.AppSecret, socketID, channelName, resp.ChannelData)
	return resp, nil
}

func parseAuthRequest(c fiber.Ctx) (AuthRequest, error) {
	var req AuthRequest
	if err := c.Bind().Body(&req); err != nil {
		req.SocketID = strings.TrimSpace(c.FormValue("socket_id"))
		req.ChannelName = strings.TrimSpace(c.FormValue("channel_name"))
		if req.SocketID == "" || req.ChannelName == "" {
			req.SocketID = strings.TrimSpace(c.Query("socket_id"))
			req.ChannelName = strings.TrimSpace(c.Query("channel_name"))
		}
		if req.SocketID == "" || req.ChannelName == "" {
			return AuthRequest{}, err
		}
	}
	req.SocketID = strings.TrimSpace(req.SocketID)
	req.ChannelName = strings.TrimSpace(req.ChannelName)
	if req.SocketID == "" || req.ChannelName == "" {
		return AuthRequest{}, errors.New("missing socket_id or channel_name")
	}
	return req, nil
}
