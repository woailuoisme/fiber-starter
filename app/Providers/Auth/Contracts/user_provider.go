package Contracts

import (
	models "fiber-starter/app/Models"
)

// UserProvider defines the contract for retrieving users for authentication (similar to Laravel's UserProvider)
type UserProvider interface {
	// RetrieveById retrieves a user by their unique identifier
	RetrieveById(id int64) (*models.User, error)

	// RetrieveByCredentials retrieves a user by the given credentials (e.g. email/password)
	RetrieveByCredentials(credentials map[string]string) (*models.User, error)

	// ValidateCredentials validates a user against the given credentials
	ValidateCredentials(user *models.User, credentials map[string]string) bool
}
