package i18n

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"fiber-starter/app/helpers"
)

// TranslateValidationErrors 翻译验证错误
// 返回字段名到错误消息的映射
func TranslateValidationErrors(err error, t *Translator) map[string]string {
	errors := make(map[string]string)

	if err == nil {
		return errors
	}

	// 检查是否是验证错误
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 如果不是验证错误，返回原始错误消息
		errors["error"] = err.Error()
		return errors
	}

	// 翻译每个验证错误
	for _, fieldError := range validationErrors {
		// 获取字段名
		fieldName := getFieldName(fieldError)

		// 翻译字段名
		translatedFieldName := GetFieldName(fieldName, t)

		// 翻译错误消息
		errorMessage := translateFieldError(fieldError, translatedFieldName, t)

		errors[fieldName] = errorMessage
	}

	return errors
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

// getDefaultErrorMessage 获取默认错误消息（当翻译不存在时）
func getDefaultErrorMessage(tag, fieldName, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldName)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fieldName, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fieldName, param)
	case "len":
		return fmt.Sprintf("%s must be %s characters", fieldName, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldName, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldName, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fieldName, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fieldName, param)
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", fieldName, param)
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", fieldName, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fieldName, param)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fieldName)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", fieldName)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", fieldName)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", fieldName)
	case "numeric":
		return fmt.Sprintf("%s must be a valid numeric value", fieldName)
	case "number":
		return fmt.Sprintf("%s must be a valid number", fieldName)
	case "hexadecimal":
		return fmt.Sprintf("%s must be a valid hexadecimal", fieldName)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", fieldName)
	case "rgb":
		return fmt.Sprintf("%s must be a valid RGB color", fieldName)
	case "rgba":
		return fmt.Sprintf("%s must be a valid RGBA color", fieldName)
	case "hsl":
		return fmt.Sprintf("%s must be a valid HSL color", fieldName)
	case "hsla":
		return fmt.Sprintf("%s must be a valid HSLA color", fieldName)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldName)
	case "uuid3":
		return fmt.Sprintf("%s must be a valid UUID v3", fieldName)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", fieldName)
	case "uuid5":
		return fmt.Sprintf("%s must be a valid UUID v5", fieldName)
	case "isbn":
		return fmt.Sprintf("%s must be a valid ISBN", fieldName)
	case "isbn10":
		return fmt.Sprintf("%s must be a valid ISBN-10", fieldName)
	case "isbn13":
		return fmt.Sprintf("%s must be a valid ISBN-13", fieldName)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", fieldName)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", fieldName)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", fieldName)
	case "ssn":
		return fmt.Sprintf("%s must be a valid SSN", fieldName)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", fieldName)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", fieldName)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", fieldName)
	case "cidr":
		return fmt.Sprintf("%s must be a valid CIDR notation", fieldName)
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", fieldName)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", fieldName)
	default:
		return fmt.Sprintf("%s failed validation for '%s'", fieldName, tag)
	}
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
func RegisterValidatorTranslations(v *validator.Validate) {
	// 这里可以注册自定义验证规则的翻译
	helpers.Info("验证器翻译已注册")
}

// GetValidationErrorsAsString 获取验证错误的字符串表示
func GetValidationErrorsAsString(err error, t *Translator) string {
	errors := TranslateValidationErrors(err, t)

	var messages []string
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
