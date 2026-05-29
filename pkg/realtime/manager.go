package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type channelSubscriptionState struct {
	sub  Subscription
	refs int
}

// ManagerImpl 实时通信主控的实现，协调会话、通道注册、分布式广播和高可用心跳
type ManagerImpl struct {
	cfg *Config

	hub      *Hub
	bus      EventBus
	presence PresenceStore
	rdb      *redis.Client
	server   *Server
	nodeID   string
	logger   Logger
	registry *channelRegistry

	mu            sync.Mutex
	subscriptions map[string]*channelSubscriptionState
	closed        atomic.Bool
	done          chan struct{}

	// 解耦业务的回调挂载
	userResolver     func(*websocket.Conn) (User, error)
	onSubscribe      func(sessionID string, channel string, user User) error
	authUserResolver func(c fiber.Ctx) (User, error)
}

func NewManager(cfg *Config, logger Logger) *ManagerImpl {
	if logger == nil {
		logger = NewNoopLogger()
	}
	nodeID := uuid.NewString()

	m := &ManagerImpl{
		cfg:           cfg,
		logger:        logger,
		hub:           NewHub(),
		nodeID:        nodeID,
		registry:      newChannelRegistry(),
		subscriptions: make(map[string]*channelSubscriptionState),
		done:          make(chan struct{}),
	}

	m.initTransport()
	m.server = NewServer(m)

	// 开启集群节点心跳
	if m.rdb != nil {
		go m.nodeHeartbeatLoop()
	}

	return m
}

func (m *ManagerImpl) GetNodeID() string {
	return m.nodeID
}

func (m *ManagerImpl) Config() *Config {
	return m.cfg
}

func (m *ManagerImpl) SetUserResolver(resolver func(*websocket.Conn) (User, error)) {
	m.userResolver = resolver
}

func (m *ManagerImpl) SetOnSubscribe(hook func(sessionID string, channel string, user User) error) {
	m.onSubscribe = hook
}

func (m *ManagerImpl) AuthorizeChannel(pattern string, auth ChannelAuthorization) {
	if m.registry == nil {
		m.registry = newChannelRegistry()
	}
	m.registry.Register(pattern, auth)
}

func (m *ManagerImpl) SetAuthUserResolver(resolver func(fiber.Ctx) (User, error)) {
	m.authUserResolver = resolver
}

func (m *ManagerImpl) initTransport() {
	if m.cfg == nil {
		m.presence = newMemoryPresenceStore()
		return
	}

	var client *redis.Client
	if m.cfg.RedisClient != nil {
		client = m.cfg.RedisClient
	} else if strings.EqualFold(strings.TrimSpace(m.cfg.BusMode), "redis") {
		client = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	}

	if client != nil {
		m.rdb = client
		m.bus = newRedisBus(client, m.busPrefix(), m.logger)
		redisStore := newRedisPresenceStore(client, m.busPrefix(), m.nodeID)
		m.presence = newFallbackPresenceStore(redisStore, m.logger, client)
		return
	}

	m.presence = newMemoryPresenceStore()
}

func (m *ManagerImpl) busPrefix() string {
	if m.cfg == nil || strings.TrimSpace(m.cfg.RedisPrefix) == "" {
		return "realtime"
	}
	return strings.TrimSpace(m.cfg.RedisPrefix)
}

func (m *ManagerImpl) Handler() fiber.Handler {
	if m.server == nil {
		m.server = NewServer(m)
	}
	return m.server.Handler()
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

		if err := m.authorizeChannel(c.Context(), req.ChannelName, user); err != nil {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}

		resp, err := BuildAuthResponse(m.cfg, req.SocketID, req.ChannelName, user)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}

		return c.JSON(resp)
	}
}

func (m *ManagerImpl) APIHandler() fiber.Handler {
	return m.handleAPI
}

