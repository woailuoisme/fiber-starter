package middleware

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"fiber-starter/app/Exceptions"
	"fiber-starter/app/Http/Resources"
	helpers "fiber-starter/app/Support"

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

	if apiErr := unwrapAPIException(err); apiErr != nil {
		return handleAPIException(c, apiErr)
	}

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return handleValidationError(c, validationErrors)
	}

	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		return handleFiberError(c, fiberErr)
	}

	return handleUnknownError(c, err)
}

func unwrapAPIException(err error) *exceptions.APIException {
	if apiErr, ok := errors.AsType[*exceptions.APIException](err); ok {
		return apiErr
	}
	if valErr, ok := errors.AsType[*exceptions.ValidationException](err); ok {
		return valErr.APIException
	}
	if authErr, ok := errors.AsType[*exceptions.AuthenticationException](err); ok {
		return authErr.APIException
	}
	if authzErr, ok := errors.AsType[*exceptions.AuthorizationException](err); ok {
		return authzErr.APIException
	}
	if notFoundErr, ok := errors.AsType[*exceptions.NotFoundException](err); ok {
		return notFoundErr.APIException
	}
	if badReqErr, ok := errors.AsType[*exceptions.BadRequestException](err); ok {
		return badReqErr.APIException
	}
	if conflictErr, ok := errors.AsType[*exceptions.ConflictException](err); ok {
		return conflictErr.APIException
	}
	if serverErr, ok := errors.AsType[*exceptions.ServerException](err); ok {
		return serverErr.APIException
	}

	return nil
}

func handleAPIException(c fiber.Ctx, apiErr *exceptions.APIException) error {
	return writeDebuggerError(c, apiErr.Code, apiErr.Message, apiErr.Errors, "APIException", 2)
}

func writeDebuggerError(c fiber.Ctx, code int, message string, details interface{}, exception string, callerSkip int) error {
	_, file, line, _ := runtime.Caller(callerSkip)
	return resources.ErrorWithDebugger(c, code, message, details, exception, file, line)
}

func handleFiberError(c fiber.Ctx, fiberErr *fiber.Error) error {
	return writeDebuggerError(c, fiberErr.Code, fiberErrorMessage(fiberErr), nil, "FiberError", 1)
}

func handleUnknownError(c fiber.Ctx, err error) error {
	message := "Internal server error"
	if isDevelopment() {
		message = fmt.Sprintf("Internal server error: %s", err.Error())
	}

	return writeDebuggerError(c, 500, message, nil, fmt.Sprintf("%T", err), 1)
}

func handleValidationError(c fiber.Ctx, validationErrors validator.ValidationErrors) error {
	return writeDebuggerError(c, 422, "Validation failed", resources.FormatValidationErrors(validationErrors), "ValidationError", 1)
}

func fiberErrorMessage(fiberErr *fiber.Error) string {
	switch fiberErr.Code {
	case fiber.StatusBadRequest:
		return "Bad request"
	case fiber.StatusUnauthorized:
		return "Unauthorized"
	case fiber.StatusForbidden:
		return "Forbidden"
	case fiber.StatusNotFound:
		return "Not found"
	case fiber.StatusMethodNotAllowed:
		return "Method not allowed"
	case fiber.StatusRequestTimeout:
		return "Request timeout"
	case fiber.StatusTooManyRequests:
		return "Too many requests"
	case fiber.StatusInternalServerError:
		return "Internal server error"
	case fiber.StatusBadGateway:
		return "Bad gateway"
	case fiber.StatusServiceUnavailable:
		return "Service unavailable"
	default:
		return fiberErr.Message
	}
}

func logError(c fiber.Ctx, err error) {
	requestID := getRequestID(c)
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("error", err.Error()),
	}

	if apiErr := unwrapAPIException(err); apiErr != nil {
		logHTTPError(fields, apiErr.Code)
		return
	}

	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		logHTTPError(fields, fiberErr.Code)
		return
	}

	if isDevelopment() {
		fields = append(fields, zap.String("stack", string(debug.Stack())))
	}

	helpers.Logger.Error("http_error", fields...)
}

func logHTTPError(fields []zap.Field, code int) {
	fields = append(fields, zap.Int("code", code))
	if code >= 500 {
		helpers.Logger.Error("http_error", fields...)
		return
	}

	helpers.Logger.Warn("http_error", fields...)
}

func isDevelopment() bool {
	env := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "development")))
	return env == "development" || env == "dev" || env == "local"
}

func getEnv(key, defaultValue string) string {
	if value := getValue(key); value != "" {
		return value
	}
	return defaultValue
}

func getValue(key string) string {
	return os.Getenv(key)
}

// RecoveryMiddleware 恢复中间件，用于捕获panic
// Requirements: 12.2, 12.3
func RecoveryMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				helpers.LogError("PANIC: "+fmt.Sprint(r), zap.String("stack", string(debug.Stack())))
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
		c.Locals("start_time", time.Now())
		return c.Next()
	}
}
