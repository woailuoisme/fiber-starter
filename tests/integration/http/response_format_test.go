package tests

import (
	"testing"

	exceptions "fiber-starter/app/Exceptions"
	helpers "fiber-starter/app/Support"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseFormat_SuccessErrorAndPagination(t *testing.T) {
	app := fiber.New()
	app.Get("/success", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "User information fetched successfully", fiber.Map{
			"id":   1,
			"name": "Alice",
		})
	})
	app.Get("/error", func(c fiber.Ctx) error {
		return helpers.HandleAppError(c, exceptions.NewValidationExceptionWithErrors("Validation failed", map[string][]string{
			"email": {"Email field is required."},
		}))
	})
	app.Get("/pagination", func(c fiber.Ctx) error {
		items := []fiber.Map{
			{"id": 1},
			{"id": 2},
		}
		return helpers.HandlePaginationResponse(c, "success", items, 150, 1, 15)
	})

	successResp := testkit.DoRequest(t, app, "GET", "/success", "")
	successPayload := testkit.JSONBody(t, successResp)
	assert.Equal(t, true, successPayload["success"])
	assert.EqualValues(t, 200, successPayload["code"])
	assert.Equal(t, "User information fetched successfully", successPayload["message"])
	data := successPayload["data"].(map[string]any)
	assert.EqualValues(t, 1, data["id"])
	assert.Equal(t, "Alice", data["name"])

	errorResp := testkit.DoRequest(t, app, "GET", "/error", "")
	errorPayload := testkit.JSONBody(t, errorResp)
	assert.Equal(t, false, errorPayload["success"])
	assert.EqualValues(t, 422, errorPayload["code"])
	assert.Equal(t, "Validation failed", errorPayload["message"])
	errorsMap := errorPayload["errors"].(map[string]any)
	emailErrors, ok := errorsMap["email"].([]any)
	require.True(t, ok)
	require.Len(t, emailErrors, 1)
	assert.Equal(t, "Email field is required.", emailErrors[0])

	paginationResp := testkit.DoRequest(t, app, "GET", "/pagination", "")
	paginationPayload := testkit.JSONBody(t, paginationResp)
	assert.Equal(t, true, paginationPayload["success"])
	assert.EqualValues(t, 200, paginationPayload["code"])
	assert.Equal(t, "success", paginationPayload["message"])
	paginationData := paginationPayload["data"].(map[string]any)
	items, ok := paginationData["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 2)
	meta := paginationData["meta"].(map[string]any)
	assert.EqualValues(t, 1, meta["current_page"])
	assert.EqualValues(t, 15, meta["per_page"])
	assert.EqualValues(t, 10, meta["last_page"])
	assert.Equal(t, true, meta["has_more"])
	assert.EqualValues(t, 150, meta["total"])
	assert.EqualValues(t, 1, meta["from"])
	assert.EqualValues(t, 15, meta["to"])
}
