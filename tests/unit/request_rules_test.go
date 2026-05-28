package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	requests "lfiber/internal/common/requests"
	helpers "lfiber/internal/support"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateQueryRules_UsesValidatorRules(t *testing.T) {
	app := fiber.New()
	app.Get("/search", func(c fiber.Ctx) error {
		err := requests.ValidateQueryRules(c, map[string]string{
			"q":     "required,min=2",
			"page":  "omitempty,gte=1",
			"limit": "omitempty,lte=100",
		})
		if err != nil {
			return helpers.HandleAppError(c, err)
		}
		return c.SendString("ok")
	})

	resp := testkit.DoRequest(t, app, "GET", "/search?q=a&page=0&limit=101", "")
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	payload := testkit.JSONBody(t, resp)
	errorsMap := payload["errors"].(map[string]any)
	assert.Contains(t, errorsMap, "q")
	assert.Contains(t, errorsMap, "limit")
}

func TestValidateUploadedFile_UsesValidatorRules(t *testing.T) {
	app := fiber.New()
	app.Post("/upload", func(c fiber.Ctx) error {
		_, err := requests.ValidateUploadedFile(c, "avatar", 1024, []string{"image/png"})
		if err != nil {
			return helpers.HandleAppError(c, err)
		}
		return c.SendString("ok")
	})

	resp := doMultipartUpload(t, app, "/upload", "avatar", "avatar.txt", "text/plain", []byte("hello"))
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	payload := testkit.JSONBody(t, resp)
	errorsMap := payload["errors"].(map[string]any)
	assert.Contains(t, errorsMap, "avatar")
}

func TestRequestBindingHelpers_UseFiberStructValidator(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler:    helpers.HandleHTTPError,
		StructValidator: requests.NewStructValidator(),
	})
	app.Post("/body", func(c fiber.Ctx) error {
		var payload struct {
			Email string `json:"email" validate:"required,email"`
		}
		return requests.Body(c, &payload)
	})
	app.Get("/query", func(c fiber.Ctx) error {
		var payload struct {
			Page int `query:"page" validate:"required,gte=1"`
		}
		return requests.Query(c, &payload)
	})
	app.Get("/users/:id", func(c fiber.Ctx) error {
		var payload struct {
			ID int64 `uri:"id" validate:"required,gt=0"`
		}
		return requests.URI(c, &payload)
	})
	app.Post("/form", func(c fiber.Ctx) error {
		var payload struct {
			File *multipart.FileHeader `form:"file" validate:"required,uploaded_file"`
		}
		return requests.Form(c, &payload)
	})

	bodyResp := testkit.DoRequest(t, app, "POST", "/body", `{}`)
	bodyPayload := testkit.AssertErrorEnvelope(t, bodyResp, fiber.StatusUnprocessableEntity)
	bodyErrors := bodyPayload["errors"].(map[string]any)
	assert.Contains(t, bodyErrors, "email")

	queryResp := testkit.DoRequest(t, app, "GET", "/query?page=0", "")
	queryPayload := testkit.AssertErrorEnvelope(t, queryResp, fiber.StatusUnprocessableEntity)
	queryErrors := queryPayload["errors"].(map[string]any)
	assert.Contains(t, queryErrors, "page")

	uriResp := testkit.DoRequest(t, app, "GET", "/users/0", "")
	uriPayload := testkit.AssertErrorEnvelope(t, uriResp, fiber.StatusUnprocessableEntity)
	uriErrors := uriPayload["errors"].(map[string]any)
	assert.Contains(t, uriErrors, "id")

	formResp := doEmptyMultipart(t, app, "/form")
	formPayload := testkit.AssertErrorEnvelope(t, formResp, fiber.StatusUnprocessableEntity)
	formErrors := formPayload["errors"].(map[string]any)
	assert.Contains(t, formErrors, "file")
}

func doMultipartUpload(t *testing.T, app *fiber.App, path, field, filename, contentType string, body []byte) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="`+field+`"; filename="`+filename+`"`)
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	require.NoError(t, err)
	_, err = part.Write(body)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func doEmptyMultipart(t *testing.T, app *fiber.App, path string) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}
