package drivers

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"lfiber/configs"

	"github.com/gofiber/storage/ristretto/v2"
	"github.com/gofiber/storage/rueidis"
)

// RedisStore implements the Store interface using Rueidis (L2) and Ristretto (L1)
type RedisStore struct {
	l1       *ristretto.Storage
	l2       *rueidis.Storage
	prefix   string
	host     string
	port     string
	password string
	db       int
	mu       sync.Mutex
}

// NewRedisStore creates a new Redis store instance with tiered caching.
func NewRedisStore(cfg *configs.Config) (s *RedisStore) {
	host := cfg.Redis.Host
	port := cfg.Redis.Port

	l1 := ristretto.New(ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	return &RedisStore{
		l1:       l1,
		prefix:   cfg.Cache.Prefix,
		host:     host,
		port:     port,
		password: cfg.Redis.Password,
		db:       cfg.Redis.DB,
	}
}

func (s *RedisStore) client() (client *rueidis.Storage, err error) {
	if s == nil {
		return nil, errors.New("redis store not available")
	}
	if s.l2 != nil {
		return s.l2, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.l2 != nil {
		return s.l2, nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("initialize redis store: %v", r)
		}
	}()

	s.l2 = rueidis.New(rueidis.Config{
		InitAddress: []string{fmt.Sprintf("%s:%s", s.host, s.port)},
		Password:    s.password,
		SelectDB:    s.db,
	})
	return s.l2, nil
}

func (s *RedisStore) buildKey(key string) string {
	if s == nil {
		return key
	}
	return s.prefix + key
}

func (s *RedisStore) Get(key string) (string, error) {
	if s == nil {
		return "", errors.New("redis store not available")
	}
	cacheKey := s.buildKey(key)

	if s.l1 != nil {
		val, err := s.l1.Get(cacheKey)
		if err == nil && len(val) > 0 {
			return string(val), nil
		}
	}

	client, err := s.client()
	if err != nil {
		return "", err
	}

	val, err := client.Get(cacheKey)
	if err != nil {
		return "", err
	}

	if len(val) > 0 && s.l1 != nil {
		_ = s.l1.Set(cacheKey, val, 10*time.Minute)
	}

	return string(val), nil
}

func (s *RedisStore) GetBytes(key string) ([]byte, error) {
	val, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (s *RedisStore) GetJSON(key string, dest interface{}) error {
	val, err := s.Get(key)
	if err != nil {
		return err
	}
	if val == "" {
		return errors.New("key not found")
	}
	return json.Unmarshal([]byte(val), dest)
}

func (s *RedisStore) Set(key string, value interface{}, expiration time.Duration) error {
	if s == nil {
		return errors.New("redis store not available")
	}
	cacheKey := s.buildKey(key)

	var valBytes []byte
	switch v := value.(type) {
	case string:
		valBytes = []byte(v)
	case []byte:
		valBytes = v
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		valBytes = jsonBytes
	}

	client, err := s.client()
	if err != nil {
		return err
	}

	err = client.Set(cacheKey, valBytes, expiration)
	if err == nil && s.l1 != nil {
		_ = s.l1.Set(cacheKey, valBytes, expiration)
	}
	return err
}

func (s *RedisStore) Delete(key string) error {
	if s == nil {
		return nil
	}
	cacheKey := s.buildKey(key)
	if s.l1 != nil {
		_ = s.l1.Delete(cacheKey)
	}
	client, err := s.client()
	if err != nil {
		return err
	}
	return client.Delete(cacheKey)
}

func (s *RedisStore) Forget(key string) error {
	return s.Delete(key)
}

func (s *RedisStore) DeletePattern(pattern string) error {
	if s == nil {
		return nil
	}
	if s.l1 != nil {
		_ = s.l1.Reset()
	}
	return errors.New("DeletePattern not supported")
}

func (s *RedisStore) Flush() error {
	if s == nil {
		return nil
	}
	if s.l1 != nil {
		_ = s.l1.Reset()
	}
	client, err := s.client()
	if err != nil {
		return err
	}
	return client.Reset()
}

func (s *RedisStore) Exists(key string) (bool, error) {
	if s == nil {
		return false, nil
	}
	val, err := s.Get(key)
	if err != nil {
		return false, err
	}
	return val != "", nil
}

func (s *RedisStore) Has(key string) (bool, error) {
	return s.Exists(key)
}

func (s *RedisStore) Put(key string, value interface{}, expiration time.Duration) error {
	return s.Set(key, value, expiration)
}

func (s *RedisStore) Add(key string, value interface{}, expiration time.Duration) (bool, error) {
	exists, err := s.Exists(key)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	err = s.Set(key, value, expiration)
	return err == nil, err
}

func (s *RedisStore) Forever(key string, value interface{}) error {
	return s.Set(key, value, 0)
}

func (s *RedisStore) Pull(key string) (string, error) {
	val, err := s.Get(key)
	if err != nil {
		return "", err
	}
	if val != "" {
		_ = s.Delete(key)
	}
	return val, nil
}

func (s *RedisStore) TTL(key string) (time.Duration, error) {
	return 0, errors.New("TTL not supported via storage abstraction")
}

func (s *RedisStore) Expire(key string, expiration time.Duration) error {
	val, err := s.Get(key)
	if err != nil || val == "" {
		return err
	}
	return s.Set(key, val, expiration)
}

func (s *RedisStore) Increment(key string) (int64, error) {
	return 0, errors.New("Increment not supported via storage abstraction")
}

func (s *RedisStore) Decrement(key string) (int64, error) {
	return 0, errors.New("Decrement not supported via storage abstraction")
}

func (s *RedisStore) HealthCheck() error {
	client, err := s.client()
	if err != nil {
		return err
	}
	_, err = client.Get("__health_check__")
	return err
}

func (s *RedisStore) Close() error {
	if s == nil {
		return nil
	}
	if s.l1 != nil {
		_ = s.l1.Close()
	}
	if s.l2 != nil {
		return s.l2.Close()
	}
	return nil
}
