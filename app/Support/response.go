package support

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	exceptions "fiber-starter/app/Exceptions"
	models "fiber-starter/app/Models"
	supporti18n "fiber-starter/app/Support/i18n"

	"github.com/gofiber/fiber/v3"
)

// APIResponse 统一 API 响应结构。
type APIResponse struct {
	Success  bool        `json:"success"`
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors,omitempty"`
	Debugger *Debugger   `json:"debugger,omitempty"`
}

// Debugger 调试信息结构体。
type Debugger struct {
	Exception   string   `json:"exception"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Trace       []string `json:"trace"`
	RequestTime float64  `json:"request_time"` // milliseconds
	MemoryUsage string   `json:"memory_usage"`
	QueryCount  int      `json:"query_count"`
}

// PaginationMeta 分页元数据。
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	LastPage    int   `json:"last_page"`
	HasMore     bool  `json:"has_more"`
	Total       int64 `json:"total"`
	From        int   `json:"from"`
	To          int   `json:"to"`
}

// PaginatedResponse 分页响应结构。
type PaginatedResponse struct {
	Items []interface{}  `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

func writeJSONResponse(ctx fiber.Ctx, status int, success bool, message string, data any, errs any) error {
	response := APIResponse{
		Success: success,
		Code:    status,
		Message: message,
	}

	if success {
		response.Data = data
	} else {
		response.Errors = errs
	}

	return ctx.Status(status).JSON(response)
}

func paginationMeta(total int64, page, limit int) PaginationMeta {
	lastPage := int((total + int64(limit) - 1) / int64(limit))
	if lastPage < 1 {
		lastPage = 1
	}

	from := (page-1)*limit + 1
	to := page * limit
	if total == 0 {
		from = 0
		to = 0
	} else if to > int(total) {
		to = int(total)
	}

	return PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		LastPage:    lastPage,
		HasMore:     page < lastPage,
		Total:       total,
		From:        from,
		To:          to,
	}
}

// JSON 发送成功响应。
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

// Created 发送创建成功响应。
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

// ErrorWithDebugger 带调试信息的错误响应。
func ErrorWithDebugger(
	c fiber.Ctx,
	code int,
	message string,
	errs interface{},
	exception string,
	file string,
	line int,
) error {
	response := APIResponse{
		Success: false,
		Code:    code,
		Message: message,
		Errors:  errs,
	}

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

// Paginated 分页响应。
func Paginated(c fiber.Ctx, items []interface{}, currentPage, perPage int, total int64) error {
	data := PaginatedResponse{
		Items: items,
		Meta:  paginationMeta(total, currentPage, perPage),
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Code:    fiber.StatusOK,
		Message: "Success",
		Data:    data,
	})
}

// HandleAppError 处理业务层显式返回的应用错误。
func HandleAppError(ctx fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := exceptions.GetAPIException(err); ok {
		return writeJSONResponse(ctx, apiErr.Code, false, apiErr.Message, nil, apiErr.Errors)
	}

	if appErr, ok := exceptions.GetAppError(err); ok {
		return writeJSONResponse(ctx, appErr.StatusCode, false, appErr.Message, nil, nil)
	}

	return writeJSONResponse(ctx, fiber.StatusInternalServerError, false, "Internal server error", nil, nil)
}

// HandleValidationError 处理参数校验失败。
func HandleValidationError(ctx fiber.Ctx, err error) error {
	if apiErr, ok := exceptions.GetAPIException(err); ok {
		return writeJSONResponse(ctx, apiErr.Code, false, apiErr.Message, nil, apiErr.Errors)
	}

	validationErrors := supporti18n.FormatValidationErrorsWithContext(ctx, err)
	if len(validationErrors) > 0 {
		return writeJSONResponse(ctx, fiber.StatusUnprocessableEntity, false, "Validation failed", nil, validationErrors)
	}

	return writeJSONResponse(ctx, fiber.StatusUnprocessableEntity, false, "Validation failed", nil, nil)
}

// HandleSuccess 返回成功响应。
func HandleSuccess(ctx fiber.Ctx, message string, data interface{}) error {
	if message == "" {
		message = "success"
	}
	return writeJSONResponse(ctx, fiber.StatusOK, true, message, data, nil)
}

