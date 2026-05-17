package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// PresenceStore tracks presence members per channel.
type PresenceStore interface {
	Join(ctx context.Context, channel string, socketID string, member PresenceMember, ttl time.Duration) error
	Leave(ctx context.Context, channel string, socketID string) error
	Members(ctx context.Context, channel string) ([]PresenceMember, error)
	Count(ctx context.Context, channel string) (int, error)
	Close() error
}

type memoryPresenceStore struct {
	mu       sync.RWMutex
	channels map[string]map[string]memoryPresenceMember
}

type memoryPresenceMember struct {
	member    PresenceMember
	expiresAt time.Time
}

func newMemoryPresenceStore() PresenceStore {
	return &memoryPresenceStore{
		channels: make(map[string]map[string]memoryPresenceMember),
	}
}

func (s *memoryPresenceStore) Join(_ context.Context, channel string, socketID string, member PresenceMember, ttl time.Duration) error {
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

func (s *memoryPresenceStore) Members(_ context.Context, channel string) ([]PresenceMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := s.channels[channel]
	if ch == nil {
		return nil, nil
	}

	now := time.Now()
	members := make([]PresenceMember, 0, len(ch))
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

type redisPresenceStore struct {
	client *redis.Client
	prefix string
}

func newRedisPresenceStore(client *redis.Client, prefix string) PresenceStore {
	return &redisPresenceStore{client: client, prefix: prefix}
}

func (s *redisPresenceStore) key(channel string) string {
	if s.prefix == "" {
		return "presence:" + channel
	}
	return fmt.Sprintf("%s:presence:%s", s.prefix, channel)
}

func (s *redisPresenceStore) Join(ctx context.Context, channel string, socketID string, member PresenceMember, ttl time.Duration) error {
	raw, err := json.Marshal(member)
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

func (s *redisPresenceStore) Members(ctx context.Context, channel string) ([]PresenceMember, error) {
	key := s.key(channel)
	values, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	members := make([]PresenceMember, 0, len(values))
	for _, raw := range values {
		var member PresenceMember
		if err := json.Unmarshal([]byte(raw), &member); err != nil {
			continue
		}
		members = append(members, member)
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
