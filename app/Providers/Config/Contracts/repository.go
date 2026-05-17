package contracts

// Repository defines the interface for configuration repository.
type Repository interface {
	// Get returns a configuration value by key.
	Get(key string, defaultValue ...interface{}) interface{}

	// Has checks if a configuration key exists.
	Has(key string) bool

	// All returns all configuration values.
	All() map[string]interface{}

	// GetString returns a configuration value as string.
	GetString(key string, defaultValue ...string) string

	// GetInt returns a configuration value as int.
	GetInt(key string, defaultValue ...int) int

	// GetBool returns a configuration value as bool.
	GetBool(key string, defaultValue ...bool) bool

	// GetFloat64 returns a configuration value as float64.
	GetFloat64(key string, defaultValue ...float64) float64

	// Set sets a configuration value.
	Set(key string, value interface{})

	// Prepend prepends a value onto an array configuration value.
	Prepend(key string, value interface{})

	// Push pushes a value onto an array configuration value.
	Push(key string, value interface{})
}
