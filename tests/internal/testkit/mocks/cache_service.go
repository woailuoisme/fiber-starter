package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// CacheService is a testify mock for the cache boundary.
type CacheService struct {
	mock.Mock
}

func (m *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *CacheService) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *CacheService) GetBytes(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *CacheService) GetJSON(key string, dest interface{}) error {
	args := m.Called(key, dest)
	return args.Error(0)
}

func (m *CacheService) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *CacheService) DeletePattern(pattern string) error {
	args := m.Called(pattern)
	return args.Error(0)
}

func (m *CacheService) Exists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *CacheService) TTL(key string) (time.Duration, error) {
	args := m.Called(key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *CacheService) Expire(key string, expiration time.Duration) error {
	args := m.Called(key, expiration)
	return args.Error(0)
}

func (m *CacheService) Increment(key string) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *CacheService) Decrement(key string) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *CacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}
