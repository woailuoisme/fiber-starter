// Package requests 处理HTTP请求验证逻辑
package requests

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	exceptions "fiber-starter/app/Exceptions"
	appresources "fiber-starter/app/Http/Resources"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// Validator 全局验证器实例
// Requirements: 22.9
var Validator *validator.Validate

// InitValidator 初始化验证器
// Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7
func InitValidator() {
	Validator = validator.New()

	registerCustomValidations()
}

// registerCustomValidations 注册自定义验证规则
func registerCustomValidations() {
	_ = Validator.RegisterValidation("phone", validatePhone)
	_ = Validator.RegisterValidation("positive_int", validatePositiveInt)
	_ = Validator.RegisterValidation("positive", validatePositive)
	_ = Validator.RegisterValidation("price", validatePrice)
}

// validatePhone 验证手机号
// Requirements: 10.2
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	matched, _ := regexp.MatchString(`^\d{10,15}$`, phone)
	return matched
}

// validatePositiveInt 验证正整数
// Requirements: 10.5
func validatePositiveInt(fl validator.FieldLevel) bool {
	value := fl.Field().Int()
	return value > 0
}

// validatePositive 验证正数
// Requirements: 10.4
func validatePositive(fl validator.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fl.Field().Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fl.Field().Uint() > 0
	case reflect.Float32, reflect.Float64:
		return fl.Field().Float() > 0
	default:
		return false
	}
}

// validatePrice 验证价格（正整数）
// Requirements: 10.4
func validatePrice(fl validator.FieldLevel) bool {
	value := fl.Field().Int()
	return value > 0
}

// ValidateStruct 验证结构体
// Requirements: 10.1, 10.6, 10.7
func ValidateStruct(s interface{}) error {
	if Validator == nil {
		InitValidator()
	}

	err := Validator.Struct(s)
	if err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			errors := appresources.FormatValidationErrors(validationErrors)
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

	return ValidateStruct(req)
}

// ValidateEmail 验证邮箱格式
// Requirements: 10.2
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

// ValidatePassword 验证密码强度
// Requirements: 10.3
func ValidatePassword(password string) bool {
	return len(password) >= 8
}

// ValidateRequired 验证必填字段
// Requirements: 10.1
func ValidateRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ValidateMinLength 验证最小长度
func ValidateMinLength(value string, minLength int) bool {
	return len(value) >= minLength
}

// ValidateMaxLength 验证最大长度
func ValidateMaxLength(value string, maxLength int) bool {
	return len(value) <= maxLength
}

// ValidateRange 验证数值范围
func ValidateRange(value, minValue, maxValue int) bool {
	return value >= minValue && value <= maxValue
}

// ValidatePositiveNumber 验证正数
// Requirements: 10.4
func ValidatePositiveNumber(value int) bool {
	return value > 0
}

// ValidatePositiveInteger 验证正整数
// Requirements: 10.5
func ValidatePositiveInteger(value int) bool {
	return value > 0
}
