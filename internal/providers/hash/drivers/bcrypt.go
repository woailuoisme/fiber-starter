package drivers

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements the Hasher interface using the bcrypt algorithm.
type BcryptHasher struct {
	rounds int
}

// NewBcryptHasher creates a new BcryptHasher instance.
func NewBcryptHasher(rounds int) *BcryptHasher {
	if rounds <= 0 {
		rounds = bcrypt.DefaultCost
	}
	return &BcryptHasher{rounds: rounds}
}

// Make creates a hash for the given value.
func (h *BcryptHasher) Make(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), h.rounds)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Check verifies the given value against a hash.
func (h *BcryptHasher) Check(value, hashedValue string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedValue), []byte(value))
	return err == nil
}

// NeedsRehash checks if the given hash has been hashed using the given options.
func (h *BcryptHasher) NeedsRehash(hashedValue string) bool {
	cost, err := bcrypt.Cost([]byte(hashedValue))
	if err != nil {
		return true
	}
	return cost != h.rounds
}

// Info returns information about the given hashed value.
func (h *BcryptHasher) Info(hashedValue string) map[string]interface{} {
	cost, _ := bcrypt.Cost([]byte(hashedValue))
	return map[string]interface{}{
		"algo":     "bcrypt",
		"algoName": "bcrypt",
		"options": map[string]interface{}{
			"rounds": cost,
		},
	}
}
