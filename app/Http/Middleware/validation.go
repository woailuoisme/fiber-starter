package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"fiber-starter/app/Exceptions"
	helpers "fiber-starter/app/Http/Requests"

	"github.com/gofiber/fiber/v3"
)

// ValidationMiddleware 验证中间件
// Requirements: 10.1, 10.6, 10.7
func ValidationMiddleware(model interface{}) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := c.Bind().Body(model); err != nil {
			return exceptions.NewBadRequestException("Invalid request body")
		}

		if err := helpers.ValidateStruct(model); err != nil {
			return err
		}

		c.Locals("validated_model", model)
		return c.Next()
	}
}

// GetValidatedModel 从上下文中获取验证后的模型
func GetValidatedModel(c fiber.Ctx, model interface{}) bool {
	if validatedModel := c.Locals("validated_model"); validatedModel != nil {
		srcValue := reflect.ValueOf(validatedModel)
		dstValue := reflect.ValueOf(model)

		if srcValue.Kind() == reflect.Ptr && dstValue.Kind() == reflect.Ptr {
			srcValue = srcValue.Elem()
			dstValue = dstValue.Elem()

			if srcValue.Type().AssignableTo(dstValue.Type()) {
				dstValue.Set(srcValue)
				return true
			}
		}
	}
	return false
}

// QueryValidationMiddleware 查询参数验证中间件
func QueryValidationMiddleware(rules map[string]string) fiber.Handler {
	return func(c fiber.Ctx) error {
		errors := make(map[string][]string)

		for field, rule := range rules {
			value := c.Query(field)
			if fieldErrors := validateQueryField(field, value, rule); len(fieldErrors) > 0 {
				errors[field] = fieldErrors
			}
		}

		if len(errors) > 0 {
			return exceptions.NewValidationExceptionWithErrors("Query validation failed", errors)
		}

		return c.Next()
	}
}

func ruleValue(rule, key string) (string, bool) {
	prefix := key + ":"
	_, value, ok := strings.Cut(rule, prefix)
	if !ok {
		return "", false
	}

	end := strings.IndexByte(value, ',')
	if end >= 0 {
		value = value[:end]
	}

	return strings.TrimSpace(value), true
}

func validateQueryField(field, value, rule string) []string {
	errors := make([]string, 0, 3)

	if strings.Contains(rule, "required") && value == "" {
		errors = append(errors, fmt.Sprintf("The %s field is required.", field))
		return errors
	}

	if value == "" {
		return errors
	}

	if strings.Contains(rule, "number") {
		var num int64
		if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
			errors = append(errors, fmt.Sprintf("The %s must be a number.", field))
		}
	}

	if minValue, ok := ruleValue(rule, "min"); ok {
		minLength := 0
		_, _ = fmt.Sscanf(minValue, "%d", &minLength)
		if len(value) < minLength {
			errors = append(errors, fmt.Sprintf("The %s must be at least %d characters.", field, minLength))
		}
	}

	if maxValue, ok := ruleValue(rule, "max"); ok {
		maxLength := 0
		_, _ = fmt.Sscanf(maxValue, "%d", &maxLength)
		if len(value) > maxLength {
			errors = append(errors, fmt.Sprintf("The %s must not exceed %d characters.", field, maxLength))
		}
	}

	return errors
}

// FileValidationMiddleware 文件验证中间件
// Requirements: 8.2, 8.3
func FileValidationMiddleware(maxSize int64, allowedTypes []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return exceptions.NewBadRequestException("File upload failed")
		}

		if file.Size > maxSize {
			return exceptions.NewBadRequestException(
				fmt.Sprintf("File size must not exceed %d MB", maxSize/(1024*1024)),
			)
		}

		contentType := file.Header.Get("Content-Type")
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return exceptions.NewBadRequestException(
				fmt.Sprintf("Unsupported file type: %s", contentType),
			)
		}

		c.Locals("uploaded_file", file)
		return c.Next()
	}
}
