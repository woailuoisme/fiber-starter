package requests

import (
	"mime/multipart"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// registerCustomValidations 注册自定义验证规则和 Laravel 别名
func registerCustomValidations(v *validator.Validate) {
	_ = v.RegisterValidation("array", validateArray)
	_ = v.RegisterValidation("list", validateList)
	_ = v.RegisterValidation("required_array_keys", validateRequiredArrayKeys)
	_ = v.RegisterValidation("alpha_dash", validateAlphaDash)
	_ = v.RegisterValidation("accepted", validateAccepted)
	_ = v.RegisterValidation("declined", validateDeclined)
	_ = v.RegisterValidation("date_format", validateDateFormat)
	_ = v.RegisterValidation("date_equals", validateDateEquals)
	_ = v.RegisterValidation("after", validateAfter)
	_ = v.RegisterValidation("after_or_equal", validateAfterOrEqual)
	_ = v.RegisterValidation("before", validateBefore)
	_ = v.RegisterValidation("before_or_equal", validateBeforeOrEqual)
	_ = v.RegisterValidation("uploaded_file", validateFile)
	_ = v.RegisterValidation("max_bytes", validateMaxBytes)
	_ = v.RegisterValidation("mime_types", validateMimeTypes)
	_ = v.RegisterValidation("phone", validatePhone)
	_ = v.RegisterValidation("positive_int", validatePositiveInt)
	_ = v.RegisterValidation("positive", validatePositive)
	_ = v.RegisterValidation("price", validatePrice)
	_ = v.RegisterValidation("mobile", validateMobile)
	_ = v.RegisterValidation("date", validateDate)

	// 注册 Laravel 风格的别名
	v.RegisterAlias("alpha_num", "alphanumunicode")
	v.RegisterAlias("starts_with", "startswith")
	v.RegisterAlias("ends_with", "endswith")
	v.RegisterAlias("doesnt_start_with", "startsnotwith")
	v.RegisterAlias("doesnt_end_with", "endsnotwith")
	v.RegisterAlias("mac_address", "mac")
}

// validatePhone 验证手机号
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	matched, _ := regexp.MatchString(`^\+?[1-9]\d{7,14}$`, phone)
	return matched
}

// validateMobile 验证中国手机号
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
	return matched
}

// validatePositiveInt 验证正整数
func validatePositiveInt(fl validator.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fl.Field().Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fl.Field().Uint() > 0
	default:
		return false
	}
}

// validatePositive 验证正数
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

