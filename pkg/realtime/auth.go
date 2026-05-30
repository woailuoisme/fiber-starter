package realtime

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type AuthRequest struct {
	Channel     string `json:"channel"`
	ChannelName string `json:"channel_name"` // 兼容 Pusher 字段名
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ConnectionClaims struct {
	jwt.RegisteredClaims
	Info map[string]any `json:"info,omitempty"`
}

type SubscriptionClaims struct {
	jwt.RegisteredClaims
	Channel string         `json:"channel"`
	Info    map[string]any `json:"info,omitempty"`
}

// GenerateConnectionToken 生成 Centrifugo 连接 JWT
func GenerateConnectionToken(secret string, userID string, ttlSeconds int64, info map[string]any) (string, error) {
	if secret == "" {
		return "", errors.New("centrifugo secret is empty")
	}

	now := time.Now()
	claims := ConnectionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		Info: info,
	}

	if ttlSeconds > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(ttlSeconds) * time.Second))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateSubscriptionToken 生成 Centrifugo 频道订阅 JWT
func GenerateSubscriptionToken(secret string, userID string, channel string, ttlSeconds int64, info map[string]any) (string, error) {
	if secret == "" {
		return "", errors.New("centrifugo secret is empty")
	}
	if channel == "" {
		return "", errors.New("channel is required")
	}

	now := time.Now()
	claims := SubscriptionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		Channel: channel,
		Info:    info,
	}

	if ttlSeconds > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(ttlSeconds) * time.Second))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func parseAuthRequest(c fiber.Ctx) (AuthRequest, error) {
	var req AuthRequest
	if err := c.Bind().JSON(&req); err != nil {
		// 尝试从 Form/Query 解析
		req.Channel = c.FormValue("channel")
		if req.Channel == "" {
			req.Channel = c.FormValue("channel_name")
		}
	}
	if req.Channel == "" && req.ChannelName != "" {
		req.Channel = req.ChannelName
	}
	if req.Channel == "" {
		return req, errors.New("missing channel parameter")
	}
	return req, nil
}
