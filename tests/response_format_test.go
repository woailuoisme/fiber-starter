package tests

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	exceptions "fiber-starter/app/Exceptions"
	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
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

	successResp := mustDoRequest(t, app, "/success")
	assertResponseField(t, successResp, "success", true)
	assertResponseField(t, successResp, "code", float64(200))
	assertResponseField(t, successResp, "message", "User information fetched successfully")
	data := mustNestedMap(t, successResp, "data")
	assertResponseField(t, data, "id", float64(1))
	assertResponseField(t, data, "name", "Alice")

	errorResp := mustDoRequest(t, app, "/error")
	assertResponseField(t, errorResp, "success", false)
	assertResponseField(t, errorResp, "code", float64(422))
	assertResponseField(t, errorResp, "message", "Validation failed")
	errorsMap := mustNestedMap(t, errorResp, "errors")
	emailErrors, ok := errorsMap["email"].([]any)
	if !ok || len(emailErrors) != 1 || emailErrors[0] != "Email field is required." {
		t.Fatalf("unexpected validation errors: %#v", errorsMap["email"])
	}

	paginationResp := mustDoRequest(t, app, "/pagination")
	assertResponseField(t, paginationResp, "success", true)
	assertResponseField(t, paginationResp, "code", float64(200))
	assertResponseField(t, paginationResp, "message", "success")
	paginationData := mustNestedMap(t, paginationResp, "data")
	items, ok := paginationData["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("unexpected pagination items: %#v", paginationData["items"])
	}
	meta := mustNestedMap(t, paginationData, "meta")
	assertResponseField(t, meta, "current_page", float64(1))
	assertResponseField(t, meta, "per_page", float64(15))
	assertResponseField(t, meta, "last_page", float64(10))
	assertResponseField(t, meta, "has_more", true)
	assertResponseField(t, meta, "total", float64(150))
	assertResponseField(t, meta, "from", float64(1))
	assertResponseField(t, meta, "to", float64(15))
}

func mustDoRequest(t *testing.T, app *fiber.App, path string) map[string]any {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest("GET", path, nil))
	if err != nil {
		t.Fatalf("request %s failed: %v", path, err)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode %s response failed: %v", path, err)
	}

	return payload
}

func assertResponseField(t *testing.T, payload map[string]any, key string, want any) {
	t.Helper()

	got, ok := payload[key]
	if !ok {
		t.Fatalf("missing field %q in payload: %#v", key, payload)
	}
	if got != want {
		t.Fatalf("unexpected %s: got=%v want=%v", key, got, want)
	}
}

func mustNestedMap(t *testing.T, payload map[string]any, key string) map[string]any {
	t.Helper()

	value, ok := payload[key]
	if !ok {
		t.Fatalf("missing nested field %q in payload: %#v", key, payload)
	}

	nested, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("field %q is not a map: %#v", key, value)
	}

	return nested
}
