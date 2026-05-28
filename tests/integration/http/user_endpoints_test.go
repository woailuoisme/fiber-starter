package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEndpoints_CurrentUserAndList(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "Alice", "alice@example.com", "password123")

	// Login to get token
	token := getAccessToken(t, app, "alice@example.com", "password123")

	// 1. Get Me
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var meResult struct {
		Success bool `json:"success"`
		Data    struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&meResult))
	assert.True(t, meResult.Success)
	assert.Equal(t, "alice@example.com", meResult.Data.Email)

	// 2. Get Users List
	req = httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var listResult struct {
		Success bool `json:"success"`
		Data    struct {
			Items []any `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResult))
	assert.True(t, listResult.Success)
	assert.NotEmpty(t, listResult.Data.Items)

	// 3. Search Users
	req = httptest.NewRequest("GET", "/api/v1/users/search?q=Ali", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var searchResult struct {
		Success bool `json:"success"`
		Data    struct {
			Items []any `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&searchResult))
	assert.True(t, searchResult.Success)
	assert.NotEmpty(t, searchResult.Data.Items)
}

func TestUserEndpoints_RejectsUserManagementWithoutPermission(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "Admin", "admin@example.com", "password123")
	registerAndVerifyUser(t, app, mailer, "Bob", "bob@example.com", "password123")

	token := getAccessToken(t, app, "bob@example.com", "password123")

	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	req = httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUserEndpoints_UserCRUDAndProfile(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "Charlie", "charlie@example.com", "password123")

	token := getAccessToken(t, app, "charlie@example.com", "password123")

	// 1. Get Me to resolve current User ID
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	var meResult struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&meResult))
	userID := meResult.Data.ID

	// 2. Get Specific User
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var userResult struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&userResult))
	assert.Equal(t, "Charlie", userResult.Data.Name)

	// 3. Update User
	updateReq := map[string]string{
		"name": "Charlie Updated",
	}
	body, err := json.Marshal(updateReq)
	require.NoError(t, err)

	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", userID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 4. Update Profile
	profileReq := map[string]string{
		"name":   "Charlie Profile",
		"phone":  "+8613800138000",
		"avatar": "https://example.com/avatar.jpg",
	}
	body, err = json.Marshal(profileReq)
	require.NoError(t, err)

	req = httptest.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 5. Delete User
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", userID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserEndpoints_DataExchange(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "Exporter", "exporter@example.com", "password123")

	token := getAccessToken(t, app, "exporter@example.com", "password123")

	// 1. Export Users
	req := httptest.NewRequest("GET", "/api/v1/users/export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")

	// 2. Import Users
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "users.xlsx")
	require.NoError(t, err)
	// Write dummy empty bytes to simulate Excel file uploading (validates endpoint logic, no parser validation needed)
	_, err = part.Write([]byte("dummy-excel-data"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req = httptest.NewRequest("POST", "/api/v1/users/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	// Expect 400 Bad Request because Excel parser will fail on dummy bytes, which proves the route hit the logic
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func getAccessToken(t *testing.T, app *fiber.App, email, password string) string {
	t.Helper()

	signInReq := map[string]string{
		"email":    email,
		"password": password,
	}
	body, err := json.Marshal(signInReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var signInResult struct {
		Data struct {
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signInResult))
	return signInResult.Data.Tokens.AccessToken
}