// validatePrice 验证价格（正数）
func validatePrice(fl validator.FieldLevel) bool {
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

// validateArray 验证数组或映射，并可限制允许的键。
func validateArray(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	case reflect.Map:
		if field.Type().Key().Kind() != reflect.String {
			return false
		}

		allowed := parseAllowedKeys(fl.Param())
		if len(allowed) == 0 {
			return true
		}

		for _, key := range field.MapKeys() {
			if _, ok := allowed[key.String()]; !ok {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// validateList 验证数组是否为列表。
func validateList(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	case reflect.Map:
		if field.Type().Key().Kind() != reflect.String {
			return false
		}

		keys := field.MapKeys()
		if len(keys) == 0 {
			return true
		}

		indexes := make([]int, 0, len(keys))
		for _, key := range keys {
			index, err := strconv.Atoi(key.String())
			if err != nil {
				return false
			}
			indexes = append(indexes, index)
		}
		sort.Ints(indexes)
		for i, index := range indexes {
			if index != i {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// validateRequiredArrayKeys 验证数组包含指定键。
func validateRequiredArrayKeys(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.Map || field.Type().Key().Kind() != reflect.String {
		return false
	}

	required := parseAllowedKeys(fl.Param())
	if len(required) == 0 {
		return false
	}

	for key := range required {
		if !field.MapIndex(reflect.ValueOf(key)).IsValid() {
			return false
		}
	}

	return true
}

// validateAlphaDash 验证字母数字横杠下划线。
func validateAlphaDash(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^[\pL\pN_-]+$`, value)
	return matched
}

// validateAccepted 验证字段是否被接受。
func validateAccepted(fl validator.FieldLevel) bool {
	switch value := fl.Field().Interface().(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "yes", "on", "1":
			return true
		default:
			return false
		}
	case int, int8, int16, int32, int64:
		return fl.Field().Int() == 1
	case uint, uint8, uint16, uint32, uint64:
		return fl.Field().Uint() == 1
	default:
		return false
	}
}

// validateDeclined 验证字段是否被拒绝。
func validateDeclined(fl validator.FieldLevel) bool {
	switch value := fl.Field().Interface().(type) {
	case bool:
		return !value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "false", "no", "off", "0":
			return true
		default:
			return false
		}
	case int, int8, int16, int32, int64:
		return fl.Field().Int() == 0
	case uint, uint8, uint16, uint32, uint64:
		return fl.Field().Uint() == 0
	default:
		return false
	}
}

// validateDateFormat 验证日期格式。
func validateDateFormat(fl validator.FieldLevel) bool {
	layout := strings.TrimSpace(fl.Param())
	if layout == "" {
		return false
	}

	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return false
	}

	_, err := time.Parse(layout, value)
	return err == nil
}

// validateDateEquals 验证日期相等。
func validateDateEquals(fl validator.FieldLevel) bool {
	left, ok := parseDateValue(fl.Field())
	if !ok {
		return false
	}

	right, ok := parseDateParam(fl.Param())
	if !ok {
		return false
	}

	return left.Equal(right)
}

// validateAfter 验证日期在指定时间之后。
func validateAfter(fl validator.FieldLevel) bool {
	left, ok := parseDateValue(fl.Field())
	if !ok {
		return false
	}

	right, ok := parseDateParam(fl.Param())
	if !ok {
		return false
	}

	return left.After(right)
}

// validateAfterOrEqual 验证日期在指定时间之后或相等。
func validateAfterOrEqual(fl validator.FieldLevel) bool {
	left, ok := parseDateValue(fl.Field())
	if !ok {
		return false
	}

	right, ok := parseDateParam(fl.Param())
	if !ok {
		return false
	}

	return left.After(right) || left.Equal(right)
}

// validateBefore 验证日期在指定时间之前。
func validateBefore(fl validator.FieldLevel) bool {
	left, ok := parseDateValue(fl.Field())
	if !ok {
		return false
	}

	right, ok := parseDateParam(fl.Param())
	if !ok {
		return false
	}

	return left.Before(right)
}

// validateBeforeOrEqual 验证日期在指定时间之前或相等。
func validateBeforeOrEqual(fl validator.FieldLevel) bool {
	left, ok := parseDateValue(fl.Field())
	if !ok {
		return false
	}

	right, ok := parseDateParam(fl.Param())
	if !ok {
		return false
	}

	return left.Before(right) || left.Equal(right)
}

// validateFile 验证上传文件对象。
func validateFile(fl validator.FieldLevel) bool {
	switch v := fl.Field().Interface().(type) {
	case *multipart.FileHeader:
		return v != nil && v.Filename != ""
	case multipart.FileHeader:
		return v.Filename != ""
	default:
		return false
	}
}

// validateMaxBytes 验证文件大小。
func validateMaxBytes(fl validator.FieldLevel) bool {
	maxBytes, err := strconv.ParseInt(strings.TrimSpace(fl.Param()), 10, 64)
	if err != nil || maxBytes <= 0 {
		return false
	}

	switch v := fl.Field().Interface().(type) {
	case *multipart.FileHeader:
		return v != nil && v.Size <= maxBytes
	case multipart.FileHeader:
		return v.Size <= maxBytes
	default:
		return false
	}
}

// validateMimeTypes 验证文件 Content-Type。
func validateMimeTypes(fl validator.FieldLevel) bool {
	allowed := parseAllowedKeys(fl.Param())
	if len(allowed) == 0 {
		return true
	}

	var contentType string
	switch v := fl.Field().Interface().(type) {
	case *multipart.FileHeader:
		if v == nil {
			return false
		}
		contentType = v.Header.Get("Content-Type")
	case multipart.FileHeader:
		contentType = v.Header.Get("Content-Type")
	default:
		return false
	}

	if contentType == "" {
		return false
	}

	if _, ok := allowed[contentType]; ok {
		return true
	}

	for allowedType := range allowed {
		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(contentType, prefix+"/") {
				return true
			}
		}
	}

	return false
}

// validateDate 验证日期。
func validateDate(fl validator.FieldLevel) bool {
	field := fl.Field()
	if parsed, ok := parseDateValue(field); ok {
		return !parsed.IsZero()
	}

	return false
}

func parseDateValue(field reflect.Value) (time.Time, bool) {
	switch field.Kind() {
	case reflect.String:
		value := strings.TrimSpace(field.String())
		if value == "" {
			return time.Time{}, false
		}

		layouts := []string{
			time.DateOnly,
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006/01/02",
		}
		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, value); err == nil {
				return parsed, true
			}
		}
		return time.Time{}, false
	case reflect.Struct:
		if t, ok := field.Interface().(time.Time); ok && !t.IsZero() {
			return t, true
		}
		return time.Time{}, false
	case reflect.Pointer:
		if field.IsNil() {
			return time.Time{}, false
		}
		return parseDateValue(field.Elem())
	default:
		return time.Time{}, false
	}
}

func parseDateParam(param string) (time.Time, bool) {
	param = strings.TrimSpace(param)
	if param == "" {
		return time.Time{}, false
	}

	if idx := strings.Index(param, ","); idx >= 0 {
		param = strings.TrimSpace(param[:idx])
	}

	layouts := []string{
		time.DateOnly,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, param); err == nil {
			return parsed, true
		}
	}

	if strings.EqualFold(param, "now") {
		return time.Now(), true
	}

	return time.Time{}, false
}

func parseAllowedKeys(param string) map[string]struct{} {
	param = strings.TrimSpace(param)
	if param == "" {
		return nil
	}

	keys := make(map[string]struct{})
	for _, key := range strings.FieldsFunc(param, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	}) {
		if key != "" {
			keys[key] = struct{}{}
		}
	}

	return keys
}
