package realtime

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/centrifugal/gocent/v3"
	"github.com/gofiber/fiber/v3"
)

// ManagerImpl 实时通信主控的 Centrifugo 实现
type ManagerImpl struct {
	cfg      *Config
	client   *gocent.Client
	logger   Logger
	registry *channelRegistry

	// 回调挂载
	authUserResolver func(fiber.Ctx) (User, error)
}

func NewManager(cfg *Config, logger Logger) *ManagerImpl {
	if logger == nil {
		logger = NewNoopLogger()
	}

	var client *gocent.Client
	if cfg.Enabled && cfg.URL != "" {
		client = gocent.New(gocent.Config{
			Addr: cfg.URL,
			Key:  cfg.APIKey,
		})
	}

	return &ManagerImpl{
		cfg:      cfg,
		client:   client,
		logger:   logger,
		registry: newChannelRegistry(),
	}
}

func (m *ManagerImpl) Config() *Config {
	return m.cfg
}

func (m *ManagerImpl) GetNodeID() string {
	return "centrifugo"
}

func (m *ManagerImpl) SetAuthUserResolver(resolver func(fiber.Ctx) (User, error)) {
	m.authUserResolver = resolver
}

func (m *ManagerImpl) AuthorizeChannel(pattern string, auth ChannelAuthorization) {
	if m.registry == nil {
		m.registry = newChannelRegistry()
	}
	m.registry.Register(pattern, auth)
}

func (m *ManagerImpl) WebSocketHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		if m.authUserResolver == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "auth user resolver not configured")
		}

		user, err := m.authUserResolver(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		// 生成 Connection Token (有效期默认 24 小时)
		token, err := GenerateConnectionToken(m.cfg.Secret, user.ID, 86400, user.Info)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"url":     m.cfg.ClientURL,
			"token":   token,
			"user_id": user.ID,
		})
	}
}

func (m *ManagerImpl) SSEHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		if m.authUserResolver == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "auth user resolver not configured")
		}

		user, err := m.authUserResolver(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		// 生成 Connection Token (有效期默认 24 小时)
		token, err := GenerateConnectionToken(m.cfg.Secret, user.ID, 86400, user.Info)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"url":     m.cfg.ClientSSEURL,
			"token":   token,
			"user_id": user.ID,
		})
	}
}

func (m *ManagerImpl) AuthHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		req, err := parseAuthRequest(c)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if m.authUserResolver == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "auth user resolver not configured")
		}

		user, err := m.authUserResolver(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		// 权限校验
		if m.registry != nil {
			if err := m.registry.Authorize(c.Context(), user, req.Channel); err != nil {
				return fiber.NewError(fiber.StatusForbidden, err.Error())
			}
		}

		// 生成 Subscription Token (有效期默认 2 小时)
		token, err := GenerateSubscriptionToken(m.cfg.Secret, user.ID, req.Channel, 7200, user.Info)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(AuthResponse{
			Token: token,
		})
	}
}

func (m *ManagerImpl) Dispatch(channelName, event string, data any) error {
	if m.client == nil {
		return errors.New("centrifugo client not initialized")
	}
	if channelName == "" || event == "" {
		return errors.New("missing channel or event")
	}

	// 封装为 Pusher 风格的标准 Envelope
	payload, err := encodeJSONBytes(fiber.Map{
		"event": event,
		"data":  data,
	})
	if err != nil {
		return err
	}

	_, err = m.client.Publish(context.Background(), channelName, payload)
	return err
}

func encodeJSONBytes(v any) ([]byte, error) {
	return json.Marshal(v)
}
