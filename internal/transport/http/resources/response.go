// Package resources 定义HTTP响应结构和格式化工具
package resources

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// APIResponse 统一API响应结构
// Requirements: 14.1, 14.2, 14.3
type APIResponse struct {
	Success  bool        `json:"success"`
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors,omitempty"`
	Debugger *Debugger   `json:"debugger,omitempty"`
}

// Debugger 调试信息结构体
// Requirements: 13.1
type Debugger struct {
	Exception   string   `json:"exception"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Trace       []string `json:"trace"`
	RequestTime float64  `json:"request_time"` // milliseconds
	MemoryUsage string   `json:"memory_usage"`
	QueryCount  int      `json:"query_count"`
}

// PaginationMeta 分页元数据
// Requirements: 14.6
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	LastPage    int   `json:"last_page"`
	HasMore     bool  `json:"has_more"`
	Total       int64 `json:"total"`
	From        int   `json:"from"`
	To          int   `json:"to"`
}

// PaginatedResponse 分页响应结构
// Requirements: 14.5, 14.6
type PaginatedResponse struct {
	Items []interface{}  `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

// JSON 发送成功响应
func JSON(c fiber.Ctx, data interface{}, message string) error {
	if message == "" {
		message = "success"
	}
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Created 发送创建成功响应
func Created(c fiber.Ctx, data interface{}, message string) error {
	if message == "" {
		message = "created"
	}
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Code:    fiber.StatusCreated,
		Message: message,
		Data:    data,
	})
}

// Error 发送错误响应
func Error(c fiber.Ctx, status int, message string, errors interface{}) error {
	// 如果是内部服务器错误，且不在调试模式下，隐藏详细错误信息
	if status == fiber.StatusInternalServerError && !IsDebugMode() {
		message = "Internal Server Error"
		errors = nil
	}

	return c.Status(status).JSON(APIResponse{
		Success: false,
		Code:    status,
		Message: message,
		Errors:  errors,
	})
}

// ErrorWithDebugger 带调试信息的错误响应
// Requirements: 12.5, 12.6, 12.7, 12.8, 12.9
func ErrorWithDebugger(
	c fiber.Ctx,
	code int,
	message string,
	errors interface{},
	exception string,
	file string,
	line int,
) error {
	response := APIResponse{
		Success: false,
		Code:    code,
		Message: message,
		Errors:  errors,
	}

	// 如果是调试模式，添加详细调试信息
	if IsDebugMode() {
		debugger := GetDebugger(c)
		debugger.Exception = exception
		debugger.File = file
		debugger.Line = line
		debugger.Trace = GetStackTrace()
		response.Debugger = debugger
	}

	return c.Status(code).JSON(response)
}

// Paginated 分页响应
// Requirements: 14.5, 14.6
func Paginated(c fiber.Ctx, items []interface{}, currentPage, perPage int, total int64) error {
	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage < 1 {
		lastPage = 1
	}

	from := (currentPage-1)*perPage + 1
	to := currentPage * perPage
	if to > int(total) {
		to = int(total)
	}
	if from > int(total) {
		from = int(total)
	}

	meta := PaginationMeta{
		CurrentPage: currentPage,
		PerPage:     perPage,
		LastPage:    lastPage,
		HasMore:     currentPage < lastPage,
		Total:       total,
		From:        from,
		To:          to,
	}

	data := PaginatedResponse{
		Items: items,
		Meta:  meta,
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Message: "Success",
		Data:    data,
	})
}

// GetDebugger 获取调试信息
// Requirements: 13.1, 12.6, 12.7, 12.8
func GetDebugger(c fiber.Ctx) *Debugger {
	startTime := c.Locals("start_time")
	var requestTime float64
	if startTime != nil {
		if t, ok := startTime.(time.Time); ok {
			requestTime = float64(time.Since(t).Microseconds()) / 1000.0 // 转换为毫秒
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024)

	queryCount := 0
	if qc := c.Locals("query_count"); qc != nil {
		if count, ok := qc.(int); ok {
			queryCount = count
		}
	}

	return &Debugger{
		RequestTime: requestTime,
		MemoryUsage: memoryUsage,
		QueryCount:  queryCount,
	}
}

// GetStackTrace 获取堆栈跟踪
// Requirements: 13.3
func GetStackTrace() []string {
	var traces []string
	for i := 2; i < 10; i++ { // 跳过前两个调用（GetStackTrace 和调用者）
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			traces = append(traces, fmt.Sprintf("%s() at %s:%d", fn.Name(), file, line))
		}
	}
	return traces
}

// IsDebugMode 检查是否为调试模式
// Requirements: 12.5, 12.9, 13.9
func IsDebugMode() bool {
	// 从环境变量或配置中读取
	// 这里简化处理，实际应该从配置中读取
	return true // 临时返回 true，后续会从配置中读取
}

// FormatValidationErrors 格式化验证错误
// Requirements: 14.7
func FormatValidationErrors(err error) map[string][]string {
	errMap := make(map[string][]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			tag := e.Tag()

			var message string
			switch tag {
			case "required":
				message = fmt.Sprintf("The %s field is required.", field)
			case "min":
				message = fmt.Sprintf("The %s must be at least %s characters.", field, e.Param())
			case "max":
				message = fmt.Sprintf("The %s must not exceed %s characters.", field, e.Param())
			case "email":
				message = fmt.Sprintf("The %s must be a valid email address.", field)
			case "e164":
				message = fmt.Sprintf("The %s must be a valid phone number.", field)
			case "url":
				message = fmt.Sprintf("The %s must be a valid URL.", field)
			case "gt":
				message = fmt.Sprintf("The %s must be greater than %s.", field, e.Param())
			case "gte":
				message = fmt.Sprintf("The %s must be greater than or equal to %s.", field, e.Param())
			case "lt":
				message = fmt.Sprintf("The %s must be less than %s.", field, e.Param())
			case "lte":
				message = fmt.Sprintf("The %s must be less than or equal to %s.", field, e.Param())
			default:
				message = fmt.Sprintf("The %s field is invalid.", field)
			}

			errMap[field] = append(errMap[field], message)
		}
	}

	return errMap
}

// SanitizeString 清理字符串
func SanitizeString(s string) string {
	// 去除首尾空格
	s = strings.TrimSpace(s)
	// 可以添加更多的清理逻辑
	return s
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	// 简单的邮箱验证
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// GenerateSlug 生成URL友好的slug
func GenerateSlug(text string) string {
	// 转换为小写
	slug := strings.ToLower(text)
	// 替换空格为连字符
	slug = strings.ReplaceAll(slug, " ", "-")
	// 移除特殊字符
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	return slug
}

// SuccessResponse 创建成功响应对象（不直接返回给客户端）
// Requirements: 14.2, 14.4
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse 创建错误响应对象（不直接返回给客户端）
// Requirements: 14.3, 14.7
func ErrorResponse(message string, errors interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Code:    fiber.StatusBadRequest,
		Message: message,
		Errors:  errors,
	}
}
