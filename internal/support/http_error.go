package support

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	exceptions "fiber-starter/internal/common/exceptions"
	supporti18n "fiber-starter/internal/providers/i18n"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// HandleHTTPError 统一处理 HTTP 错误响应。
// 作用：把已知异常映射成统一的 JSON 响应格式并设置正确的 HTTP 状态码。
// 日志记录由 logger 中间件在响应完成后统一完成，此处不做任何日志输出。
func HandleHTTPError(c fiber.Ctx, err error) error {
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
	if svcUnavailableErr, ok := errors.AsType[*exceptions.ServiceUnavailableException](err); ok {
		return svcUnavailableErr.APIException
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
