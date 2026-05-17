package contracts

import (
	"time"
)

// Store defines the contract for a cache driver (similar to Laravel's Cache Store)
type Store interface {
	// --- Basic Retrieval & Storage ---

	// Get retrieves an item from the cache by key
	Get(key string) (string, error)
	// GetBytes retrieves raw bytes from the cache
	GetBytes(key string) ([]byte, error)
	// GetJSON unmarshals a cached JSON string into a destination object
	GetJSON(key string, dest interface{}) error
	// Set stores an item in the cache for a given duration
	Set(key string, value interface{}, expiration time.Duration) error
	// Put is an alias for Set
	Put(key string, value interface{}, expiration time.Duration) error

	// --- Advanced Storage ---

	// Add stores an item only if it does not already exist
	Add(key string, value interface{}, expiration time.Duration) (bool, error)
	// Forever stores an item in the cache indefinitely
	Forever(key string, value interface{}) error

	// --- Removal & Existence ---

	// Delete removes an item from the cache
	Delete(key string) error
	// Forget is an alias for Delete
	Forget(key string) error
	// DeletePattern removes items matching a glob pattern
	DeletePattern(pattern string) error
	// Flush removes all items from the cache
	Flush() error
	// Exists determines if an item exists in the cache
	Exists(key string) (bool, error)
	// Has is an alias for Exists
	Has(key string) (bool, error)

	// --- Atomic & Helper Operations ---

	// Pull retrieves an item and then deletes it
	Pull(key string) (string, error)
	// Increment increments the value of an item in the cache
	Increment(key string) (int64, error)
	// Decrement decrements the value of an item in the cache
	Decrement(key string) (int64, error)
	// TTL returns the remaining time-to-live for a key
	TTL(key string) (time.Duration, error)
	// Expire sets a new expiration for an existing key
	Expire(key string, expiration time.Duration) error

	// HealthCheck verifies the cache connection is alive
	HealthCheck() error

	// Close closes the cache connection
	Close() error
}
