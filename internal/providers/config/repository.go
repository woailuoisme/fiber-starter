package config

import (
	"sync"

	"github.com/knadh/koanf/v2"
)

// Repository implements the Config Repository contract using koanf.
type Repository struct {
	koanf *koanf.Koanf
	mu    sync.RWMutex
}

// NewRepository creates a new Repository instance.
func NewRepository(k *koanf.Koanf) *Repository {
	return &Repository{koanf: k}
}

// Get returns a configuration value by key.
func (r *Repository) Get(key string, defaultValue ...interface{}) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.koanf.Exists(key) {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return r.koanf.Get(key)
}

// LoadDatabaseConfig loads database configuration.
func (r *Repository) LoadDatabaseConfig(dbConfig interface{}) error {
	if err := r.koanf.Unmarshal("database", dbConfig); err != nil {
		return err
	}
	return nil
}

// Has checks if a configuration key exists.
func (r *Repository) Has(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.koanf.Exists(key)
}

// All returns all configuration values.
func (r *Repository) All() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.koanf.All()
}

// GetString returns a configuration value as string.
func (r *Repository) GetString(key string, defaultValue ...string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	val := r.koanf.String(key)
	if val == "" && len(defaultValue) > 0 && !r.koanf.Exists(key) {
		return defaultValue[0]
	}
	return val
}

// GetInt returns a configuration value as int.
func (r *Repository) GetInt(key string, defaultValue ...int) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.koanf.Exists(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return r.koanf.Int(key)
}

// GetBool returns a configuration value as bool.
func (r *Repository) GetBool(key string, defaultValue ...bool) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.koanf.Exists(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return r.koanf.Bool(key)
}

// GetFloat64 returns a configuration value as float64.
func (r *Repository) GetFloat64(key string, defaultValue ...float64) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.koanf.Exists(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return r.koanf.Float64(key)
}

// Set sets a configuration value.
func (r *Repository) Set(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = r.koanf.Set(key, value)
}

// Prepend prepends a value onto an array configuration value.
func (r *Repository) Prepend(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var arr []interface{}
	if r.koanf.Exists(key) {
		if v, ok := r.koanf.Get(key).([]interface{}); ok {
			arr = v
		}
	}
	_ = r.koanf.Set(key, append([]interface{}{value}, arr...))
}

// Push pushes a value onto an array configuration value.
func (r *Repository) Push(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var arr []interface{}
	if r.koanf.Exists(key) {
		if v, ok := r.koanf.Get(key).([]interface{}); ok {
			arr = v
		}
	}
	_ = r.koanf.Set(key, append(arr, value))
}
