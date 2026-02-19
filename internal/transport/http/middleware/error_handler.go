// Package middleware 提供应用程序的中间件
package middleware

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"fiber-starter/internal/platform/exceptions"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/transport/http/resources"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorHandler 全局错误处理中间件
// Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.10, 12.11
func ErrorHandler(c fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}
	return HandleError(c, err)
}

// HandleError 统一处理错误响应与日志
func HandleError(c fiber.Ctx, err error) error {
	logError(c, err)

	apiErr := &exceptions.APIException{}
	if errors.As(err, &apiErr) {
		return handleAPIException(c, apiErr)
	}

	valErr := &exceptions.ValidationException{}
	if errors.As(err, &valErr) {
		return handleAPIException(c, valErr.APIException)
	}

	authErr := &exceptions.AuthenticationException{}
	if errors.As(err, &authErr) {
		return handleAPIException(c, authErr.APIException)
	}

	authzErr := &exceptions.AuthorizationException{}
	if errors.As(err, &authzErr) {
		return handleAPIException(c, authzErr.APIException)
	}

	notFoundErr := &exceptions.NotFoundException{}
	if errors.As(err, &notFoundErr) {
		return handleAPIException(c, notFoundErr.APIException)
	}

	badReqErr := &exceptions.BadRequestException{}
	if errors.As(err, &badReqErr) {
		return handleAPIException(c, badReqErr.APIException)
	}

	conflictErr := &exceptions.ConflictException{}
	if errors.As(err, &conflictErr) {
		return handleAPIException(c, conflictErr.APIException)
	}

	serverErr := &exceptions.ServerException{}
	if errors.As(err, &serverErr) {
		return handleAPIException(c, serverErr.APIException)
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return handleValidationError(c, validationErrors)
	}

	fiberErr := &fiber.Error{}
	if errors.As(err, &fiberErr) {
		return handleFiberError(c, fiberErr)
	}

	return handleUnknownError(c, err)
}

// handleAPIException 处理 API 异常
// Requirements: 12.1, 11.13
func handleAPIException(c fiber.Ctx, apiErr *exceptions.APIException) error {
	// 获取调用栈信息
	_, file, line, _ := runtime.Caller(2)

	// 使用带调试信息的错误响应
	return resources.ErrorWithDebugger(
		c,
		apiErr.Code,
		apiErr.Message,
		apiErr.Errors,
		"APIException",
		file,
		line,
	)
}

// handleValidationError 处理验证错误
func handleValidationError(c fiber.Ctx, validationErrors validator.ValidationErrors) error {
	errors := resources.FormatValidationErrors(validationErrors)
	_, file, line, _ := runtime.Caller(1)

	return resources.ErrorWithDebugger(
		c,
		422,
		"Validation failed",
		errors,
		"ValidationError",
		file,
		line,
	)
}

// handleFiberError 处理 Fiber 框架错误
func handleFiberError(c fiber.Ctx, fiberErr *fiber.Error) error {
	message := fiberErr.Message

	// 根据状态码提供更友好的错误信息
	switch fiberErr.Code {
	case fiber.StatusBadRequest:
		message = "Bad request"
	case fiber.StatusUnauthorized:
		message = "Unauthorized"
	case fiber.StatusForbidden:
		message = "Forbidden"
	case fiber.StatusNotFound:
		message = "Not found"
	case fiber.StatusMethodNotAllowed:
		message = "Method not allowed"
	case fiber.StatusRequestTimeout:
		message = "Request timeout"
	case fiber.StatusTooManyRequests:
		message = "Too many requests"
	case fiber.StatusInternalServerError:
		message = "Internal server error"
	case fiber.StatusBadGateway:
		message = "Bad gateway"
	case fiber.StatusServiceUnavailable:
		message = "Service unavailable"
	}

	_, file, line, _ := runtime.Caller(1)

	return resources.ErrorWithDebugger(
		c,
		fiberErr.Code,
		message,
		nil,
		"FiberError",
		file,
		line,
	)
}

// handleUnknownError 处理未知错误
// Requirements: 12.2, 12.3
func handleUnknownError(c fiber.Ctx, err error) error {
	message := "Internal server error"

	// 在开发环境中，可以返回更详细的错误信息
	if isDevelopment() {
		message = fmt.Sprintf("Internal server error: %s", err.Error())
	}

	_, file, line, _ := runtime.Caller(1)

	return resources.ErrorWithDebugger(
		c,
		500,
		message,
		nil,
		fmt.Sprintf("%T", err),
		file,
		line,
	)
}

// logError 记录错误日志
// Requirements: 12.4, 12.11
func logError(c fiber.Ctx, err error) {
	requestID := getRequestID(c)
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("error", err.Error()),
	}

	// 根据错误类型选择日志级别
	apiErr := &exceptions.APIException{}
	if errors.As(err, &apiErr) {
		fields = append(fields, zap.Int("code", apiErr.Code))
		if apiErr.Code >= 500 {
			helpers.Logger.Error("http_error", fields...)
		} else {
			helpers.Logger.Warn("http_error", fields...)
		}
		return
	}

	fiberErr := &fiber.Error{}
	if errors.As(err, &fiberErr) {
		fields = append(fields, zap.Int("code", fiberErr.Code))
		if fiberErr.Code >= 500 {
			helpers.Logger.Error("http_error", fields...)
		} else {
			helpers.Logger.Warn("http_error", fields...)
		}
		return
	}

	if isDevelopment() {
		fields = append(fields, zap.String("stack", string(debug.Stack())))
	}

	helpers.Logger.Error("http_error", fields...)
}

// isDevelopment 检查是否为开发环境
func isDevelopment() bool {
	env := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "development")))
	return env == "development" || env == "dev" || env == "local"
}

// getEnv 获取环境变量，提供默认值
func getEnv(key, defaultValue string) string {
	if value := getValue(key); value != "" {
		return value
	}
	return defaultValue
}

// getValue 获取环境变量值
func getValue(key string) string {
	return os.Getenv(key)
}

// RecoveryMiddleware 恢复中间件，用于捕获panic
// Requirements: 12.2, 12.3
func RecoveryMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic信息
				helpers.LogError("PANIC: "+fmt.Sprint(r), zap.String("stack", string(debug.Stack())))

				// 返回内部服务器错误
				_ = resources.ErrorWithDebugger(c, 500, "Internal server error", nil, "Panic", "", 0)
			}
		}()

		return c.Next()
	}
}

// RequestTimerMiddleware 请求计时中间件
// Requirements: 12.6, 12.8
func RequestTimerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 记录请求开始时间
		c.Locals("start_time", time.Now())

		return c.Next()
	}
}
