package i18n

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var fallbackValidationMessages = map[string]string{
	"accepted":                  "%s must be accepted",
	"accepted_if":               "%s must be accepted when %s",
	"declined":                  "%s must be declined",
	"declined_if":               "%s must be declined when %s",
	"active_url":                "%s must be a valid active URL",
	"after":                     "%s must be a date after %s",
	"after_or_equal":            "%s must be a date after or equal to %s",
	"required":                  "%s is required",
	"required_if":               "%s is required when %s",
	"required_if_accepted":      "%s is required when %s is accepted",
	"required_if_declined":      "%s is required when %s is declined",
	"required_unless":           "%s is required unless %s",
	"required_with":             "%s is required when %s is present",
	"required_with_all":         "%s is required when all of [%s] are present",
	"required_without":          "%s is required when %s is missing",
	"required_without_all":      "%s is required when all of [%s] are missing",
	"array":                     "%s must be an array",
	"array_keys":                "%s contains invalid keys",
	"required_array_keys":       "%s must contain the required keys",
	"distinct":                  "%s contains duplicate values",
	"in_array":                  "%s must be present in %s",
	"list":                      "%s must be a list",
	"filled":                    "%s must not be empty",
	"present":                   "%s must be present",
	"nullable":                  "%s may be null",
	"missing":                   "%s must be missing",
	"missing_if":                "%s must be missing when %s",
	"missing_unless":            "%s must be missing unless %s",
	"missing_with":              "%s must be missing when %s is present",
	"missing_with_all":          "%s must be missing when all of [%s] are present",
	"prohibited":                "%s is prohibited",
	"prohibited_if":             "%s is prohibited when %s",
	"prohibited_unless":         "%s is prohibited unless %s",
	"prohibited_with":           "%s is prohibited when %s is present",
	"prohibited_with_all":       "%s is prohibited when all of [%s] are present",
	"prohibited_without":        "%s is prohibited when %s is missing",
	"prohibited_without_all":    "%s is prohibited when all of [%s] are missing",
	"prohibits":                 "%s prohibits %s",
	"present_if":                "%s must be present when %s",
	"present_unless":            "%s must be present unless %s",
	"present_with":              "%s must be present when %s is present",
	"present_with_all":          "%s must be present when all of [%s] are present",
	"exclude_if":                "%s is excluded when %s",
	"exclude_unless":            "%s is excluded unless %s",
	"exclude_with":              "%s is excluded when %s is present",
	"exclude_without":           "%s is excluded when %s is missing",
	"email":                     "%s must be a valid email address",
	"confirmed":                 "%s confirmation does not match",
	"different":                 "%s must be different from %s",
	"same":                      "%s must match %s",
	"min":                       "%s must be at least %s characters",
	"max":                       "%s must be at most %s characters",
	"len":                       "%s must be %s characters",
	"between":                   "%s must be between %s and %s",
	"size":                      "%s must be %s",
	"gte":                       "%s must be greater than or equal to %s",
	"lte":                       "%s must be less than or equal to %s",
	"gt":                        "%s must be greater than %s",
	"lt":                        "%s must be less than %s",
	"eqfield":                   "%s must be equal to %s",
	"nefield":                   "%s must not be equal to %s",
	"oneof":                     "%s must be one of [%s]",
	"url":                       "%s must be a valid URL",
	"uri":                       "%s must be a valid URI",
	"alpha":                     "%s must contain only alphabetic characters",
	"alpha_dash":                "%s must contain only letters, numbers, dashes, or underscores",
	"alpha_num":                 "%s must contain only alphanumeric characters",
	"alphaunicode":              "%s must contain only alphabetic Unicode characters",
	"alphanum":                  "%s must contain only alphanumeric characters",
	"alphanumunicode":           "%s must contain only alphanumeric Unicode characters",
	"alpha_space":               "%s must contain only alphabetic characters and spaces",
	"alphanumspace":             "%s must contain only alphanumeric characters and spaces",
	"ascii":                     "%s must contain only ASCII characters",
	"printascii":                "%s must contain only printable ASCII characters",
	"base64":                    "%s must be valid base64",
	"base64url":                 "%s must be valid base64url",
	"base64rawurl":              "%s must be valid raw base64url",
	"contains":                  "%s must contain the required substring",
	"containsany":               "%s must contain at least one required substring",
	"containsrune":              "%s must contain the required rune",
	"endswith":                  "%s must end with the required suffix",
	"ends_with":                 "%s must end with the required suffix",
	"endsnotwith":               "%s must not end with the forbidden suffix",
	"doesnt_end_with":           "%s must not end with the forbidden suffix",
	"startswith":                "%s must start with the required prefix",
	"starts_with":               "%s must start with the required prefix",
	"startsnotwith":             "%s must not start with the forbidden prefix",
	"doesnt_start_with":         "%s must not start with the forbidden prefix",
	"excludes":                  "%s must not contain the forbidden substring",
	"excludesall":               "%s must not contain any forbidden substrings",
	"excludesrune":              "%s must not contain the forbidden rune",
	"numeric":                   "%s must be a valid numeric value",
	"number":                    "%s must be a valid number",
	"integer":                   "%s must be an integer",
	"date_format":               "%s must match the expected date format",
	"date_equals":               "%s must be a date equal to %s",
	"digits":                    "%s must have %s digits",
	"digits_between":            "%s must have between %s and %s digits",
	"decimal":                   "%s must be a decimal number",
	"multiple_of":               "%s must be a multiple of %s",
	"regex":                     "%s format is invalid",
	"not_regex":                 "%s format is invalid",
	"in":                        "%s must be one of the allowed values",
	"not_in":                    "%s must not be one of the disallowed values",
	"string":                    "%s must be a string",
	"lowercase":                 "%s must contain only lowercase characters",
	"uppercase":                 "%s must contain only uppercase characters",
	"hexadecimal":               "%s must be a valid hexadecimal",
	"hexcolor":                  "%s must be a valid hex color",
	"rgb":                       "%s must be a valid RGB color",
	"rgba":                      "%s must be a valid RGBA color",
	"hsl":                       "%s must be a valid HSL color",
	"hsla":                      "%s must be a valid HSLA color",
	"uuid":                      "%s must be a valid UUID",
	"uuid3":                     "%s must be a valid UUID v3",
	"uuid4":                     "%s must be a valid UUID v4",
	"uuid5":                     "%s must be a valid UUID v5",
	"isbn":                      "%s must be a valid ISBN",
	"isbn10":                    "%s must be a valid ISBN-10",
	"isbn13":                    "%s must be a valid ISBN-13",
	"json":                      "%s must be valid JSON",
	"cron":                      "%s must be a valid cron expression",
	"cve":                       "%s must be a valid CVE identifier",
	"semver":                    "%s must be a valid semantic version",
	"ulid":                      "%s must be a valid ULID",
	"dir":                       "%s must be a valid directory path",
	"dirpath":                   "%s must be a valid directory path",
	"file":                      "%s must be a valid file path",
	"uploaded_file":             "%s must be a valid uploaded file",
	"current_password":          "%s must be a valid current password",
	"filepath":                  "%s must be a valid file path",
	"image":                     "%s must be a valid image path",
	"fqdn":                      "%s must be a valid FQDN",
	"fqdn_rfc1123":              "%s must be a valid RFC1123 FQDN",
	"luhn_checksum":             "%s must pass the Luhn checksum",
	"mongodb":                   "%s must be a valid MongoDB ObjectID",
	"mongodb_connection_string": "%s must be a valid MongoDB connection string",
	"bcp47_language_tag":        "%s must be a valid BCP 47 language tag",
	"bcp47_language_tag_loose":  "%s must be a valid BCP 47 language tag",
	"latitude":                  "%s must be a valid latitude",
	"longitude":                 "%s must be a valid longitude",
	"ssn":                       "%s must be a valid SSN",
	"ipv4":                      "%s must be a valid IPv4 address",
	"ipv6":                      "%s must be a valid IPv6 address",
	"ip":                        "%s must be a valid IP address",
	"cidr":                      "%s must be a valid CIDR notation",
	"mac":                       "%s must be a valid MAC address",
	"mac_address":               "%s must be a valid MAC address",
	"e164":                      "%s must be a valid E.164 phone number",
	"max_bytes":                 "%s must not exceed the maximum file size",
	"mime_types":                "%s must be a valid file type",
	"mimes":                     "%s must be a valid file type",
	"mimetypes":                 "%s must be a valid MIME type",
	"extensions":                "%s must have an allowed file extension",
	"dimensions":                "%s must have valid dimensions",
	"arraycontains":             "%s must contain the required values",
	"boolean":                   "%s must be true or false",
	"date":                      "%s must be a valid date",
	"datetime":                  "%s must be a valid datetime",
	"timezone":                  "%s must be a valid timezone",
	"phone":                     "%s must be a valid phone number",
	"mobile":                    "%s must be a valid mobile number",
	"positive_int":              "%s must be a positive integer",
	"positive":                  "%s must be a positive number",
	"price":                     "%s must be a valid price",
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
	translated := Trans(c, messageID, data)
	if translated == messageID {
		return fallback, nil
	}
	return translated, nil
}

func fallbackValidationMessage(tag, fieldName, param string) string {
	format, ok := fallbackValidationMessages[tag]
	if !ok {
		return fmt.Sprintf("%s failed validation for '%s'", fieldName, tag)
	}

	placeholders := strings.Count(format, "%s")
	if placeholders == 0 {
		return format
	}

	args := make([]any, 0, placeholders)
	args = append(args, fieldName)
	if placeholders > 1 {
		for _, part := range splitValidationParam(param) {
			args = append(args, part)
		}
	}
	for len(args) < placeholders {
		args = append(args, "")
	}
	if len(args) > placeholders {
		args = args[:placeholders]
	}

	return fmt.Sprintf(format, args...)
}

func splitValidationParam(param string) []string {
	param = strings.TrimSpace(param)
	if param == "" {
		return nil
	}

	return strings.FieldsFunc(param, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
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