func (m *ManagerImpl) Close() error {
	if m == nil || m.closed.Swap(true) {
		return nil
	}

	close(m.done)

	// 1. 优雅平滑退出：通知客户端本节点即将关闭，下发 reconnect 事件引导其使用随机抖动时间重连
	var activeSessions []*Session
	if m.hub != nil {
		m.hub.mu.RLock()
		for _, sess := range m.hub.sessions {
			if sess != nil {
				activeSessions = append(activeSessions, sess)
			}
		}
		m.hub.mu.RUnlock()
	}

	if len(activeSessions) > 0 {
		m.logger.Info("realtime_graceful_shutdown_started", "active_sessions", len(activeSessions))
		var wg sync.WaitGroup
		for _, sess := range activeSessions {
			wg.Add(1)
			go func(s *Session) {
				defer wg.Done()
				// 产生 100 到 3000 毫秒的随机抖动延迟，以防止 thundering herd (雪崩重连) 拖垮后端
				//nolint:gosec // weak random is fine for reconnection jitter
				jitterMs := 100 + rand.Intn(2900)
				_ = s.SendMessage(Message{
					Event: "realtime:reconnect",
					Data:  encodePusherData(map[string]any{"reconnect_after_ms": jitterMs}),
				})
			}(sess)
		}

		// 等待事件分发排队，最多等待 500ms
		doneCh := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneCh)
		}()
		select {
		case <-doneCh:
		case <-time.After(500 * time.Millisecond):
		}

		// 给客户端 1s 时间（Connection Drain）使其能平滑建立新物理连接，随后断开
		time.Sleep(1 * time.Second)
	}

	// 2. 清理当前节点在 Redis 中的心跳
	if m.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = m.rdb.Del(ctx, m.heartbeatKey()).Err()
		cancel()
	}

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
	if m.rdb != nil && m.cfg.RedisClient == nil {
		if err := m.rdb.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *ManagerImpl) registerSession(session *Session) {
	if session == nil {
		return
	}
	m.hub.Register(session)
}

func (m *ManagerImpl) authorizeChannel(ctx context.Context, channel string, user User) error {
	if m.registry != nil {
		if err := m.registry.Authorize(ctx, user, channel); err != nil {
			return err
		}
	}
	if m.onSubscribe != nil {
		return m.onSubscribe("", channel, user)
	}
	return nil
}

func (m *ManagerImpl) removeSession(sessionID string) {
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

func (m *ManagerImpl) subscribeSession(session *Session, payload SubscribePayload) error {
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
			if err := ValidateChannelAuth(m.cfg.AppKey, m.cfg.AppSecret, session.ID(), channel.Name, payload.Auth, payload.ChannelData); err != nil {
				return err
			}
		} else {
			if err := m.authorizeChannel(context.Background(), channel.Name, user); err != nil {
				return err
			}
			if user.ID == "" {
				return errors.New("authentication required for private channel")
			}
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
			Data:    encodePusherData(snapshot.PusherData()),
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
			Data:    encodePusherData(map[string]any{}),
		})
	}

	return nil
}

