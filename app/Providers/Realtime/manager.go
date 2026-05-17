package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	models "fiber-starter/app/Models"
	helpers "fiber-starter/app/Support"
	"fiber-starter/configs"

	"github.com/gofiber/contrib/v3/socketio"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type channelSubscriptionState struct {
	sub  Subscription
	refs int
}

// Manager coordinates sessions, channels, auth, and cluster fanout.
type Manager struct {
	cfg *configs.Config

	hub      *Hub
	bus      EventBus
	presence PresenceStore
	rdb      *redis.Client
	server   *Server
	nodeID   string

	mu            sync.Mutex
	subscriptions map[string]*channelSubscriptionState
	closed        atomic.Bool
}

func NewManager(cfg *configs.Config) *Manager {
	m := &Manager{
		cfg:           cfg,
		hub:           NewHub(),
		nodeID:        uuid.NewString(),
		subscriptions: make(map[string]*channelSubscriptionState),
	}
	m.initTransport()
	setActiveManager(m)
	m.server = NewServer(m)
	return m
}

func (m *Manager) initTransport() {
	if m.cfg == nil {
		m.presence = newMemoryPresenceStore()
		return
	}

	if strings.EqualFold(strings.TrimSpace(m.cfg.WebSocket.BusMode), "redis") {
		client := newRedisClient(m.cfg)
		m.rdb = client
		m.bus = newRedisBus(client, m.busPrefix())
		m.presence = newRedisPresenceStore(client, m.busPrefix())
		return
	}

	m.presence = newMemoryPresenceStore()
}

func (m *Manager) busPrefix() string {
	if m.cfg == nil || strings.TrimSpace(m.cfg.WebSocket.RedisPrefix) == "" {
		return "realtime"
	}
	return strings.TrimSpace(m.cfg.WebSocket.RedisPrefix)
}

func (m *Manager) Handler() fiber.Handler {
	if m.server == nil {
		m.server = NewServer(m)
	}
	return m.server.Handler()
}

func (m *Manager) AuthHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		req, err := parseAuthRequest(c)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		user := userFromContext(c)
		resp, err := BuildAuthResponse(m.cfg, req.SocketID, req.ChannelName, user)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}

		return c.JSON(resp)
	}
}

