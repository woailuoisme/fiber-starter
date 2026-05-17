package Contracts

// UserProvider defines the contract for retrieving users for authentication (similar to Laravel's UserProvider)
type UserProvider interface {
	// RetrieveById retrieves a user by their unique identifier
	RetrieveById(id int64) (any, error)

	// RetrieveByCredentials retrieves a user by the given credentials (e.g. email/password)
	RetrieveByCredentials(credentials map[string]string) (any, error)

	// ValidateCredentials validates a user against the given credentials
	ValidateCredentials(user any, credentials map[string]string) bool
}
