package presence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"lfiber/pkg/realtime/internal/pusher"

	"github.com/redis/go-redis/v9"
)

type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// PresenceStore 维护和查询在线状态成员的底层驱动接口
type PresenceStore interface {
	Join(ctx context.Context, channel string, socketID string, member pusher.PresenceMember, ttl time.Duration) error
	Leave(ctx context.Context, channel string, socketID string) error
	Members(ctx context.Context, channel string) ([]pusher.PresenceMember, error)
	Count(ctx context.Context, channel string) (int, error)
	Close() error
}

type memoryPresenceStore struct {
	mu       sync.RWMutex
	channels map[string]map[string]memoryPresenceMember
}

type memoryPresenceMember struct {
	member    pusher.PresenceMember
	expiresAt time.Time
}

func NewMemoryStore() PresenceStore {
	return &memoryPresenceStore{
		channels: make(map[string]map[string]memoryPresenceMember),
	}
}

func (s *memoryPresenceStore) Join(_ context.Context, channel string, socketID string, member pusher.PresenceMember, ttl time.Duration) error {
	if channel == "" || socketID == "" {
		return errors.New("invalid presence identity")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ch := s.channels[channel]
	if ch == nil {
		ch = make(map[string]memoryPresenceMember)
		s.channels[channel] = ch
	}
	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	ch[socketID] = memoryPresenceMember{member: member, expiresAt: expiresAt}
	return nil
}

func (s *memoryPresenceStore) Leave(_ context.Context, channel string, socketID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := s.channels[channel]
	if ch == nil {
		return nil
	}
	delete(ch, socketID)
	if len(ch) == 0 {
		delete(s.channels, channel)
	}
	return nil
}

func (s *memoryPresenceStore) Members(_ context.Context, channel string) ([]pusher.PresenceMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := s.channels[channel]
	if ch == nil {
		return nil, nil
	}

	now := time.Now()
	members := make([]pusher.PresenceMember, 0, len(ch))
	for socketID, item := range ch {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			delete(ch, socketID)
			continue
		}
		members = append(members, item.member)
	}
	if len(ch) == 0 {
		delete(s.channels, channel)
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].UserID < members[j].UserID
	})
	return members, nil
}

func (s *memoryPresenceStore) Count(ctx context.Context, channel string) (int, error) {
	members, err := s.Members(ctx, channel)
	if err != nil {
		return 0, err
	}
	return len(members), nil
}

func (s *memoryPresenceStore) Close() error { return nil }

// redisPresenceEnvelope 带有 NodeID 标识的 Redis 在线成员包装，用于在宕机时踢出残留数据
type RedisEnvelope struct {
	Member pusher.PresenceMember `json:"member"`
	NodeID string                `json:"node_id"`
}

type redisPresenceStore struct {
	client *redis.Client
	prefix string
	nodeID string
}

func NewRedisStore(client *redis.Client, prefix string, nodeID string) PresenceStore {
	return &redisPresenceStore{client: client, prefix: prefix, nodeID: nodeID}
}

func (s *redisPresenceStore) key(channel string) string {
	if s.prefix == "" {
		return "presence:" + channel
	}
	return fmt.Sprintf("%s:presence:%s", s.prefix, channel)
}

func (s *redisPresenceStore) Join(ctx context.Context, channel string, socketID string, member pusher.PresenceMember, ttl time.Duration) error {
	envelope := RedisEnvelope{
		Member: member,
		NodeID: s.nodeID,
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	key := s.key(channel)
	if err := s.client.HSet(ctx, key, socketID, raw).Err(); err != nil {
		return err
	}
	if ttl > 0 {
		return s.client.Expire(ctx, key, ttl).Err()
	}
	return nil
}

func (s *redisPresenceStore) Leave(ctx context.Context, channel string, socketID string) error {
	key := s.key(channel)
	if err := s.client.HDel(ctx, key, socketID).Err(); err != nil {
		return err
	}
	n, err := s.client.HLen(ctx, key).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return s.client.Del(ctx, key).Err()
	}
	return nil
}

