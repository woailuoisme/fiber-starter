package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"lfiber/configs"
	requests "lfiber/internal/common/requests"
	validation "lfiber/internal/providers/validation"
	helpers "lfiber/internal/support"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateQueryRules_UsesValidatorRules(t *testing.T) {
	v, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)
	requests.InitValidator(v)
	t.Cleanup(func() {
		requests.InitValidator(nil)
	})

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
	v, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)
	requests.InitValidator(v)
	t.Cleanup(func() {
		requests.InitValidator(nil)
	})

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
