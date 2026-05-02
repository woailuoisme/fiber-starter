package support

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	exceptions "fiber-starter/app/Exceptions"
	supporti18n "fiber-starter/app/Support/i18n"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.uber.org/zap"
)

// HandleHTTPError 统一处理 HTTP 错误响应与日志。
// 作用：把已知异常映射成统一响应，并输出结构化日志。
// 场景：全局 ErrorHandler、timeout 兜底、panic recovery 等 HTTP 错误入口。
// 使用方式：只在 HTTP 错误入口调用，不建议业务代码直接依赖它。
func HandleHTTPError(c fiber.Ctx, err error) error {
	logHTTPError(c, err)

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
	return ErrorWithDebugger(c, code, message, details, exception, file, line)
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
	return writeDebuggerError(c, 422, "Validation failed", supporti18n.FormatValidationErrorsWithContext(c, validationErrors), "ValidationError", 1)
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

func logHTTPError(c fiber.Ctx, err error) {
	requestID := getRequestID(c)
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("error", err.Error()),
	}

	if apiErr := unwrapAPIException(err); apiErr != nil {
		logHTTPStatus(fields, apiErr.Code)
		return
	}

	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		logHTTPStatus(fields, fiberErr.Code)
		return
	}

	if isDevelopment() {
		fields = append(fields, zap.String("stack", string(debug.Stack())))
	}

	Logger.Error("http_error", fields...)
}

func logHTTPStatus(fields []zap.Field, code int) {
	fields = append(fields, zap.Int("code", code))
	if code >= 500 {
		Logger.Error("http_error", fields...)
		return
	}

	Logger.Warn("http_error", fields...)
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

func getRequestID(c fiber.Ctx) string {
	if v := requestid.FromContext(c); v != "" {
		return v
	}

	if v := c.Get(fiber.HeaderXRequestID); v != "" {
		return v
	}

	if v := c.Get("X-Request-ID"); v != "" {
		return v
	}

	if v := c.Locals(fiber.HeaderXRequestID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}