func (s *redisPresenceStore) Members(ctx context.Context, channel string) ([]pusher.PresenceMember, error) {
	key := s.key(channel)
	values, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	members := make([]pusher.PresenceMember, 0, len(values))
	for _, raw := range values {
		var env RedisEnvelope
		if err := json.Unmarshal([]byte(raw), &env); err != nil {
			continue
		}
		members = append(members, env.Member)
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].UserID < members[j].UserID
	})
	return members, nil
}

func (s *redisPresenceStore) Count(ctx context.Context, channel string) (int, error) {
	key := s.key(channel)
	n, err := s.client.HLen(ctx, key).Result()
	return int(n), err
}

func (s *redisPresenceStore) Close() error { return nil }

// fallbackPresenceStore 当 Redis 不可用时自动降级为本地内存，提供高可用守护
type fallbackPresenceStore struct {
	redisStore PresenceStore
	memStore   PresenceStore
	redisAlive atomic.Bool
	logger     Logger
	client     *redis.Client
	done       chan struct{}
}

func NewFallbackStore(redisStore PresenceStore, logger Logger, client *redis.Client) PresenceStore {
	fps := &fallbackPresenceStore{
		redisStore: redisStore,
		memStore:   NewMemoryStore(),
		logger:     logger,
		client:     client,
		done:       make(chan struct{}),
	}
	fps.redisAlive.Store(true)
	go fps.detector()
	return fps
}

func (s *fallbackPresenceStore) Join(ctx context.Context, channel string, socketID string, member pusher.PresenceMember, ttl time.Duration) error {
	if s.redisAlive.Load() {
		err := s.redisStore.Join(ctx, channel, socketID, member, ttl)
		if err == nil {
			return nil
		}
		s.logger.Error("realtime_redis_presence_join_failed_falling_back", "channel", channel, "error", err.Error())
		s.redisAlive.Store(false)
	}
	return s.memStore.Join(ctx, channel, socketID, member, ttl)
}

func (s *fallbackPresenceStore) Leave(ctx context.Context, channel string, socketID string) error {
	if s.redisAlive.Load() {
		err := s.redisStore.Leave(ctx, channel, socketID)
		if err == nil {
			return nil
		}
		s.logger.Error("realtime_redis_presence_leave_failed_falling_back", "channel", channel, "error", err.Error())
		s.redisAlive.Store(false)
	}
	return s.memStore.Leave(ctx, channel, socketID)
}

func (s *fallbackPresenceStore) Members(ctx context.Context, channel string) ([]pusher.PresenceMember, error) {
	if s.redisAlive.Load() {
		members, err := s.redisStore.Members(ctx, channel)
		if err == nil {
			return members, nil
		}
		s.logger.Error("realtime_redis_presence_members_failed_falling_back", "channel", channel, "error", err.Error())
		s.redisAlive.Store(false)
	}
	return s.memStore.Members(ctx, channel)
}

func (s *fallbackPresenceStore) Count(ctx context.Context, channel string) (int, error) {
	if s.redisAlive.Load() {
		n, err := s.redisStore.Count(ctx, channel)
		if err == nil {
			return n, nil
		}
		s.logger.Error("realtime_redis_presence_count_failed_falling_back", "channel", channel, "error", err.Error())
		s.redisAlive.Store(false)
	}
	return s.memStore.Count(ctx, channel)
}

func (s *fallbackPresenceStore) detector() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			if s.redisAlive.Load() {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := s.client.Ping(ctx).Err()
			cancel()
			if err == nil {
				s.logger.Info("realtime_redis_presence_recovered_restoring_redis")
				s.redisAlive.Store(true)
			}
		}
	}
}

func (s *fallbackPresenceStore) Close() error {
	close(s.done)
	_ = s.redisStore.Close()
	_ = s.memStore.Close()
	return nil
}