func (m *ManagerImpl) unsubscribeSession(session *Session, channelName string) {
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

func (m *ManagerImpl) Dispatch(channelName, event string, data any) error {
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

func (m *ManagerImpl) publishEnvelope(env Envelope) {
	if env.Channel == "" || env.Event == "" {
		return
	}

	m.broadcastToChannel(env.Channel, Message{
		Event:    env.Event,
		Channel:  env.Channel,
		Data:     encodePusherRawData(env.Data),
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
		m.logger.Warn("realtime_bus_publish_failed", "channel", env.Channel, "error", err.Error())
	}
}

func (m *ManagerImpl) handleBusMessage(channel string, payload []byte) {
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
		Data:     encodePusherRawData(env.Data),
		SocketID: env.OriginSocketID,
	}, env.OriginSocketID)
}

func (m *ManagerImpl) broadcastToChannel(channelName string, msg Message, excludeSocketID string) {
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

func (m *ManagerImpl) ensureSubscription(channel string) {
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
		m.logger.Warn("realtime_bus_subscribe_failed", "channel", channel, "error", err.Error())
		return
	}

	state = &channelSubscriptionState{sub: sub, refs: 1}
	m.subscriptions[channel] = state
	go m.consumeSubscription(channel, sub)
}

func (m *ManagerImpl) consumeSubscription(channel string, sub Subscription) {
	for payload := range sub.Messages() {
		m.handleBusMessage(channel, payload)
	}
}

func (m *ManagerImpl) releaseSubscription(channel string) {
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

func (m *ManagerImpl) presenceTTL() time.Duration {
	if m.cfg == nil || m.cfg.PresenceTTL <= 0 {
		return 2 * time.Minute
	}
	return time.Duration(m.cfg.PresenceTTL) * time.Second
}

func (m *ManagerImpl) presenceMember(session *Session, payload SubscribePayload) (PresenceMember, error) {
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
	if user.ID == "" {
		return PresenceMember{}, errors.New("presence member unavailable")
	}

	return PresenceMember{
		UserID:   user.ID,
		UserInfo: user.Info,
	}, nil
}

func (m *ManagerImpl) presenceSnapshot(channel string) (PresenceSnapshot, error) {
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

func (m *ManagerImpl) heartbeatIntervalSeconds() int {
	if m.cfg == nil || m.cfg.HeartbeatInterval <= 0 {
		return 30
	}
	return m.cfg.HeartbeatInterval
}

func (m *ManagerImpl) writeQueueSize() int {
	if m.cfg == nil || m.cfg.WriteQueueSize <= 0 {
		return 128
	}
	return m.cfg.WriteQueueSize
}

func (m *ManagerImpl) clientEventsEnabled() bool {
	return false
}

func (m *ManagerImpl) handleConnect(conn *websocket.Conn) {
	if conn == nil {
		return
	}

	if m.cfg != nil && m.cfg.AppKey != "" {
		if appKey := conn.Params("appKey"); appKey != "" && appKey != m.cfg.AppKey {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid app key"))
			_ = conn.Close()
			return
		}
	}

	if m.cfg != nil && m.cfg.MaxMessageSize > 0 {
		conn.SetReadLimit(int64(m.cfg.MaxMessageSize))
	}

	session := newSession(m, conn)
	m.registerSession(session)

	_ = session.SendMessage(Message{
		Event: EventConnectEstablished,
		Data: encodePusherData(ConnectionEstablishedData{
			SocketID:        session.ID(),
			ActivityTimeout: m.heartbeatIntervalSeconds(),
		}),
	})
	m.logger.Info("realtime_connected", "socket_id", session.ID())

	session.Start()
}

// ---------------- 集群高可用节点监控与清理逻辑 ----------------

func (m *ManagerImpl) heartbeatKey() string {
	return fmt.Sprintf("%s:node:%s:heartbeat", m.busPrefix(), m.nodeID)
}

func (m *ManagerImpl) nodeHeartbeatLoop() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := m.rdb.Set(ctx, m.heartbeatKey(), "1", 60*time.Second).Err()
		cancel()
		if err != nil {
			m.logger.Error("realtime_node_heartbeat_failed", "error", err.Error())
		}

		select {
		case <-m.done:
			return
		case <-ticker.C:
		}
	}
}

// CleanupPresence 扫描 Redis 中所有 presence 频道，清理属于已宕机节点残留的 presence 数据并广播 MemberRemoved
func (m *ManagerImpl) CleanupPresence(ctx context.Context) error {
	if m.rdb == nil {
		return errors.New("redis client not configured")
	}

	heartbeatPattern := fmt.Sprintf("%s:node:*:heartbeat", m.busPrefix())
	nodeKeys, err := m.rdb.Keys(ctx, heartbeatPattern).Result()
	if err != nil {
		return err
	}

	activeNodes := make(map[string]bool)
	for _, key := range nodeKeys {
		parts := strings.Split(key, ":")
		if len(parts) >= 3 {
			nodeID := parts[2]
			activeNodes[nodeID] = true
		}
	}
	activeNodes[m.nodeID] = true

	presencePattern := fmt.Sprintf("%s:presence:*", m.busPrefix())
	presenceKeys, err := m.rdb.Keys(ctx, presencePattern).Result()
	if err != nil {
		return err
	}

	m.logger.Info("realtime_cleanup_presence_started", "active_nodes_count", len(activeNodes), "presence_channels_count", len(presenceKeys))

	for _, presenceKey := range presenceKeys {
		prefixParts := fmt.Sprintf("%s:presence:", m.busPrefix())
		channelName := strings.TrimPrefix(presenceKey, prefixParts)

		records, err := m.rdb.HGetAll(ctx, presenceKey).Result()
		if err != nil {
			m.logger.Error("realtime_cleanup_presence_hgetall_failed", "key", presenceKey, "error", err.Error())
			continue
		}

		for socketID, raw := range records {
			var env redisPresenceEnvelope
			if err := json.Unmarshal([]byte(raw), &env); err != nil {
				continue
			}

			if !activeNodes[env.NodeID] {
				m.logger.Warn("realtime_cleanup_presence_removing_dead_session", "channel", channelName, "socket_id", socketID, "dead_node_id", env.NodeID)

				if err := m.rdb.HDel(ctx, presenceKey, socketID).Err(); err != nil {
					m.logger.Error("realtime_cleanup_hdel_failed", "key", presenceKey, "socket_id", socketID, "error", err.Error())
					continue
				}

				m.publishEnvelope(Envelope{
					NodeID:         m.nodeID,
					Event:          EventMemberRemoved,
					Channel:        channelName,
					Data:           encodeJSON(env.Member),
					OriginSocketID: socketID,
				})
			}
		}

		n, err := m.rdb.HLen(ctx, presenceKey).Result()
		if err == nil && n == 0 {
			_ = m.rdb.Del(ctx, presenceKey).Err()
		}
	}

	return nil
}
