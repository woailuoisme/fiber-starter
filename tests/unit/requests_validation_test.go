package tests

import (
	"testing"

	"lfiber/internal/common/exceptions"
	requests "lfiber/internal/common/requests"
	"lfiber/internal/features/auth"
	"lfiber/internal/features/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestValidationHelpers_UseValidatorRules(t *testing.T) {
	assert.True(t, requests.ValidateEmail("user@example.com"))
	assert.False(t, requests.ValidateEmail("invalid"))

	assert.True(t, requests.ValidateURL("https://example.com"))
	assert.False(t, requests.ValidateURL("not-a-url"))

	assert.True(t, requests.ValidateE164("+8613800138000"))
	assert.False(t, requests.ValidateE164("13800138000"))

	assert.True(t, requests.ValidatePassword("password123"))
	assert.False(t, requests.ValidatePassword("short"))

	assert.True(t, requests.ValidateRequired("hello"))
	assert.False(t, requests.ValidateRequired("   "))

	assert.True(t, requests.ValidateMinLength("hello", 3))
	assert.False(t, requests.ValidateMinLength("hi", 3))

	assert.True(t, requests.ValidateMaxLength("hi", 3))
	assert.False(t, requests.ValidateMaxLength("hello", 3))

	assert.True(t, requests.ValidateRange(5, 1, 10))
	assert.False(t, requests.ValidateRange(0, 1, 10))

	assert.True(t, requests.ValidatePositiveNumber(1))
	assert.False(t, requests.ValidatePositiveNumber(0))

	assert.True(t, requests.ValidatePositiveInteger(1))
	assert.False(t, requests.ValidatePositiveInteger(0))
}

func TestRequestStructTags_EnforceAuthAndUserRules(t *testing.T) {
	registerErr := requests.ValidateStruct(&auth.RegisterRequest{
		Name:     "A",
		Email:    "not-an-email",
		Password: "short",
	})
	require.Error(t, registerErr)
	var valRegErr *exceptions.ValidationException
	require.ErrorAs(t, registerErr, &valRegErr)
	registerErrors := valRegErr.Errors
	assert.Contains(t, registerErrors, "name")
	assert.Contains(t, registerErrors, "email")
	assert.Contains(t, registerErrors, "password")

	profileErr := requests.ValidateStruct(&user.UpdateProfileRequest{
		Name:   "A",
		Phone:  "not-a-phone",
		Avatar: "not-a-url",
	})
	require.Error(t, profileErr)
	var valProfileErr *exceptions.ValidationException
	require.ErrorAs(t, profileErr, &valProfileErr)
	profileErrors := valProfileErr.Errors
	assert.Contains(t, profileErrors, "name")
	assert.Contains(t, profileErrors, "phone")
	assert.Contains(t, profileErrors, "avatar")

	searchErr := requests.ValidateStruct(&user.SearchUsersRequest{
		Q:     "",
		Page:  -1,
		Limit: 101,
	})
	require.Error(t, searchErr)
	var valSearchErr *exceptions.ValidationException
	require.ErrorAs(t, searchErr, &valSearchErr)
	searchErrors := valSearchErr.Errors
	assert.Contains(t, searchErrors, "q")
	assert.Contains(t, searchErrors, "page")
	assert.Contains(t, searchErrors, "limit")
}

func TestValidationFallbackMessages_CoverConditionalAndTypeRules(t *testing.T) {
	type payload struct {
		Kind      string `json:"kind" validate:"required"`
		Value     string `json:"value" validate:"required_if=Kind special"`
		Enabled   string `json:"enabled" validate:"boolean"`
		StartedAt string `json:"started_at" validate:"date"`
	}

	err := requests.ValidateStruct(&payload{
		Kind:      "special",
		Value:     "",
		Enabled:   "maybe",
		StartedAt: "not-a-date",
	})
	require.Error(t, err)

	var valErr *exceptions.ValidationException
	require.ErrorAs(t, err, &valErr)
	errorsMap := valErr.Errors
	valueErrors, ok := errorsMap["value"]
	require.True(t, ok)
	assert.NotEmpty(t, valueErrors)
	assert.Contains(t, valueErrors[0], "required when")

	enabledErrors, ok := errorsMap["enabled"]
	require.True(t, ok)
	assert.NotEmpty(t, enabledErrors)
	assert.Contains(t, enabledErrors[0], "true or false")

	startedAtErrors, ok := errorsMap["started_at"]
	require.True(t, ok)
	assert.NotEmpty(t, startedAtErrors)
	assert.Contains(t, startedAtErrors[0], "valid date")

	type arrayPayload struct {
		Profile map[string]any `json:"profile" validate:"array=name username"`
	}

	arrayErr := requests.ValidateStruct(&arrayPayload{
		Profile: map[string]any{
			"name":     "alice",
			"username": "alice",
			"admin":    true,
		},
	})
	require.Error(t, arrayErr)
	var valArrayErr *exceptions.ValidationException
	require.ErrorAs(t, arrayErr, &valArrayErr)
	arrayErrors := valArrayErr.Errors
	profileErrors, ok := arrayErrors["profile"]
	require.True(t, ok)
	assert.NotEmpty(t, profileErrors)
	assert.Contains(t, profileErrors[0], "array")
}

func TestValidationFallbackMessages_CoverLaravelAliases(t *testing.T) {
	type payload struct {
		Username  string         `json:"username" validate:"alpha_num"`
		Slug      string         `json:"slug" validate:"alpha_dash"`
		Prefix    string         `json:"prefix" validate:"starts_with=pre"`
		Mac       string         `json:"mac" validate:"mac_address"`
		Items     map[string]any `json:"items" validate:"list"`
		Profile   map[string]any `json:"profile" validate:"required_array_keys=name username"`
		StartedAt string         `json:"started_at" validate:"date_format=2006-01-02"`
	}

	err := requests.ValidateStruct(&payload{
		Username:  "bad username!",
		Slug:      "bad slug!",
		Prefix:    "postfix",
		Mac:       "not-a-mac",
		Items:     map[string]any{"0": "first", "2": "third"},
		Profile:   map[string]any{"name": "alice"},
		StartedAt: "01-02-2006",
	})
	require.Error(t, err)

	var valErr *exceptions.ValidationException
	require.ErrorAs(t, err, &valErr)
	errorsMap := valErr.Errors
	assert.Contains(t, errorsMap, "username")
	assert.Contains(t, errorsMap, "slug")
	assert.Contains(t, errorsMap, "prefix")
	assert.Contains(t, errorsMap, "mac")
	assert.Contains(t, errorsMap, "items")
	assert.Contains(t, errorsMap, "profile")
	assert.Contains(t, errorsMap, "started_at")
}
