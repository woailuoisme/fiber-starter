package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"fiber-starter/app/exceptions"
	helpers "fiber-starter/app/http/requests"

	"github.com/gofiber/fiber/v3"
)

// ValidationMiddleware 验证中间件
// Requirements: 10.1, 10.6, 10.7
func ValidationMiddleware(model interface{}) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 解析请求体
		if err := c.Bind().Body(model); err != nil {
			return exceptions.NewBadRequestException("Invalid request body")
		}

		// 验证模型
		if err := helpers.ValidateStruct(model); err != nil {
			return err
		}

		// 将验证后的模型存储到上下文中
		c.Locals("validated_model", model)

		return c.Next()
	}
}

// GetValidatedModel 从上下文中获取验证后的模型
func GetValidatedModel(c fiber.Ctx, model interface{}) bool {
	if validatedModel := c.Locals("validated_model"); validatedModel != nil {
		// 使用反射复制值
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

func validateQueryField(field, value, rule string) []string {
	var errors []string

	// 检查必填字段
	if strings.Contains(rule, "required") && value == "" {
		errors = append(errors, fmt.Sprintf("The %s field is required.", field))
		return errors
	}

	// 如果字段为空且不是必填，跳过其他验证
	if value == "" {
		return errors
	}

	// 验证数字
	if strings.Contains(rule, "number") {
		var num int64
		if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
			errors = append(errors, fmt.Sprintf("The %s must be a number.", field))
		}
	}

	// 验证最小长度
	if strings.Contains(rule, "min:") {
		minLength := 0
		_, _ = fmt.Sscanf(strings.Split(rule, "min:")[1], "%d", &minLength)
		if len(value) < minLength {
			errors = append(errors, fmt.Sprintf("The %s must be at least %d characters.", field, minLength))
		}
	}

	// 验证最大长度
	if strings.Contains(rule, "max:") {
		maxLength := 0
		_, _ = fmt.Sscanf(strings.Split(rule, "max:")[1], "%d", &maxLength)
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

		// 验证文件大小
		// Requirements: 8.2
		if file.Size > maxSize {
			return exceptions.NewBadRequestException(
				fmt.Sprintf("File size must not exceed %d MB", maxSize/(1024*1024)),
			)
		}

		// 验证文件类型
		// Requirements: 8.3
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

		// 将文件信息存储到上下文中
		c.Locals("uploaded_file", file)

		return c.Next()
	}
}
