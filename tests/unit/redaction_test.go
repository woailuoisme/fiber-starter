package tests

import (
	"testing"

	helpers "fiber-starter/internal/support"

	"github.com/stretchr/testify/assert"
)

func TestRedaction_RemovesTokensPasswordsAPIKeysAndConnectionStrings(t *testing.T) {
	samples := []string{
		"Authorization: Bearer abc.def.ghi",
		"password=plain-text",
		"api_key=secret-key",
		"token=secret-token",
		"postgres://user:pass@localhost:5432/app",
	}

	for _, sample := range samples {
		redacted := helpers.RedactSensitive(sample)
		assert.Contains(t, redacted, helpers.RedactionSentinel())
		assert.NotContains(t, redacted, "plain-text")
		assert.NotContains(t, redacted, "secret-key")
		assert.NotContains(t, redacted, "secret-token")
		assert.NotContains(t, redacted, "pass@")
	}
}

func TestRedaction_RedactsSensitiveHeaders(t *testing.T) {
	assert.Equal(t, helpers.RedactionSentinel(), helpers.RedactHeaderValue("Authorization", "Bearer secret"))
	assert.Equal(t, helpers.RedactionSentinel(), helpers.RedactHeaderValue("X-API-Key", "secret"))
	assert.Equal(t, "application/json", helpers.RedactHeaderValue("Content-Type", "application/json"))
}
