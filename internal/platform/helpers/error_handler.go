package helpers

import (
	models "fiber-starter/internal/domain/model"
	apierrors "fiber-starter/internal/platform/apierrors"
	"fiber-starter/internal/transport/http/resources"

	"github.com/gofiber/fiber/v3"
)

// HandleError Unified error handler function
func HandleError(ctx fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// If it's an application error, return directly
	if apierrors.IsAppError(err) {
		appErr, _ := apierrors.GetAppError(err)
		return ctx.Status(appErr.StatusCode).JSON(resources.ErrorResponse(appErr.Message, fiber.Map{
			"code":    appErr.Code,
			"details": appErr.Details,
		}))
	}

	// Handle other types of errors
	return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse("Internal server error", fiber.Map{
		"code": apierrors.ErrCodeInternalServer,
	}))
}

// HandleValidationError Handle validation error
func HandleValidationError(ctx fiber.Ctx, err error) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(
		"Request parameter validation failed",
		resources.FormatValidationErrors(err),
	))
}

// HandleSuccess Handle success response
func HandleSuccess(ctx fiber.Ctx, message string, data interface{}) error {
	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse(message, data))
}

// HandleCreated Handle created success response
func HandleCreated(ctx fiber.Ctx, message string, data interface{}) error {
	return ctx.Status(fiber.StatusCreated).JSON(resources.SuccessResponse(message, data))
}

// HandleNotFound Handle not found error
func HandleNotFound(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusNotFound).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeNotFound,
	}))
}

// HandleBadRequest Handle bad request error
func HandleBadRequest(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeBadRequest,
	}))
}

// HandleUnauthorized Handle unauthorized error
func HandleUnauthorized(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusUnauthorized).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeUnauthorized,
	}))
}

// HandleForbidden Handle forbidden error
func HandleForbidden(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusForbidden).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeForbidden,
	}))
}

// HandleConflict Handle conflict error
func HandleConflict(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusConflict).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeConflict,
	}))
}

// ParseAndValidate Parse and validate request parameters
func ParseAndValidate(ctx fiber.Ctx, req interface{}, validate interface{}) error {
	// Parse request body
	if err := ctx.Bind().Body(req); err != nil {
		return apierrors.BadRequestWithDetails("Failed to parse request parameters", err.Error())
	}

	// Validate request parameters
	if validator, ok := validate.(interface{ Struct(interface{}) error }); ok {
		if err := validator.Struct(req); err != nil {
			return apierrors.ValidationWithDetails("Failed to validate request parameters", FormatValidationErrorsToString(err))
		}
	}

	return nil
}

// FormatValidationErrorsToString Convert validation errors to string
func FormatValidationErrorsToString(err error) string {
	if validationErrors := resources.FormatValidationErrors(err); validationErrors != nil {
		result := ""
		for field, messages := range validationErrors {
			for _, message := range messages {
				if result != "" {
					result += "; "
				}
				result += field + ": " + message
			}
		}
		return result
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
		"users": userResponses,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// HandlePaginationResponse Handle pagination response
func HandlePaginationResponse(ctx fiber.Ctx, message string, data interface{}, total int64, page, limit int) error {
	return HandleSuccess(ctx, message, fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
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
	var result int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, apierrors.BadRequest("Invalid number format")
		}
		result = result*10 + int(r-'0')
	}
	return result, nil
}
