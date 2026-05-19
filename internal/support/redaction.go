package support

import (
	"regexp"
	"strings"
)

const redactedValue = "<redacted>"

var (
	credentialInURLPattern = regexp.MustCompile(`(?i)(://[^:/\s]+:)[^@\s]+(@)`)
	keyValueSecretPattern  = regexp.MustCompile(`(?i)\b(token|password|passwd|secret|api[_-]?key|access[_-]?key|private[_-]?key|dsn|connection[_-]?string)=([^&\s,;]+)`)
	authBearerPattern      = regexp.MustCompile(`(?i)\b(bearer\s+)[A-Za-z0-9._~+/=-]+`)
)

// RedactSensitive removes common credential material from log and error strings.
func RedactSensitive(value string) string {
	value = credentialInURLPattern.ReplaceAllString(value, "${1}"+redactedValue+"${2}")
	value = keyValueSecretPattern.ReplaceAllString(value, "${1}="+redactedValue)
	value = authBearerPattern.ReplaceAllString(value, "${1}"+redactedValue)
	return value
}

// RedactError returns a sanitized error message suitable for readiness payloads.
func RedactError(err error) string {
	if err == nil {
		return ""
	}
	return RedactSensitive(err.Error())
}

// RedactHeaderValue hides values for headers that commonly contain credentials.
func RedactHeaderValue(key, value string) string {
	if value == "" {
		return value
	}

	name := strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{"authorization", "cookie", "set-cookie", "x-api-key", "x-auth-token", "proxy-authorization"} {
		if name == token {
			return redactedValue
		}
	}
	if strings.Contains(name, "token") || strings.Contains(name, "secret") || strings.Contains(name, "password") || strings.Contains(name, "key") {
		return redactedValue
	}

	return RedactSensitive(value)
}

// IsRedacted reports whether a sanitized value had sensitive material removed.
func IsRedacted(value string) bool {
	return strings.Contains(value, redactedValue)
}

// RedactionSentinel exposes the stable redaction marker for tests.
func RedactionSentinel() string {
	return redactedValue
}
