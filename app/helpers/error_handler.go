package helpers

import (
	apierrors "fiber-starter/app/apierrors"
	"fiber-starter/app/http/resources"
	"fiber-starter/app/models"

	"github.com/gofiber/fiber/v3"
)

// HandleError 统一错误处理函数
func HandleError(ctx fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// 如果是应用程序错误，直接返回
	if apierrors.IsAppError(err) {
		appErr, _ := apierrors.GetAppError(err)
		return ctx.Status(appErr.StatusCode).JSON(resources.ErrorResponse(appErr.Message, fiber.Map{
			"code":    appErr.Code,
			"details": appErr.Details,
		}))
	}

	// 处理其他类型的错误
	return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse("内部服务器错误", fiber.Map{
		"code": apierrors.ErrCodeInternalServer,
	}))
}

// HandleValidationError 处理验证错误
func HandleValidationError(ctx fiber.Ctx, err error) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(
		"请求参数验证失败",
		resources.FormatValidationErrors(err),
	))
}

// HandleSuccess 处理成功响应
func HandleSuccess(ctx fiber.Ctx, message string, data interface{}) error {
	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse(message, data))
}

// HandleCreated 处理创建成功响应
func HandleCreated(ctx fiber.Ctx, message string, data interface{}) error {
	return ctx.Status(fiber.StatusCreated).JSON(resources.SuccessResponse(message, data))
}

// HandleNotFound 处理未找到错误
func HandleNotFound(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusNotFound).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeNotFound,
	}))
}

// HandleBadRequest 处理请求错误
func HandleBadRequest(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeBadRequest,
	}))
}

// HandleUnauthorized 处理未授权错误
func HandleUnauthorized(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusUnauthorized).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeUnauthorized,
	}))
}

// HandleForbidden 处理禁止访问错误
func HandleForbidden(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusForbidden).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeForbidden,
	}))
}

// HandleConflict 处理冲突错误
func HandleConflict(ctx fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusConflict).JSON(resources.ErrorResponse(message, fiber.Map{
		"code": apierrors.ErrCodeConflict,
	}))
}

// ParseAndValidate 解析和验证请求参数
func ParseAndValidate(ctx fiber.Ctx, req interface{}, validate interface{}) error {
	// 解析请求体
	if err := ctx.Bind().Body(req); err != nil {
		return apierrors.BadRequestWithDetails("请求参数解析失败", err.Error())
	}

	// 验证请求参数
	if validator, ok := validate.(interface{ Struct(interface{}) error }); ok {
		if err := validator.Struct(req); err != nil {
			return apierrors.ValidationWithDetails("请求参数验证失败", FormatValidationErrorsToString(err))
		}
	}

	return nil
}

// FormatValidationErrorsToString 将验证错误转换为字符串
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

// HandleUserResponse 处理用户响应
func HandleUserResponse(ctx fiber.Ctx, message string, user *models.User) error {
	return HandleSuccess(ctx, message, user.ToSafeUser())
}

// HandleUserListResponse 处理用户列表响应
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

// HandlePaginationResponse 处理分页响应
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

// ValidatePaginationParams 验证分页参数
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

// parseInt 安全的整数解析
func parseInt(s string) (int, error) {
	var result int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, apierrors.BadRequest("无效的数字格式")
		}
		result = result*10 + int(r-'0')
	}
	return result, nil
}
