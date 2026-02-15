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

	"fiber-starter/app/exceptions"
	"fiber-starter/app/helpers"
	"fiber-starter/app/http/resources"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorHandler 全局错误处理中间件
// Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.10, 12.11
func ErrorHandler(c fiber.Ctx) error {
	// 执行下一个处理器
	err := c.Next()

	// 如果没有错误，直接返回
	if err == nil {
		return nil
	}

	// 记录错误日志
	// Requirements: 12.4, 12.11
	logError(c, err)

	// 处理 APIException
	// Requirements: 12.1, 11.13
	apiErr := &exceptions.APIException{}
	if errors.As(err, &apiErr) {
		return handleAPIException(c, apiErr)
	}

	// 处理 ValidationException
	valErr := &exceptions.ValidationException{}
	if errors.As(err, &valErr) {
		return handleAPIException(c, valErr.APIException)
	}

	// 处理 AuthenticationException
	authErr := &exceptions.AuthenticationException{}
	if errors.As(err, &authErr) {
		return handleAPIException(c, authErr.APIException)
	}

	// 处理 AuthorizationException
	authzErr := &exceptions.AuthorizationException{}
	if errors.As(err, &authzErr) {
		return handleAPIException(c, authzErr.APIException)
	}

	// 处理 NotFoundException
	notFoundErr := &exceptions.NotFoundException{}
	if errors.As(err, &notFoundErr) {
		return handleAPIException(c, notFoundErr.APIException)
	}

	// 处理 BadRequestException
	badReqErr := &exceptions.BadRequestException{}
	if errors.As(err, &badReqErr) {
		return handleAPIException(c, badReqErr.APIException)
	}

	// 处理 ConflictException
	conflictErr := &exceptions.ConflictException{}
	if errors.As(err, &conflictErr) {
		return handleAPIException(c, conflictErr.APIException)
	}

	// 处理 ServerException
	serverErr := &exceptions.ServerException{}
	if errors.As(err, &serverErr) {
		return handleAPIException(c, serverErr.APIException)
	}

	// 处理验证错误
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return handleValidationError(c, validationErrors)
	}

	// 处理 Fiber 框架错误
	fiberErr := &fiber.Error{}
	if errors.As(err, &fiberErr) {
		return handleFiberError(c, fiberErr)
	}

	// 处理未捕获的异常
	// Requirements: 12.2, 12.3
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
	// 构建日志信息
	logMsg := fmt.Sprintf(
		"[%s] %s %s - %s",
		c.Method(),
		c.Path(),
		c.IP(),
		err.Error(),
	)

	// 添加请求ID（如果存在）
	if requestID := c.Get("X-Request-ID"); requestID != "" {
		logMsg = fmt.Sprintf("%s [RequestID: %s]", logMsg, requestID)
	}

	// 根据错误类型选择日志级别
	apiErr := &exceptions.APIException{}
	if errors.As(err, &apiErr) {
		if apiErr.Code >= 500 {
			helpers.Logger.Error(logMsg, zap.Int("code", apiErr.Code))
		} else {
			helpers.Logger.Warn(logMsg, zap.Int("code", apiErr.Code))
		}
	} else {
		// 未知错误，记录堆栈信息
		if isDevelopment() {
			helpers.LogError(logMsg, zap.String("stack", string(debug.Stack())))
		} else {
			helpers.LogError(logMsg)
		}
	}
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

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 获取现有的请求ID
		requestID := c.Get("X-Request-ID")

		// 如果没有请求ID，生成一个新的
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置请求ID到响应头
		c.Set("X-Request-ID", requestID)

		// 将请求ID存储到本地存储中
		c.Locals("requestID", requestID)

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

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 这里可以使用UUID或其他方式生成唯一ID
	// 为了简单起见，使用时间戳和随机数
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
