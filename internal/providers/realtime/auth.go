package realtime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"lfiber/configs"
	"lfiber/internal/features/user"

	"github.com/gofiber/fiber/v3"
)

// AuthRequest matches the broadcast auth request payload used by clients.
type AuthRequest struct {
	SocketID    string `json:"socket_id" form:"socket_id"`
	ChannelName string `json:"channel_name" form:"channel_name"`
}

// AuthResponse matches the Pusher auth response shape.
type AuthResponse struct {
	Auth        string `json:"auth"`
	ChannelData string `json:"channel_data,omitempty"`
}

func BuildAuthSignature(appKey, appSecret, socketID, channel string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write([]byte(socketID + ":" + channel))
	return appKey + ":" + hex.EncodeToString(mac.Sum(nil))
}

func BuildPresenceChannelData(user *user.User) string {
	if user == nil {
		return ""
	}

	payload := map[string]any{
		"user_id": strconv.FormatInt(user.ID, 10),
		"user_info": map[string]any{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(raw)
}

func ValidateChannelAuth(appKey, appSecret, socketID, channelName, auth string) error {
	expected := BuildAuthSignature(appKey, appSecret, socketID, channelName)
	if !hmac.Equal([]byte(expected), []byte(auth)) {
		return errors.New("invalid channel auth signature")
	}
	return nil
}

func isPrivateLike(channel Channel) bool {
	return channel.Kind == ChannelPrivate || channel.Kind == ChannelPresence
}

func BuildAuthResponse(cfg *configs.Config, socketID, channelName string, user *user.User) (AuthResponse, error) {
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
	if user == nil {
		return AuthResponse{}, errors.New("authenticated user required")
	}

	auth := BuildAuthSignature(cfg.WebSocket.AppKey, cfg.WebSocket.AppSecret, socketID, channelName)
	resp := AuthResponse{Auth: auth}
	if channel.IsPresence() {
		resp.ChannelData = BuildPresenceChannelData(user)
	}
	return resp, nil
}

func parseAuthRequest(c fiber.Ctx) (AuthRequest, error) {
	var req AuthRequest
	if err := c.Bind().Body(&req); err != nil {
		// Allow form/query based fallback for Pusher-style clients.
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
