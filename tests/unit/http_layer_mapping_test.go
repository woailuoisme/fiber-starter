package tests

import (
	"testing"
	"time"

	requests "fiber-starter/app/Http/Requests"
	resources "fiber-starter/app/Http/Resources"
	services "fiber-starter/app/Http/Services"
	models "fiber-starter/app/Models"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequests_ToServiceInputs(t *testing.T) {
	register := requests.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	}
	assert.Equal(t, services.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	}, register.ToInput())

	profile := requests.UpdateProfileRequest{
		Name:   "Alice",
		Phone:  "+8613800138000",
		Avatar: "https://example.com/avatar.jpg",
	}
	input := profile.ToInput()
	require.NotNil(t, input.Name)
	require.NotNil(t, input.Phone)
	require.NotNil(t, input.Avatar)
	assert.Equal(t, "Alice", *input.Name)
	assert.Equal(t, "+8613800138000", *input.Phone)
	assert.Equal(t, "https://example.com/avatar.jpg", *input.Avatar)
}

func TestRequests_ToQueryNormalizesPagination(t *testing.T) {
	list := requests.UserListRequest{Page: 0, Limit: 0}.ToQuery()
	assert.Equal(t, services.UserListQuery{Page: 1, Limit: 10}, list)

	search := requests.SearchUsersRequest{Q: "alice", Page: 2, Limit: 20}.ToQuery()
	assert.Equal(t, services.UserListQuery{Search: "alice", Page: 2, Limit: 20}, search)
}

func TestResources_KeepUserAndTokenShape(t *testing.T) {
	createdAt := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
	user := &models.User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Status:    models.UserStatusActive,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	userResponse := resources.NewUserResource(user).ToResponse()
	assert.Equal(t, int64(1), userResponse.ID)
	assert.Equal(t, "Alice", userResponse.Name)
	assert.Equal(t, "alice@example.com", userResponse.Email)

	authResponse := resources.NewAuthResultResource(services.AuthResult{
		User: user,
		Tokens: services.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},
	}).ToResponse()
	assert.Equal(t, userResponse, authResponse["user"])
	tokens, ok := authResponse["tokens"].(map[string]interface{})
	if !ok {
		tokens = map[string]interface{}(authResponse["tokens"].(fiber.Map))
	}
	assert.Equal(t, "access-token", tokens["access_token"])
}
