package tests

import (
	"errors"
	"testing"

	exceptions "fiber-starter/internal/common/exceptions"
	helpers "fiber-starter/internal/support"
	"fiber-starter/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestHTTPBehaviorContract_ResponseEnvelopeShape(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Get("/ok", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "ok", fiber.Map{"items": []string{}})
	})
	app.Post("/validate", func(c fiber.Ctx) error {
		return helpers.HandleAppError(c, exceptions.NewValidationException("invalid payload"))
	})
	app.Post("/validate-with-errors", func(c fiber.Ctx) error {
		return helpers.HandleAppError(c, exceptions.NewValidationExceptionWithErrors("Validation failed", map[string][]string{
			"email": {"Email field is required."},
		}))
	})

	okResp := testkit.DoRequest(t, app, "GET", "/ok", "")
	okPayload := testkit.AssertSuccessEnvelope(t, okResp, fiber.StatusOK)
	assert.Equal(t, "ok", okPayload["message"])

	validationResp := testkit.DoRequest(t, app, "POST", "/validate", "{}")
	validationPayload := testkit.AssertErrorEnvelope(t, validationResp, fiber.StatusUnprocessableEntity)
	assert.NotContains(t, validationPayload, "errors")

	validationWithErrorsResp := testkit.DoRequest(t, app, "POST", "/validate-with-errors", "{}")
	validationWithErrorsPayload := testkit.AssertErrorEnvelope(t, validationWithErrorsResp, fiber.StatusUnprocessableEntity)
	assert.Contains(t, validationWithErrorsPayload, "errors")
}

func TestHTTPBehaviorContract_NotFoundAndMethodNotAllowedUseEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Get("/resource", func(c fiber.Ctx) error {
		return helpers.HandleSuccess(c, "ok", fiber.Map{})
	})

	notFoundResp := testkit.DoRequest(t, app, "GET", "/missing", "")
	notFoundPayload := testkit.AssertErrorEnvelope(t, notFoundResp, fiber.StatusNotFound)
	assert.NotContains(t, notFoundPayload, "errors")

	methodResp := testkit.DoRequest(t, app, "POST", "/resource", "")
	methodPayload := testkit.AssertErrorEnvelope(t, methodResp, fiber.StatusMethodNotAllowed)
	assert.NotContains(t, methodPayload, "errors")
}

func TestHTTPBehaviorContract_RateLimitServiceUnavailableAndUnknownErrorUseEnvelope(t *testing.T) {
	t.Setenv("APP_DEBUG", "false")

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Get("/limited", func(c fiber.Ctx) error {
		return fiber.ErrTooManyRequests
	})
	app.Get("/unavailable", func(c fiber.Ctx) error {
		return exceptions.NewServiceUnavailableException("Service unavailable")
	})
	app.Get("/panic-safe", func(c fiber.Ctx) error {
		return errors.New("database password=secret-token failed")
	})

	limitedResp := testkit.DoRequest(t, app, "GET", "/limited", "")
	testkit.AssertErrorEnvelope(t, limitedResp, fiber.StatusTooManyRequests)

	unavailableResp := testkit.DoRequest(t, app, "GET", "/unavailable", "")
	testkit.AssertErrorEnvelope(t, unavailableResp, fiber.StatusServiceUnavailable)

	unknownResp := testkit.DoRequest(t, app, "GET", "/panic-safe", "")
	unknownPayload := testkit.AssertErrorEnvelope(t, unknownResp, fiber.StatusInternalServerError)
	assert.Equal(t, "Internal server error", unknownPayload["message"])
	assert.NotContains(t, unknownPayload, "exception")
}

func TestHTTPBehaviorContract_DebugExceptionOnlyForThrownErrors(t *testing.T) {
	t.Setenv("APP_DEBUG", "true")

	app := fiber.New(fiber.Config{ErrorHandler: helpers.HandleHTTPError})
	app.Get("/exception", func(c fiber.Ctx) error {
		return exceptions.NewAPIException("failed", fiber.StatusInternalServerError)
	})

	notFoundResp := testkit.DoRequest(t, app, "GET", "/missing", "")
	notFoundPayload := testkit.AssertErrorEnvelope(t, notFoundResp, fiber.StatusNotFound)
	assert.NotContains(t, notFoundPayload, "exception")

	exceptionResp := testkit.DoRequest(t, app, "GET", "/exception", "")
	exceptionPayload := testkit.AssertErrorEnvelope(t, exceptionResp, fiber.StatusInternalServerError)
	assert.Contains(t, exceptionPayload, "exception")
}
