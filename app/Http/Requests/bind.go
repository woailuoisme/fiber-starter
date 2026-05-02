package requests

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	exceptions "fiber-starter/app/Exceptions"
	supporti18n "fiber-starter/app/Support/i18n"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// BindAndValidateBody 绑定请求体并执行结构体校验。
func BindAndValidateBody(c fiber.Ctx, req interface{}) error {
	if err := c.Bind().Body(req); err != nil {
		return exceptions.BadRequestWithDetails("Invalid request body", err.Error())
	}

	return validateStructAsAppError(c, req)
}

// BindAndValidateQuery 绑定查询参数并执行结构体校验。
func BindAndValidateQuery(c fiber.Ctx, req interface{}) error {
	if err := c.Bind().Query(req); err != nil {
		return exceptions.BadRequestWithDetails("Invalid query parameters", err.Error())
	}

	return validateStructAsAppError(c, req)
}

// ValidateQueryRules 在请求层校验 query 规则。
// 作用：替代 middleware/query validation 的旧实现。
func ValidateQueryRules(c fiber.Ctx, rules map[string]string) error {
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

	return nil
}

// ValidateUploadedFile 校验上传文件。
// 作用：替代 middleware/file validation 的旧实现。
func ValidateUploadedFile(c fiber.Ctx, field string, maxSize int64, allowedTypes []string) (*multipart.FileHeader, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return nil, exceptions.BadRequestWithDetails("File upload failed", err.Error())
	}

	if file.Size > maxSize {
		return nil, exceptions.BadRequestWithDetails(
			"File size must not exceed limit",
			fmt.Sprintf("%d MB", maxSize/(1024*1024)),
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
		return nil, exceptions.BadRequestWithDetails("Unsupported file type", contentType)
	}

	return file, nil
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

func validateStructAsAppError(c fiber.Ctx, s interface{}) error {
	if Validator == nil {
		InitValidator()
	}

	err := Validator.Struct(s)
	if err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			return exceptions.NewValidationExceptionWithErrors("Validation failed", supporti18n.FormatValidationErrorsWithContext(c, validationErrors))
		}
		return exceptions.Validation(err.Error())
	}

	return nil
}
