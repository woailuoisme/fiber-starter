package i18n

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	helpers "fiber-starter/app/Support"

	"github.com/go-playground/validator/v10"
)

func TranslateValidationErrors(err error, t *Translator) map[string]string {
	errMap := make(map[string]string)

	if err == nil {
		return errMap
	}

	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		errMap["error"] = err.Error()
		return errMap
	}

	for _, fieldError := range validationErrors {
		fieldName := getFieldName(fieldError)
		errMap[fieldName] = translateFieldError(fieldError, translatedFieldName(fieldName, t), t)
	}

	return errMap
}

func translatedFieldName(field string, t *Translator) string {
	if t == nil {
		return toFriendlyName(field)
	}

	return GetFieldName(field, t)
}

func translateFieldError(fe validator.FieldError, fieldName string, t *Translator) string {
	tag := fe.Tag()
	messageID := fmt.Sprintf("validation.%s", tag)

	data := map[string]interface{}{
		"Field": fieldName,
		"Value": fe.Value(),
		"Param": fe.Param(),
	}

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

	translation := messageID
	if t != nil {
		translation = t.TWithData(messageID, data)
	}

	if translation == messageID {
		return getDefaultErrorMessage(tag, fieldName, fe.Param())
	}

	return translation
}

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

func getDefaultErrorMessage(tag, fieldName, param string) string {
	if format, ok := defaultErrorMessages[tag]; ok {
		if strings.Count(format, "%s") == 2 {
			return fmt.Sprintf(format, fieldName, param)
		}
		return fmt.Sprintf(format, fieldName)
	}
	return fmt.Sprintf("%s failed validation for '%s'", fieldName, tag)
}

func GetFieldName(field string, t *Translator) string {
	if t == nil {
		return toFriendlyName(field)
	}

	messageID := fmt.Sprintf("fields.%s", strings.ToLower(field))
	translation := t.T(messageID)
	if translation == messageID {
		return toFriendlyName(field)
	}

	return translation
}

func getFieldName(fe validator.FieldError) string {
	return fe.Field()
}

func toFriendlyName(field string) string {
	if strings.Contains(field, "_") {
		parts := strings.Split(field, "_")
		for i, part := range parts {
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
		return strings.Join(parts, " ")
	}

	var result []rune
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, ' ')
		}
		result = append(result, r)
	}

	return string(result)
}

func RegisterValidatorTranslations(_ *validator.Validate) {
	helpers.Info("Validator translations registered")
}

func GetValidationErrorsAsString(err error, t *Translator) string {
	return strings.Join(validationMessages(err, t), "; ")
}

func GetFirstValidationError(err error, t *Translator) string {
	messages := validationMessages(err, t)
	if len(messages) == 0 {
		return ""
	}
	return messages[0]
}

func ValidateStruct(v *validator.Validate, s interface{}, t *Translator) map[string]string {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	return TranslateValidationErrors(err, t)
}

func GetStructFieldName(structType reflect.Type, fieldName string) string {
	field, found := structType.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	return fieldName
}

func validationMessages(err error, t *Translator) []string {
	validationErrors := TranslateValidationErrors(err, t)
	messages := make([]string, 0, len(validationErrors))
	for _, msg := range validationErrors {
		messages = append(messages, msg)
	}
	return messages
}
