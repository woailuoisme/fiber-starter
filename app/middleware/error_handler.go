package middleware

import (
	"fiber-starter/app/errors"
	"fiber-starter/app/helpers"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler 全局错误处理中间件
func ErrorHandler(c *fiber.Ctx) error {
	// 执行下一个处理器
	err := c.Next()

	// 如果没有错误，直接返回
	if err == nil {
		return nil
	}

	// 处理不同类型的错误
	var appErr *errors.AppError
	var statusCode int
	var response interface{}

	// 检查是否为应用程序错误
	if errors.IsAppError(err) {
		appErr, _ = errors.GetAppError(err)
		statusCode = appErr.StatusCode
		response = handleAppError(appErr)
	} else if validationErrors, ok := err.(validator.ValidationErrors); ok {
		// 处理验证错误
		statusCode = fiber.StatusBadRequest
		response = handleValidationError(validationErrors)
	} else if fiberErr, ok := err.(*fiber.Error); ok {
		// 处理Fiber框架错误
		statusCode = fiberErr.Code
		response = handleFiberError(fiberErr)
	} else {
		// 处理未知错误
		statusCode = fiber.StatusInternalServerError
		response = handleUnknownError(err)
	}

	// 记录错误日志
	logError(c, err, appErr)

	// 返回错误响应
	return c.Status(statusCode).JSON(response)
}

// handleAppError 处理应用程序错误
func handleAppError(appErr *errors.AppError) interface{} {
	return helpers.ErrorResponse(appErr.Message, fiber.Map{
		"code":    appErr.Code,
		"details": appErr.Details,
	})
}

// handleValidationError 处理验证错误
func handleValidationError(validationErrors validator.ValidationErrors) interface{} {
	return helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(validationErrors))
}

// handleFiberError 处理Fiber框架错误
func handleFiberError(fiberErr *fiber.Error) interface{} {
	message := fiberErr.Message
	
	// 根据状态码提供更友好的错误信息
	switch fiberErr.Code {
	case fiber.StatusBadRequest:
		message = "请求参数错误"
	case fiber.StatusUnauthorized:
		message = "未授权访问"
	case fiber.StatusForbidden:
		message = "禁止访问"
	case fiber.StatusNotFound:
		message = "请求的资源不存在"
	case fiber.StatusMethodNotAllowed:
		message = "请求方法不被允许"
	case fiber.StatusRequestTimeout:
		message = "请求超时"
	case fiber.StatusTooManyRequests:
		message = "请求过于频繁，请稍后再试"
	case fiber.StatusInternalServerError:
		message = "内部服务器错误"
	case fiber.StatusBadGateway:
		message = "网关错误"
	case fiber.StatusServiceUnavailable:
		message = "服务暂时不可用"
	}

	return helpers.ErrorResponse(message, fiber.Map{
		"code": fmt.Sprintf("FIBER_%d", fiberErr.Code),
	})
}

// handleUnknownError 处理未知错误
func handleUnknownError(err error) interface{} {
	// 在生产环境中，不应该暴露详细的错误信息
	message := "内部服务器错误"
	
	// 在开发环境中，可以返回更详细的错误信息
	// 这里可以根据环境变量来判断
	if isDevelopment() {
		message = fmt.Sprintf("内部服务器错误: %s", err.Error())
	}

	return helpers.ErrorResponse(message, fiber.Map{
		"code": errors.ErrCodeInternalServer,
	})
}

// logError 记录错误日志
func logError(c *fiber.Ctx, err error, appErr *errors.AppError) {
	// 构建日志信息
	logMsg := fmt.Sprintf(
		"[%s] %s %s - %s",
		c.Method(),
		c.Path(),
		c.IP(),
		err.Error(),
	)

	// 如果是应用程序错误，添加错误码
	if appErr != nil {
		logMsg = fmt.Sprintf("%s (Code: %s)", logMsg, appErr.Code)
	}

	// 添加请求ID（如果存在）
	if requestID := c.Get("X-Request-ID"); requestID != "" {
		logMsg = fmt.Sprintf("%s [RequestID: %s]", logMsg, requestID)
	}

	// 根据错误类型选择日志级别
	if errors.IsAppError(err) {
		if appErr.StatusCode >= 500 {
			log.Printf("ERROR: %s", logMsg)
		} else {
			log.Printf("WARN: %s", logMsg)
		}
	} else {
		// 未知错误，记录堆栈信息
		log.Printf("ERROR: %s\n%s", logMsg, string(debug.Stack()))
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
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic信息
				log.Printf("PANIC: %s\n%s", r, string(debug.Stack()))
				
				// 返回内部服务器错误
				err := errors.InternalServerError("服务器内部错误")
				c.Status(fiber.StatusInternalServerError).JSON(helpers.ErrorResponse(err.Message, fiber.Map{
					"code": err.Code,
				}))
			}
		}()

		return c.Next()
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
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

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 这里可以使用UUID或其他方式生成唯一ID
	// 为了简单起见，使用时间戳和随机数
	return fmt.Sprintf("%d", time.Now().UnixNano())
}