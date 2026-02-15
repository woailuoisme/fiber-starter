package i18n

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"fiber-starter/app/helpers"

	"github.com/go-playground/validator/v10"
)

// TranslateValidationErrors 翻译验证错误
// 返回字段名到错误消息的映射
func TranslateValidationErrors(err error, t *Translator) map[string]string {
	errMap := make(map[string]string)

	if err == nil {
		return errMap
	}

	// 检查是否是验证错误
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		// 如果不是验证错误，返回原始错误消息
		errMap["error"] = err.Error()
		return errMap
	}

	// 翻译每个验证错误
	for _, fieldError := range validationErrors {
		// 获取字段名
		fieldName := getFieldName(fieldError)

		// 翻译字段名
		translatedFieldName := GetFieldName(fieldName, t)

		// 翻译错误消息
		errorMessage := translateFieldError(fieldError, translatedFieldName, t)

		errMap[fieldName] = errorMessage
	}

	return errMap
}

// translateFieldError 翻译单个字段错误
func translateFieldError(fe validator.FieldError, fieldName string, t *Translator) string {
	tag := fe.Tag()

	// 构建翻译键
	messageID := fmt.Sprintf("validation.%s", tag)

	// 准备模板数据
	data := map[string]interface{}{
		"Field": fieldName,
		"Value": fe.Value(),
		"Param": fe.Param(),
	}

	// 根据不同的验证规则添加特定参数
	switch tag {
	case "min", "max", "len":
		data["Min"] = fe.Param()
		data["Max"] = fe.Param()
		data["Length"] = fe.Param()
	case "gte", "lte", "gt", "lt":
		data["Number"] = fe.Param()
	case "oneof":
		data["Options"] = fe.Param()
	}

	// 尝试翻译
	translation := t.TWithData(messageID, data)

	// 如果翻译不存在，使用默认消息
	if translation == messageID {
		return getDefaultErrorMessage(tag, fieldName, fe.Param())
	}

	return translation
}

// defaultErrorMessages 默认错误消息模板
var defaultErrorMessages = map[string]string{
	"required":    "%s is required",
	"email":       "%s must be a valid email address",
	"min":         "%s must be at least %s characters",
	"max":         "%s must be at most %s characters",
	"len":         "%s must be %s characters",
	"gte":         "%s must be greater than or equal to %s",
	"lte":         "%s must be less than or equal to %s",
	"gt":          "%s must be greater than %s",
	"lt":          "%s must be less than %s",
	"eqfield":     "%s must be equal to %s",
	"nefield":     "%s must not be equal to %s",
	"oneof":       "%s must be one of [%s]",
	"url":         "%s must be a valid URL",
	"uri":         "%s must be a valid URI",
	"alpha":       "%s must contain only alphabetic characters",
	"alphanum":    "%s must contain only alphanumeric characters",
	"numeric":     "%s must be a valid numeric value",
	"number":      "%s must be a valid number",
	"hexadecimal": "%s must be a valid hexadecimal",
	"hexcolor":    "%s must be a valid hex color",
	"rgb":         "%s must be a valid RGB color",
	"rgba":        "%s must be a valid RGBA color",
	"hsl":         "%s must be a valid HSL color",
	"hsla":        "%s must be a valid HSLA color",
	"uuid":        "%s must be a valid UUID",
	"uuid3":       "%s must be a valid UUID v3",
	"uuid4":       "%s must be a valid UUID v4",
	"uuid5":       "%s must be a valid UUID v5",
	"isbn":        "%s must be a valid ISBN",
	"isbn10":      "%s must be a valid ISBN-10",
	"isbn13":      "%s must be a valid ISBN-13",
	"json":        "%s must be valid JSON",
	"latitude":    "%s must be a valid latitude",
	"longitude":   "%s must be a valid longitude",
	"ssn":         "%s must be a valid SSN",
	"ipv4":        "%s must be a valid IPv4 address",
	"ipv6":        "%s must be a valid IPv6 address",
	"ip":          "%s must be a valid IP address",
	"cidr":        "%s must be a valid CIDR notation",
	"mac":         "%s must be a valid MAC address",
	"datetime":    "%s must be a valid datetime",
}

// getDefaultErrorMessage 获取默认错误消息（当翻译不存在时）
func getDefaultErrorMessage(tag, fieldName, param string) string {
	if format, ok := defaultErrorMessages[tag]; ok {
		// 根据格式化占位符数量决定参数
		if strings.Count(format, "%s") == 2 {
			return fmt.Sprintf(format, fieldName, param)
		}
		return fmt.Sprintf(format, fieldName)
	}
	return fmt.Sprintf("%s failed validation for '%s'", fieldName, tag)
}

// GetFieldName 获取字段的翻译名称
func GetFieldName(field string, t *Translator) string {
	// 构建翻译键
	messageID := fmt.Sprintf("fields.%s", strings.ToLower(field))

	// 尝试翻译
	translation := t.T(messageID)

	// 如果翻译不存在，返回原字段名（转换为友好格式）
	if translation == messageID {
		return toFriendlyName(field)
	}

	return translation
}

// getFieldName 从 FieldError 获取字段名
func getFieldName(fe validator.FieldError) string {
	// 优先使用 JSON 标签名
	field := fe.Field()

	// 如果有结构体标签，尝试获取 json 标签
	if fe.StructField() != "" {
		// 这里简化处理，实际使用中可能需要反射获取标签
		field = fe.Field()
	}

	return field
}

// toFriendlyName 将字段名转换为友好格式
// 例如：UserName -> User Name, user_name -> User Name
func toFriendlyName(field string) string {
	// 处理下划线分隔
	if strings.Contains(field, "_") {
		parts := strings.Split(field, "_")
		for i, part := range parts {
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
		return strings.Join(parts, " ")
	}

	// 处理驼峰命名
	var result []rune
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, ' ')
		}
		result = append(result, r)
	}

	return string(result)
}

// RegisterValidatorTranslations 注册验证器翻译（预留接口）
func RegisterValidatorTranslations(_ *validator.Validate) {
	// 这里可以注册自定义验证规则的翻译
	helpers.Info("验证器翻译已注册")
}

// GetValidationErrorsAsString 获取验证错误的字符串表示
func GetValidationErrorsAsString(err error, t *Translator) string {
	errors := TranslateValidationErrors(err, t)

	messages := make([]string, 0, len(errors))
	for _, msg := range errors {
		messages = append(messages, msg)
	}

	return strings.Join(messages, "; ")
}

// GetFirstValidationError 获取第一个验证错误
func GetFirstValidationError(err error, t *Translator) string {
	errors := TranslateValidationErrors(err, t)

	for _, msg := range errors {
		return msg
	}

	return ""
}

// ValidateStruct 验证结构体并返回翻译后的错误
func ValidateStruct(v *validator.Validate, s interface{}, t *Translator) map[string]string {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	return TranslateValidationErrors(err, t)
}

// GetStructFieldName 通过反射获取结构体字段的 JSON 标签名
func GetStructFieldName(structType reflect.Type, fieldName string) string {
	field, found := structType.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	// 获取 json 标签
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		// 处理 json 标签中的选项（如 omitempty）
		parts := strings.Split(jsonTag, ",")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	return fieldName
}
