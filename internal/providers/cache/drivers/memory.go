package drivers

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/storage/ristretto/v2"
)

type MemoryStore struct {
	storage *ristretto.Storage
	prefix  string
}

func NewMemoryStore(prefix string) *MemoryStore {
	store := ristretto.New(ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	return &MemoryStore{
		storage: store,
		prefix:  prefix,
	}
}

func (s *MemoryStore) buildKey(key string) string {
	return s.prefix + key
}

func (s *MemoryStore) Get(key string) (string, error) {
	val, err := s.storage.Get(s.buildKey(key))
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func (s *MemoryStore) GetBytes(key string) ([]byte, error) {
	return s.storage.Get(s.buildKey(key))
}

func (s *MemoryStore) GetJSON(key string, dest interface{}) error {
	val, err := s.Get(key)
	if err != nil {
		return err
	}
	if val == "" {
		return errors.New("key not found")
	}
	return json.Unmarshal([]byte(val), dest)
}

func (s *MemoryStore) Set(key string, value interface{}, expiration time.Duration) error {
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
	return s.storage.Set(s.buildKey(key), valBytes, expiration)
}

func (s *MemoryStore) Delete(key string) error {
	return s.storage.Delete(s.buildKey(key))
}

func (s *MemoryStore) Forget(key string) error {
	return s.Delete(key)
}

func (s *MemoryStore) DeletePattern(pattern string) error {
	return s.storage.Reset() // Ristretto doesn't support pattern delete easily
}

func (s *MemoryStore) Flush() error {
	return s.storage.Reset()
}

func (s *MemoryStore) Exists(key string) (bool, error) {
	val, err := s.Get(key)
	if err != nil {
		return false, err
	}
	return val != "", nil
}

func (s *MemoryStore) Has(key string) (bool, error) {
	return s.Exists(key)
}

func (s *MemoryStore) Put(key string, value interface{}, expiration time.Duration) error {
	return s.Set(key, value, expiration)
}

func (s *MemoryStore) Add(key string, value interface{}, expiration time.Duration) (bool, error) {
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

func (s *MemoryStore) Forever(key string, value interface{}) error {
	return s.Set(key, value, 0)
}

func (s *MemoryStore) Pull(key string) (string, error) {
	val, err := s.Get(key)
	if err != nil {
		return "", err
	}
	if val != "" {
		_ = s.Delete(key)
	}
	return val, nil
}

func (s *MemoryStore) TTL(key string) (time.Duration, error) {
	return 0, errors.New("TTL not supported by memory driver")
}

func (s *MemoryStore) Expire(key string, expiration time.Duration) error {
	val, err := s.Get(key)
	if err != nil || val == "" {
		return err
	}
	return s.Set(key, val, expiration)
}

func (s *MemoryStore) Increment(key string) (int64, error) {
	return 0, errors.New("Increment not supported by memory driver")
}

func (s *MemoryStore) Decrement(key string) (int64, error) {
	return 0, errors.New("Decrement not supported by memory driver")
}

func (s *MemoryStore) HealthCheck() error {
	return nil
}

func (s *MemoryStore) Close() error {
	return s.storage.Close()
}

func (s *MemoryStore) Wait() {
	s.storage.Conn().Wait()
}