func (m *Manager) Close() error {
	if m == nil || m.closed.Swap(true) {
		return nil
	}

	clearActiveManager(m)

	m.mu.Lock()
	for channel, state := range m.subscriptions {
		if state != nil && state.sub != nil {
			_ = state.sub.Close()
		}
		delete(m.subscriptions, channel)
	}
	m.mu.Unlock()

	var errs []error
	if m.bus != nil {
		if err := m.bus.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.presence != nil {
		if err := m.presence.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.rdb != nil {
		if err := m.rdb.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *Manager) handler() *Server {
	if m.server == nil {
		m.server = NewServer(m)
	}
	return m.server
}

func (m *Manager) registerSession(session *Session) {
	if session == nil {
		return
	}
	m.hub.Register(session)
}

func (m *Manager) removeSession(sessionID string) {
	if sessionID == "" {
		return
	}

	session := m.hub.Session(sessionID)
	if session == nil {
		return
	}

	channels := session.ChannelNames()
	for _, channelName := range channels {
		m.unsubscribeSession(session, channelName)
	}
	m.hub.Unregister(sessionID)
	session.shutdown()
}

func (m *Manager) subscribeSession(session *Session, payload SubscribePayload) error {
	channel, err := ParseChannel(payload.Channel)
	if err != nil {
		return err
	}

	user := session.User()
	if channel.IsPrivate() {
		if payload.Auth != "" {
			if m.cfg == nil {
				return errors.New("realtime config unavailable")
			}
			if err := ValidateChannelAuth(m.cfg.WebSocket.AppKey, m.cfg.WebSocket.AppSecret, session.ID(), channel.Name, payload.Auth); err != nil {
				return err
			}
		} else if user == nil {
			return errors.New("authentication required for private channel")
		}
	}

	m.registerSession(session)
	m.hub.Join(session, channel.Name)
	m.ensureSubscription(channel.Name)

	switch channel.Kind {
	case ChannelPresence:
		member, err := m.presenceMember(session, payload)
		if err != nil {
			m.hub.Leave(session, channel.Name)
			m.releaseSubscription(channel.Name)
			return err
		}
		session.setPresenceMember(channel.Name, member)
		if err := m.presence.Join(context.Background(), channel.Name, session.ID(), member, m.presenceTTL()); err != nil {
			m.hub.Leave(session, channel.Name)
			m.releaseSubscription(channel.Name)
			return err
		}
		snapshot, _ := m.presenceSnapshot(channel.Name)
		_ = session.SendMessage(Message{
			Event:   EventSubscriptionSucceeded,
			Channel: channel.Name,
			Data:    encodeJSON(map[string]any{"presence": snapshot}),
		})
		m.publishEnvelope(Envelope{
			NodeID:         m.nodeID,
			Event:          EventMemberAdded,
			Channel:        channel.Name,
			Data:           encodeJSON(member),
			OriginSocketID: session.ID(),
		})
	default:
		_ = session.SendMessage(Message{
			Event:   EventSubscriptionSucceeded,
			Channel: channel.Name,
			Data:    encodeJSON(map[string]any{"channel": channel.Name}),
		})
	}

	return nil
}

func (m *Manager) unsubscribeSession(session *Session, channelName string) {
	if session == nil || channelName == "" {
		return
	}

	channel, err := ParseChannel(channelName)
	if err != nil {
		return
	}

	if !session.hasChannel(channelName) {
		return
	}

	if channel.Kind == ChannelPresence {
		if member, ok := session.presenceMember(channelName); ok {
			_ = m.presence.Leave(context.Background(), channelName, session.ID())
			m.publishEnvelope(Envelope{
				NodeID:         m.nodeID,
				Event:          EventMemberRemoved,
				Channel:        channelName,
				Data:           encodeJSON(member),
				OriginSocketID: session.ID(),
			})
		}
	}

	m.hub.Leave(session, channelName)
	m.releaseSubscription(channelName)
}

func (m *Manager) Dispatch(channelName, event string, data any) error {
	if channelName == "" || event == "" {
		return errors.New("missing channel or event")
	}
	m.publishEnvelope(Envelope{
		NodeID:  m.nodeID,
		Event:   event,
		Channel: channelName,
		Data:    encodeJSON(data),
	})
	return nil
}

func (m *Manager) publishEnvelope(env Envelope) {
	if env.Channel == "" || env.Event == "" {
		return
	}

	m.broadcastToChannel(env.Channel, Message{
		Event:    env.Event,
		Channel:  env.Channel,
		Data:     env.Data,
		SocketID: env.OriginSocketID,
	}, env.OriginSocketID)

	if m.bus == nil {
		return
	}

	raw, err := json.Marshal(env)
	if err != nil {
		return
	}
	if err := m.bus.Publish(context.Background(), env.Channel, raw); err != nil {
		m.logWarn("realtime_bus_publish_failed", zap.String("channel", env.Channel), zap.Error(err))
	}
}

func (m *Manager) handleBusMessage(channel string, payload []byte) {
	if len(payload) == 0 {
		return
	}

	var env Envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return
	}
	if env.Channel == "" {
		env.Channel = channel
	}
	if env.Channel == "" || env.Event == "" {
		return
	}

	if env.NodeID == m.nodeID {
		return
	}

	m.broadcastToChannel(env.Channel, Message{
		Event:    env.Event,
		Channel:  env.Channel,
		Data:     env.Data,
		SocketID: env.OriginSocketID,
	}, env.OriginSocketID)
}

func (m *Manager) broadcastToChannel(channelName string, msg Message, excludeSocketID string) {
	members := m.hub.Members(channelName)
	if len(members) == 0 {
		return
	}

	for _, session := range members {
		if session == nil {
			continue
		}
		if excludeSocketID != "" && session.ID() == excludeSocketID {
			continue
		}
		_ = session.SendMessage(msg)
	}
}

func (m *Manager) ensureSubscription(channel string) {
	if m.bus == nil || channel == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.subscriptions[channel]
	if state != nil {
		state.refs++
		return
	}

	sub, err := m.bus.Subscribe(context.Background(), channel)
	if err != nil {
		m.logWarn("realtime_bus_subscribe_failed", zap.String("channel", channel), zap.Error(err))
		return
	}

	state = &channelSubscriptionState{sub: sub, refs: 1}
	m.subscriptions[channel] = state
	go m.consumeSubscription(channel, sub)
}

func (m *Manager) consumeSubscription(channel string, sub Subscription) {
	for payload := range sub.Messages() {
		m.handleBusMessage(channel, payload)
	}
}

func (m *Manager) releaseSubscription(channel string) {
	if m.bus == nil || channel == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.subscriptions[channel]
	if state == nil {
		return
	}

	state.refs--
	if state.refs > 0 {
		return
	}

	_ = state.sub.Close()
	delete(m.subscriptions, channel)
}

func (m *Manager) presenceTTL() time.Duration {
	if m.cfg == nil || m.cfg.WebSocket.PresenceTTL <= 0 {
		return 2 * time.Minute
	}
	return time.Duration(m.cfg.WebSocket.PresenceTTL) * time.Second
}

func (m *Manager) presenceMember(session *Session, payload SubscribePayload) (PresenceMember, error) {
	if payload.ChannelData != "" {
		var member PresenceMember
		if err := json.Unmarshal([]byte(payload.ChannelData), &member); err != nil {
			return PresenceMember{}, err
		}
		if member.UserID == "" {
			return PresenceMember{}, errors.New("presence channel data missing user_id")
		}
		if member.UserInfo == nil {
			member.UserInfo = map[string]any{}
		}
		return member, nil
	}

	user := session.User()
	if user == nil {
		return PresenceMember{}, errors.New("presence member unavailable")
	}

	return PresenceMember{
		UserID: fmt.Sprintf("%d", user.ID),
		UserInfo: map[string]any{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	}, nil
}

func (m *Manager) presenceSnapshot(channel string) (PresenceSnapshot, error) {
	if m.presence == nil {
		return PresenceSnapshot{}, nil
	}
	members, err := m.presence.Members(context.Background(), channel)
	if err != nil {
		return PresenceSnapshot{}, err
	}
	channelInfo, _ := ParseChannel(channel)
	return channelInfo.Snapshot(members), nil
}

func (m *Manager) heartbeatIntervalSeconds() int {
	if m.cfg == nil || m.cfg.WebSocket.HeartbeatInterval <= 0 {
		return 30
	}
	return m.cfg.WebSocket.HeartbeatInterval
}

func (m *Manager) writeQueueSize() int {
	if m.cfg == nil || m.cfg.WebSocket.WriteQueueSize <= 0 {
		return 128
	}
	return m.cfg.WebSocket.WriteQueueSize
}

func (m *Manager) handleConnect(kws *socketio.Websocket) {
	if kws == nil {
		return
	}

	if m.cfg != nil && m.cfg.WebSocket.MaxMessageSize > 0 && kws.Conn != nil {
		kws.Conn.SetReadLimit(int64(m.cfg.WebSocket.MaxMessageSize))
	}

	session := newSession(m, kws)
	m.registerSession(session)
	kws.SetAttribute("realtime.socket_id", session.ID())
	kws.SetAttribute("realtime.session", session)

	session.Start()
	_ = session.SendMessage(Message{
		Event: EventConnectEstablished,
		Data: encodeJSON(ConnectionEstablishedData{
			SocketID:        session.ID(),
			ActivityTimeout: m.heartbeatIntervalSeconds(),
		}),
	})
	m.logInfo("realtime_connected", zap.String("socket_id", session.ID()))
}

func (m *Manager) handleDisconnect(payload *socketio.EventPayload) {
	if payload == nil {
		return
	}
	m.removeSession(payload.SocketUUID)
}

func (m *Manager) handleMessage(payload *socketio.EventPayload) {
	if payload == nil || len(payload.Data) == 0 {
		return
	}

	session := m.hub.Session(payload.SocketUUID)
	if session == nil {
		return
	}
	session.Inbound(payload.Data)
}

func (m *Manager) handlePong(payload *socketio.EventPayload) {
	if payload == nil {
		return
	}
	if session := m.hub.Session(payload.SocketUUID); session != nil {
		session.TouchPong()
	}
}

func (m *Manager) logInfo(msg string, fields ...zap.Field) {
	helpers.Info(msg, fields...)
}

func (m *Manager) logWarn(msg string, fields ...zap.Field) {
	helpers.Warn(msg, fields...)
}

func userFromContext(c fiber.Ctx) *models.User {
	if user, ok := c.Locals("user").(*models.User); ok {
		return user
	}
	return nil
}
