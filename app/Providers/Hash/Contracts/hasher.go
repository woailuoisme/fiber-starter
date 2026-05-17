package contracts

// Hasher defines the interface for hashing values.
type Hasher interface {
	// Make creates a hash for the given value.
	Make(value string) (string, error)

	// Check verifies the given value against a hash.
	Check(value, hashedValue string) bool

	// NeedsRehash checks if the given hash has been hashed using the given options.
	NeedsRehash(hashedValue string) bool

	// Info returns information about the given hashed value.
	Info(hashedValue string) map[string]interface{}
}
