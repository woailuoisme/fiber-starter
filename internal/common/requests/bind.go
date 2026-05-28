package requests

import (
	"errors"
	"fmt"

	exceptions "lfiber/internal/common/exceptions"
	supporti18n "lfiber/internal/providers/i18n"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// Body 绑定请求体，并通过 Fiber StructValidator 执行结构体校验。
func Body(c fiber.Ctx, req any) error {
	return normalizeBindError(c, c.Bind().Body(req), "body")
}

// Query 绑定查询参数，并通过 Fiber StructValidator 执行结构体校验。
func Query(c fiber.Ctx, req any) error {
	return normalizeBindError(c, c.Bind().Query(req), "query")
}

// URI 绑定路径参数，并通过 Fiber StructValidator 执行结构体校验。
func URI(c fiber.Ctx, req any) error {
	return normalizeBindError(c, c.Bind().URI(req), "uri")
}

// Form 绑定表单参数和 multipart 文件，并通过 Fiber StructValidator 执行结构体校验。
func Form(c fiber.Ctx, req any) error {
	return normalizeBindError(c, c.Bind().Form(req), "form")
}

func normalizeBindError(c fiber.Ctx, err error, source string) error {
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		formatted := supporti18n.FormatValidationErrorsWithContext(c, validationErrors)
		return exceptions.NewValidationExceptionWithErrors("Validation failed", formatted)
	}

	var bindErr *fiber.BindError
	if errors.As(err, &bindErr) {
		source = bindErr.Source
	}

	return exceptions.BadRequestWithDetails(bindErrorMessage(source), err.Error())
}

func bindErrorMessage(source string) string {
	switch source {
	case fiber.BindSourceBody:
		return "Invalid request body"
	case fiber.BindSourceQuery:
		return "Invalid query parameters"
	case fiber.BindSourceURI:
		return "Invalid path parameters"
	case "form":
		return "Invalid form data"
	default:
		return fmt.Sprintf("Invalid %s parameters", source)
	}
}