// HandleCreated 返回创建成功响应。
func HandleCreated(ctx fiber.Ctx, message string, data interface{}) error {
	if message == "" {
		message = "created"
	}
	return writeJSONResponse(ctx, fiber.StatusCreated, true, message, data, nil)
}

// HandleError 返回指定状态码的错误响应。
func HandleError(ctx fiber.Ctx, status int, message string, errs any) error {
	return writeJSONResponse(ctx, status, false, message, nil, errs)
}

// HandleNotFound 返回资源不存在响应。
func HandleNotFound(ctx fiber.Ctx, message string) error {
	return HandleError(ctx, fiber.StatusNotFound, message, nil)
}

// HandleBadRequest 返回错误请求响应。
func HandleBadRequest(ctx fiber.Ctx, message string) error {
	return HandleError(ctx, fiber.StatusBadRequest, message, nil)
}

// HandleUnauthorized 返回未授权响应。
func HandleUnauthorized(ctx fiber.Ctx, message string) error {
	return HandleError(ctx, fiber.StatusUnauthorized, message, nil)
}

// HandleForbidden 返回禁止访问响应。
func HandleForbidden(ctx fiber.Ctx, message string) error {
	return HandleError(ctx, fiber.StatusForbidden, message, nil)
}

// HandleConflict 返回冲突响应。
func HandleConflict(ctx fiber.Ctx, message string) error {
	return HandleError(ctx, fiber.StatusConflict, message, nil)
}

// HandleInternalServerError 返回服务器错误响应。
func HandleInternalServerError(ctx fiber.Ctx, message string) error {
	return HandleError(ctx, fiber.StatusInternalServerError, message, nil)
}

// FormatValidationErrorsToString 将校验错误格式化为字符串。
func FormatValidationErrorsToString(err error) string {
	validationErrors := supporti18n.FormatValidationErrors(err)
	if len(validationErrors) == 0 {
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

// HandleUserResponse 返回用户对象响应。
func HandleUserResponse(ctx fiber.Ctx, message string, user *models.User) error {
	return HandleSuccess(ctx, message, user.ToSafeUser())
}

// HandleUserListResponse 返回用户列表响应。
func HandleUserListResponse(ctx fiber.Ctx, message string, users []models.User, total int64, page, limit int) error {
	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
	}

	return HandlePaginationResponse(ctx, message, userResponses, total, page, limit)
}

// HandlePaginationResponse 返回分页响应。
func HandlePaginationResponse(ctx fiber.Ctx, message string, data interface{}, total int64, page, limit int) error {
	if message == "" {
		message = "success"
	}

	return HandleSuccess(ctx, message, fiber.Map{
		"items": data,
		"meta":  paginationMeta(total, page, limit),
	})
}

// ValidatePaginationParams 校验分页参数。
func ValidatePaginationParams(pageStr, limitStr string) (int, int, error) {
	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := parseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := parseInt(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return page, limit, nil
}

// GetDebugger 获取调试信息。
func GetDebugger(c fiber.Ctx) *Debugger {
	startTime := c.Locals("start_time")
	var requestTime float64
	if startTime != nil {
		if t, ok := startTime.(time.Time); ok {
			requestTime = float64(time.Since(t).Microseconds()) / 1000.0
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

// GetStackTrace 获取堆栈跟踪。
func GetStackTrace() []string {
	var traces []string
	for i := 2; i < 10; i++ {
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

// IsDebugMode 检查是否为调试模式。
func IsDebugMode() bool {
	return true
}

// FormatValidationErrors 格式化验证错误。
func FormatValidationErrors(err error) map[string][]string {
	return supporti18n.FormatValidationErrors(err)
}

// SanitizeString 清理字符串。
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// IsValidEmail 验证邮箱格式。
func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// GenerateSlug 生成 URL 友好的 slug。
func GenerateSlug(text string) string {
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	return slug
}

func parseInt(s string) (int, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, exceptions.BadRequest("Invalid number format")
	}
	return value, nil
}
