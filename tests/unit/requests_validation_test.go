package tests

import (
	"testing"

	"lfiber/configs"
	requests "lfiber/internal/common/requests"
	"lfiber/internal/features/auth"
	"lfiber/internal/features/user"
	supporti18n "lfiber/internal/providers/i18n"
	validation "lfiber/internal/providers/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestValidationHelpers_UseValidatorRules(t *testing.T) {
	v, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)
	requests.InitValidator(v)
	t.Cleanup(func() {
		requests.InitValidator(nil)
	})

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
	factory, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)

	registerErr := factory.Make(&auth.RegisterRequest{
		Name:     "A",
		Email:    "not-an-email",
		Password: "short",
	}, nil, nil, nil).Validate()
	require.Error(t, registerErr)
	registerErrors := supporti18n.FormatValidationErrors(registerErr)
	assert.Contains(t, registerErrors, "name")
	assert.Contains(t, registerErrors, "email")
	assert.Contains(t, registerErrors, "password")

	profileErr := factory.Make(&user.UpdateProfileRequest{
		Name:   "A",
		Phone:  "not-a-phone",
		Avatar: "not-a-url",
	}, nil, nil, nil).Validate()
	require.Error(t, profileErr)
	profileErrors := supporti18n.FormatValidationErrors(profileErr)
	assert.Contains(t, profileErrors, "name")
	assert.Contains(t, profileErrors, "phone")
	assert.Contains(t, profileErrors, "avatar")

	searchErr := factory.Make(&user.SearchUsersRequest{
		Q:     "",
		Page:  -1,
		Limit: 101,
	}, nil, nil, nil).Validate()
	require.Error(t, searchErr)
	searchErrors := supporti18n.FormatValidationErrors(searchErr)
	assert.Contains(t, searchErrors, "q")
	assert.Contains(t, searchErrors, "page")
	assert.Contains(t, searchErrors, "limit")
}

func TestValidationFallbackMessages_CoverConditionalAndTypeRules(t *testing.T) {
	factory, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)

	type payload struct {
		Kind      string `json:"kind" validate:"required"`
		Value     string `json:"value" validate:"required_if=Kind special"`
		Enabled   string `json:"enabled" validate:"boolean"`
		StartedAt string `json:"started_at" validate:"date"`
	}

	err = factory.Make(&payload{
		Kind:      "special",
		Value:     "",
		Enabled:   "maybe",
		StartedAt: "not-a-date",
	}, nil, nil, nil).Validate()
	require.Error(t, err)

	errorsMap := supporti18n.FormatValidationErrors(err)
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

	arrayErr := factory.Make(&arrayPayload{
		Profile: map[string]any{
			"name":     "alice",
			"username": "alice",
			"admin":    true,
		},
	}, nil, nil, nil).Validate()
	require.Error(t, arrayErr)
	arrayErrors := supporti18n.FormatValidationErrors(arrayErr)
	profileErrors, ok := arrayErrors["profile"]
	require.True(t, ok)
	assert.NotEmpty(t, profileErrors)
	assert.Contains(t, profileErrors[0], "array")
}

func TestValidationFallbackMessages_CoverLaravelAliases(t *testing.T) {
	factory, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)

	type payload struct {
		Username  string         `json:"username" validate:"alpha_num"`
		Slug      string         `json:"slug" validate:"alpha_dash"`
		Prefix    string         `json:"prefix" validate:"starts_with=pre"`
		Mac       string         `json:"mac" validate:"mac_address"`
		Items     map[string]any `json:"items" validate:"list"`
		Profile   map[string]any `json:"profile" validate:"required_array_keys=name username"`
		StartedAt string         `json:"started_at" validate:"date_format=2006-01-02"`
	}

	err = factory.Make(&payload{
		Username:  "bad username!",
		Slug:      "bad slug!",
		Prefix:    "postfix",
		Mac:       "not-a-mac",
		Items:     map[string]any{"0": "first", "2": "third"},
		Profile:   map[string]any{"name": "alice"},
		StartedAt: "01-02-2006",
	}, nil, nil, nil).Validate()
	require.Error(t, err)

	errorsMap := supporti18n.FormatValidationErrors(err)
	assert.Contains(t, errorsMap, "username")
	assert.Contains(t, errorsMap, "slug")
	assert.Contains(t, errorsMap, "prefix")
	assert.Contains(t, errorsMap, "mac")
	assert.Contains(t, errorsMap, "items")
	assert.Contains(t, errorsMap, "profile")
	assert.Contains(t, errorsMap, "started_at")
}
