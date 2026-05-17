package cache

import (
	"errors"
	"time"

	"fiber-starter/internal/providers/cache/Contracts"
	"fiber-starter/internal/support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// store returns the default cache store from the container.
func store() Contracts.Store {
	if app := appctx.App(); app != nil {
		return app.CacheStore()
	}
	return nil
}

// GetStore returns a specific cache store
func GetStore(name ...string) (Contracts.Store, error) {
	if app := appctx.App(); app != nil {
		if manager := app.CacheManagerValue(); manager != nil {
			return manager.Store(name...), nil
		}
	}
	return nil, ErrContainerNotInitialized
}

// Global Facade methods for the default store

func Set(key string, value interface{}, expiration time.Duration) error {
	if s := store(); s != nil {
		return s.Set(key, value, expiration)
	}
	return ErrContainerNotInitialized
}

func Put(key string, value interface{}, expiration time.Duration) error {
	if s := store(); s != nil {
		return s.Put(key, value, expiration)
	}
	return ErrContainerNotInitialized
}

func Add(key string, value interface{}, expiration time.Duration) (bool, error) {
	if s := store(); s != nil {
		return s.Add(key, value, expiration)
	}
	return false, ErrContainerNotInitialized
}

func Forever(key string, value interface{}) error {
	if s := store(); s != nil {
		return s.Forever(key, value)
	}
	return ErrContainerNotInitialized
}

func Get(key string) (string, error) {
	if s := store(); s != nil {
		return s.Get(key)
	}
	return "", ErrContainerNotInitialized
}

func GetBytes(key string) ([]byte, error) {
	if s := store(); s != nil {
		return s.GetBytes(key)
	}
	return nil, ErrContainerNotInitialized
}

func GetJSON(key string, dest interface{}) error {
	if s := store(); s != nil {
		return s.GetJSON(key, dest)
	}
	return ErrContainerNotInitialized
}

func Delete(key string) error {
	if s := store(); s != nil {
		return s.Delete(key)
	}
	return ErrContainerNotInitialized
}

func Forget(key string) error {
	if s := store(); s != nil {
		return s.Forget(key)
	}
	return ErrContainerNotInitialized
}

func DeletePattern(pattern string) error {
	if s := store(); s != nil {
		return s.DeletePattern(pattern)
	}
	return ErrContainerNotInitialized
}

func Flush() error {
	if s := store(); s != nil {
		return s.Flush()
	}
	return ErrContainerNotInitialized
}

func Exists(key string) (bool, error) {
	if s := store(); s != nil {
		return s.Exists(key)
	}
	return false, ErrContainerNotInitialized
}

func Has(key string) (bool, error) {
	if s := store(); s != nil {
		return s.Has(key)
	}
	return false, ErrContainerNotInitialized
}

func Pull(key string) (string, error) {
	if s := store(); s != nil {
		return s.Pull(key)
	}
	return "", ErrContainerNotInitialized
}

func Increment(key string) (int64, error) {
	if s := store(); s != nil {
		return s.Increment(key)
	}
	return 0, ErrContainerNotInitialized
}

func Decrement(key string) (int64, error) {
	if s := store(); s != nil {
		return s.Decrement(key)
	}
	return 0, ErrContainerNotInitialized
}

func TTL(key string) (time.Duration, error) {
	if s := store(); s != nil {
		return s.TTL(key)
	}
	return 0, ErrContainerNotInitialized
}

func Expire(key string, expiration time.Duration) error {
	if s := store(); s != nil {
		return s.Expire(key, expiration)
	}
	return ErrContainerNotInitialized
}
