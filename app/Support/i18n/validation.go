package i18n

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

var fallbackValidationMessages = map[string]string{
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

// FormatValidationErrors formats validation errors using the default language.
func FormatValidationErrors(err error) map[string][]string {
	return formatValidationErrors(nil, err)
}

// FormatValidationErrorsWithContext formats validation errors using the request language.
func FormatValidationErrorsWithContext(c fiber.Ctx, err error) map[string][]string {
	return formatValidationErrors(c, err)
}

// FormatValidationErrorsToString converts validation errors to a semicolon-separated string.
func FormatValidationErrorsToString(err error) string {
	validationErrors := FormatValidationErrors(err)
	if len(validationErrors) == 0 {
		if err == nil {
			return ""
		}
		return err.Error()
	}

	parts := make([]string, 0, len(validationErrors))
	for field, messages := range validationErrors {
		for _, message := range messages {
			parts = append(parts, field+": "+message)
		}
	}

	return strings.Join(parts, "; ")
}

// GetFirstValidationError returns the first formatted validation error message.
func GetFirstValidationError(err error) string {
	messages := validationMessages(nil, err)
	if len(messages) == 0 {
		if err == nil {
			return ""
		}
		return err.Error()
	}
	return messages[0]
}

// GetFirstValidationErrorWithContext returns the first localized validation error message.
func GetFirstValidationErrorWithContext(c fiber.Ctx, err error) string {
	messages := validationMessages(c, err)
	if len(messages) == 0 {
		if err == nil {
			return ""
		}
		return err.Error()
	}
	return messages[0]
}

func formatValidationErrors(c fiber.Ctx, err error) map[string][]string {
	errMap := make(map[string][]string)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		if err != nil {
			errMap["error"] = []string{err.Error()}
		}
		return errMap
	}

	for _, fe := range validationErrors {
		fieldKey := normalizeFieldKey(fe.Field())
		fieldName := localizedFieldName(c, fieldKey)
		message := localizedValidationMessage(c, fe, fieldName)
		errMap[fieldKey] = append(errMap[fieldKey], message)
	}

	return errMap
}

func validationMessages(c fiber.Ctx, err error) []string {
	validationErrors := formatValidationErrors(c, err)
	messages := make([]string, 0, len(validationErrors))
	for _, msg := range validationErrors {
		messages = append(messages, msg...)
	}
	return messages
}

func localizedFieldName(c fiber.Ctx, fieldKey string) string {
	if fieldKey == "" {
		return fieldKey
	}

	messageID := fmt.Sprintf("fields.%s", fieldKey)
	fallback := toFriendlyName(fieldKey)
	if translated, err := localize(c, messageID, fallback, nil); err == nil && translated != "" {
		return translated
	}

	return fallback
}

func localizedValidationMessage(c fiber.Ctx, fe validator.FieldError, fieldName string) string {
	tag := fe.Tag()
	messageID := fmt.Sprintf("validation.%s", tag)
	fallback := fallbackValidationMessage(tag, fieldName, fe.Param())

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

	translated, err := localize(c, messageID, fallback, data)
	if err != nil || translated == "" {
		return fallback
	}

	return translated
}

func localize(c fiber.Ctx, messageID, fallback string, data map[string]interface{}) (string, error) {
	if c == nil {
		return fallback, nil
	}

	svc := Default()
	if svc == nil {
		return fallback, nil
	}

	translated, err := svc.Localize(c, &goi18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &goi18n.Message{
			ID:    messageID,
			Other: fallback,
		},
		TemplateData: data,
	})
	if err != nil || translated == "" || translated == messageID {
		return fallback, err
	}

	return translated, nil
}

func fallbackValidationMessage(tag, fieldName, param string) string {
	format, ok := fallbackValidationMessages[tag]
	if !ok {
		return fmt.Sprintf("%s failed validation for '%s'", fieldName, tag)
	}

	if strings.Count(format, "%s") == 2 {
		return fmt.Sprintf(format, fieldName, param)
	}

	return fmt.Sprintf(format, fieldName)
}

func normalizeFieldKey(field string) string {
	field = strings.TrimSpace(field)
	if field == "" {
		return ""
	}

	if strings.Contains(field, "_") {
		return strings.ToLower(field)
	}

	var b strings.Builder
	for i, r := range field {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}

	return strings.ToLower(b.String())
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
