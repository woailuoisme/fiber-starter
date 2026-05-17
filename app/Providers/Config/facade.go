package config

import (
	contracts "fiber-starter/app/Providers/Config/Contracts"
	"fiber-starter/app/Support/appctx"
)

// Facade provides a static interface to the configuration repository.
type Facade struct{}

// repo returns the configuration repository instance from the container.
func repo() contracts.Repository {
	if app := appctx.App(); app != nil {
		if rt, ok := app.(interface{ ConfigRepository() contracts.Repository }); ok {
			return rt.ConfigRepository()
		}
	}
	return nil
}

// Get returns a configuration value by key.
func Get(key string, defaultValue ...interface{}) interface{} {
	if r := repo(); r != nil {
		return r.Get(key, defaultValue...)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// Has checks if a configuration key exists.
func Has(key string) bool {
	if r := repo(); r != nil {
		return r.Has(key)
	}
	return false
}

// All returns all configuration values.
func All() map[string]interface{} {
	if r := repo(); r != nil {
		return r.All()
	}
	return make(map[string]interface{})
}

// GetString returns a configuration value as string.
func GetString(key string, defaultValue ...string) string {
	if r := repo(); r != nil {
		return r.GetString(key, defaultValue...)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetInt returns a configuration value as int.
func GetInt(key string, defaultValue ...int) int {
	if r := repo(); r != nil {
		return r.GetInt(key, defaultValue...)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetBool returns a configuration value as bool.
func GetBool(key string, defaultValue ...bool) bool {
	if r := repo(); r != nil {
		return r.GetBool(key, defaultValue...)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// GetFloat64 returns a configuration value as float64.
func GetFloat64(key string, defaultValue ...float64) float64 {
	if r := repo(); r != nil {
		return r.GetFloat64(key, defaultValue...)
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0.0
}

// Set sets a configuration value.
func Set(key string, value interface{}) {
	if r := repo(); r != nil {
		r.Set(key, value)
	}
}

// Prepend prepends a value onto an array configuration value.
func Prepend(key string, value interface{}) {
	if r := repo(); r != nil {
		r.Prepend(key, value)
	}
}

// Push pushes a value onto an array configuration value.
func Push(key string, value interface{}) {
	if r := repo(); r != nil {
		r.Push(key, value)
	}
}
