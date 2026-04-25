package support

import (
	"strconv"
	"strings"

	"fiber-starter/app/Exceptions"
	"fiber-starter/app/Http/Resources"
	models "fiber-starter/app/Models"

	"github.com/gofiber/fiber/v3"
)

func writeResponse(ctx fiber.Ctx, status int, payload any) error {
	return ctx.Status(status).JSON(payload)
}

func errorPayload(code exceptions.ErrorCode, details any) fiber.Map {
	return fiber.Map{
		"code":    code,
		"details": details,
	}
}

func paginationPayload(data any, total int64, page, limit int) fiber.Map {
	return fiber.Map{
		"data":       data,
		"pagination": paginationMeta(total, page, limit),
	}
}

func paginationMeta(total int64, page, limit int) fiber.Map {
	return fiber.Map{
		"page":  page,
		"limit": limit,
		"total": total,
		"pages": (total + int64(limit) - 1) / int64(limit),
	}
}

// HandleError Unified error handler function
func HandleError(ctx fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if exceptions.IsAppError(err) {
		appErr, _ := exceptions.GetAppError(err)
		return writeResponse(ctx, appErr.StatusCode, resources.ErrorResponse(appErr.Message, errorPayload(appErr.Code, appErr.Details)))
	}

	return writeResponse(ctx, fiber.StatusInternalServerError, resources.ErrorResponse("Internal server error", errorPayload(exceptions.ErrCodeInternalServer, nil)))
}

// HandleValidationError Handle validation error
func HandleValidationError(ctx fiber.Ctx, err error) error {
	return writeResponse(ctx, fiber.StatusBadRequest, resources.ErrorResponse(
		"Request parameter validation failed",
		resources.FormatValidationErrors(err),
	))
}

// HandleSuccess Handle success response
func HandleSuccess(ctx fiber.Ctx, message string, data interface{}) error {
	return writeResponse(ctx, fiber.StatusOK, resources.SuccessResponse(message, data))
}

// HandleCreated Handle created success response
func HandleCreated(ctx fiber.Ctx, message string, data interface{}) error {
	return writeResponse(ctx, fiber.StatusCreated, resources.SuccessResponse(message, data))
}

// HandleNotFound Handle not found error
func HandleNotFound(ctx fiber.Ctx, message string) error {
	return writeResponse(ctx, fiber.StatusNotFound, resources.ErrorResponse(message, errorPayload(exceptions.ErrCodeNotFound, nil)))
}

// HandleBadRequest Handle bad request error
func HandleBadRequest(ctx fiber.Ctx, message string) error {
	return writeResponse(ctx, fiber.StatusBadRequest, resources.ErrorResponse(message, errorPayload(exceptions.ErrCodeBadRequest, nil)))
}

// HandleUnauthorized Handle unauthorized error
func HandleUnauthorized(ctx fiber.Ctx, message string) error {
	return writeResponse(ctx, fiber.StatusUnauthorized, resources.ErrorResponse(message, errorPayload(exceptions.ErrCodeUnauthorized, nil)))
}

// HandleForbidden Handle forbidden error
func HandleForbidden(ctx fiber.Ctx, message string) error {
	return writeResponse(ctx, fiber.StatusForbidden, resources.ErrorResponse(message, errorPayload(exceptions.ErrCodeForbidden, nil)))
}

// HandleConflict Handle conflict error
func HandleConflict(ctx fiber.Ctx, message string) error {
	return writeResponse(ctx, fiber.StatusConflict, resources.ErrorResponse(message, errorPayload(exceptions.ErrCodeConflict, nil)))
}

// ParseAndValidate Parse and validate request parameters
func ParseAndValidate(ctx fiber.Ctx, req interface{}, validate interface{}) error {
	if err := ctx.Bind().Body(req); err != nil {
		return exceptions.BadRequestWithDetails("Failed to parse request parameters", err.Error())
	}

	if validator, ok := validate.(interface{ Struct(interface{}) error }); ok {
		if err := validator.Struct(req); err != nil {
			return exceptions.ValidationWithDetails("Failed to validate request parameters", FormatValidationErrorsToString(err))
		}
	}

	return nil
}

// FormatValidationErrorsToString Convert validation errors to string
func FormatValidationErrorsToString(err error) string {
	if validationErrors := resources.FormatValidationErrors(err); validationErrors != nil {
		parts := make([]string, 0, len(validationErrors))
		for field, messages := range validationErrors {
			for _, message := range messages {
				parts = append(parts, field+": "+message)
			}
		}
		return strings.Join(parts, "; ")
	}
	return err.Error()
}

// HandleUserResponse Handle user response
func HandleUserResponse(ctx fiber.Ctx, message string, user *models.User) error {
	return HandleSuccess(ctx, message, user.ToSafeUser())
}

// HandleUserListResponse Handle user list response
func HandleUserListResponse(ctx fiber.Ctx, message string, users []models.User, total int64, page, limit int) error {
	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
	}

	return HandleSuccess(ctx, message, fiber.Map{
		"users":      userResponses,
		"pagination": paginationMeta(total, page, limit),
	})
}

// HandlePaginationResponse Handle pagination response
func HandlePaginationResponse(ctx fiber.Ctx, message string, data interface{}, total int64, page, limit int) error {
	return HandleSuccess(ctx, message, paginationPayload(data, total, page, limit))
}

// ValidatePaginationParams Validate pagination parameters
func ValidatePaginationParams(pageStr, limitStr string) (int, int, error) {
	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := parseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := parseInt(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return page, limit, nil
}

// parseInt Safe integer parsing
func parseInt(s string) (int, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, exceptions.BadRequest("Invalid number format")
	}
	return value, nil
}
