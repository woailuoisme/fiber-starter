package drivers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Hasher implements the Hasher interface using the Argon2id algorithm.
type Argon2Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// NewArgon2Hasher creates a new Argon2Hasher instance.
func NewArgon2Hasher(memory, iterations uint32, parallelism uint8) *Argon2Hasher {
	return &Argon2Hasher{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  16,
		keyLength:   32,
	}
}

// Make creates a hash for the given value.
func (h *Argon2Hasher) Make(value string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(value), salt, h.iterations, h.memory, h.parallelism, h.keyLength)

	// Format: $argon2id$v=19$m=65536,t=4,p=2$salt$hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.iterations, h.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// Check verifies the given value against a hash.
func (h *Argon2Hasher) Check(value, hashedValue string) bool {
	parts := strings.Split(hashedValue, "$")
	if len(parts) != 6 {
		return false
	}

	var version int
	var memory, iterations uint32
	var parallelism uint8

	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// gosec G115: check for overflow before conversion
	hashLenInt := len(hash)
	if hashLenInt <= 0 || hashLenInt > 1024 {
		return false
	}
	hashLen := uint32(hashLenInt)
	otherHash := argon2.IDKey([]byte(value), salt, iterations, memory, parallelism, hashLen)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1
}

// NeedsRehash checks if the given hash has been hashed using the given options.
func (h *Argon2Hasher) NeedsRehash(hashedValue string) bool {
	parts := strings.Split(hashedValue, "$")
	if len(parts) != 6 {
		return true
	}

	var memory, iterations uint32
	var parallelism uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return true
	}

	return memory != h.memory || iterations != h.iterations || parallelism != h.parallelism
}

// Info returns information about the given hashed value.
func (h *Argon2Hasher) Info(hashedValue string) map[string]interface{} {
	parts := strings.Split(hashedValue, "$")
	info := map[string]interface{}{
		"algo":     "argon2",
		"algoName": "argon2",
	}

	if len(parts) == 6 {
		var memory, iterations uint32
		var parallelism uint8
		_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
		if err == nil {
			info["options"] = map[string]interface{}{
				"memory":  memory,
				"time":    iterations,
				"threads": parallelism,
			}
		}
	}

	return info
}
