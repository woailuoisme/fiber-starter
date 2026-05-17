// Package requests 处理HTTP请求验证逻辑
package requests

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	exceptions "fiber-starter/app/Exceptions"
	supporti18n "fiber-starter/app/Providers/I18n"
	"fiber-starter/app/Providers/Validation/Contracts"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// ValidatorFactory 全局验证工厂实例。
var ValidatorFactory Contracts.Factory

// InitValidator 初始化验证器。
func InitValidator(v Contracts.Factory) {
	ValidatorFactory = v
}

// ValidateStruct 验证结构体。
// Requirements: 10.1, 10.6, 10.7
func ValidateStruct(s interface{}) error {
	return ValidateStructWithContext(nil, s)
}

// ValidateStructWithContext 验证结构体并使用请求语言格式化错误。
// Requirements: 10.1, 10.6, 10.7
func ValidateStructWithContext(c fiber.Ctx, s interface{}) error {
	if ValidatorFactory == nil {
		return errors.New("validator not initialized")
	}

	err := ValidatorFactory.Make(s, nil, nil, nil).Validate()
	if err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			errors := supporti18n.FormatValidationErrorsWithContext(c, validationErrors)
			return exceptions.NewValidationExceptionWithErrors("Validation failed", errors)
		}
		return exceptions.NewValidationException(err.Error())
	}

	return nil
}

// ValidateRequest 验证请求并解析到结构体
// Requirements: 10.1, 10.6, 10.7
func ValidateRequest(c fiber.Ctx, req interface{}) error {
	if err := c.Bind().Body(req); err != nil {
		return exceptions.NewBadRequestException("Invalid request body")
	}

	return ValidateStructWithContext(c, req)
}

func validateScalar(value any, rule string) bool {
	if rule == "" {
		return false
	}

	if ValidatorFactory != nil {
		return ValidatorFactory.Make(map[string]any{"value": value}, map[string]string{"value": rule}, nil, nil).Validate() == nil
	}

	v := validator.New()
	return v.Var(value, rule) == nil
}

func validateRuleSet(c fiber.Ctx, values map[string]any, rules map[string]string) error {
	if ValidatorFactory == nil {
		return errors.New("validator not initialized")
	}

	validationErrors := make(map[string][]string, len(rules))
	for field, rule := range rules {
		value := values[field]
		err := ValidatorFactory.Make(map[string]any{field: value}, map[string]string{field: rule}, nil, nil).Validate()
		if err == nil {
			continue
		}

		if validationErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			formatted := supporti18n.FormatValidationErrorsWithContext(c, validationErrs)
			if messages, ok := formatted[field]; ok {
				validationErrors[field] = append(validationErrors[field], messages...)
				continue
			}
			if messages, ok := formatted[""]; ok {
				validationErrors[field] = append(validationErrors[field], messages...)
				continue
			}
			for _, messages := range formatted {
				validationErrors[field] = append(validationErrors[field], messages...)
			}
			continue
		}

		validationErrors[field] = append(validationErrors[field], err.Error())
	}

	if len(validationErrors) > 0 {
		return exceptions.NewValidationExceptionWithErrors("Validation failed", validationErrors)
	}

	return nil
}

// ValidateEmail 验证邮箱格式
// Requirements: 10.2
func ValidateEmail(email string) bool {
	return validateScalar(email, "required,email")
}

// ValidateURL 验证 URL 格式。
func ValidateURL(url string) bool {
	return validateScalar(url, "required,url")
}

// ValidateE164 验证 E.164 电话号码。
func ValidateE164(phone string) bool {
	return strings.HasPrefix(phone, "+") && validateScalar(phone, "required,e164")
}

// ValidatePassword 验证密码强度
// Requirements: 10.3
func ValidatePassword(password string) bool {
	return validateScalar(password, "required,min=8")
}

// ValidateRequired 验证必填字段
// Requirements: 10.1
func ValidateRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ValidateMinLength 验证最小长度
func ValidateMinLength(value string, minLength int) bool {
	return validateScalar(value, fmt.Sprintf("min=%d", minLength))
}

// ValidateMaxLength 验证最大长度
func ValidateMaxLength(value string, maxLength int) bool {
	return validateScalar(value, fmt.Sprintf("max=%d", maxLength))
}

// ValidateRange 验证数值范围
func ValidateRange(value, minValue, maxValue int) bool {
	return validateScalar(value, fmt.Sprintf("gte=%d,lte=%d", minValue, maxValue))
}

// ValidatePositiveNumber 验证正数
// Requirements: 10.4
func ValidatePositiveNumber(value int) bool {
	return validateScalar(value, "gt=0")
}

// ValidatePositiveInteger 验证正整数
// Requirements: 10.5
func ValidatePositiveInteger(value int) bool {
	return validateScalar(value, "gt=0")
}

// ValidateQueryRules 在请求层校验 query 规则。
// 作用：替代 middleware/query validation 的旧实现。
func ValidateQueryRules(c fiber.Ctx, rules map[string]string) error {
	values := make(map[string]any, len(rules))
	for field := range rules {
		values[field] = coerceQueryValue(c.Query(field), rules[field])
	}

	return validateRuleSet(c, values, rules)
}

// ValidateUploadedFile 校验上传文件。
// 作用：替代 middleware/file validation 的旧实现。
func ValidateUploadedFile(c fiber.Ctx, field string, maxSize int64, allowedTypes []string) (*multipart.FileHeader, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return nil, exceptions.BadRequestWithDetails("File upload failed", err.Error())
	}

	validationErrors := make(map[string][]string)
	if file.Filename == "" {
		validationErrors[field] = append(validationErrors[field], fmt.Sprintf("%s must be a valid uploaded file", field))
	}
	if maxSize > 0 && file.Size > maxSize {
		validationErrors[field] = append(validationErrors[field], fmt.Sprintf("%s must not exceed the maximum file size", field))
	}
	if len(allowedTypes) > 0 && !mimeTypeAllowed(file.Header.Get("Content-Type"), allowedTypes) {
		validationErrors[field] = append(validationErrors[field], fmt.Sprintf("%s must be a valid file type", field))
	}

	if len(validationErrors) > 0 {
		return nil, exceptions.NewValidationExceptionWithErrors("Validation failed", validationErrors)
	}

	return file, nil
}

func coerceQueryValue(value string, rule string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	rule = strings.ToLower(rule)
	if strings.Contains(rule, "boolean") {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}

	if strings.Contains(rule, "number") || strings.Contains(rule, "numeric") || strings.Contains(rule, "integer") ||
		strings.Contains(rule, "gt=") || strings.Contains(rule, "gte=") || strings.Contains(rule, "lt=") || strings.Contains(rule, "lte=") {
		if strings.Contains(value, ".") {
			if parsed, err := strconv.ParseFloat(value, 64); err == nil {
				return parsed
			}
		} else if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}

	return value
}

func mimeTypeAllowed(contentType string, allowedTypes []string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if contentType == "" {
		return false
	}

	for _, allowedType := range allowedTypes {
		allowedType = strings.TrimSpace(strings.ToLower(allowedType))
		if allowedType == "" {
			continue
		}
		if allowedType == contentType {
			return true
		}
		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(contentType, prefix+"/") {
				return true
			}
		}
	}

	return false
}
