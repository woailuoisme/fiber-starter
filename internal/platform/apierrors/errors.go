// Package apierrors defines error types and codes for the application.
package apierrors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

// 预定义错误码
const (
	ErrCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeTimeout        ErrorCode = "TIMEOUT_ERROR"

	// ErrCodeInvalidCredentials 认证相关错误码
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS" //nolint:gosec // Error code constant, not a credential
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserExists         ErrorCode = "USER_EXISTS"
	ErrCodeInvalidPassword    ErrorCode = "INVALID_PASSWORD"
	ErrCodeAccountLocked      ErrorCode = "ACCOUNT_LOCKED"

	// ErrCodeUserCreateFailed 用户相关错误码
	ErrCodeUserCreateFailed    ErrorCode = "USER_CREATE_FAILED"
	ErrCodeUserUpdateFailed    ErrorCode = "USER_UPDATE_FAILED"
	ErrCodeUserDeleteFailed    ErrorCode = "USER_DELETE_FAILED"
	ErrCodeProfileUpdateFailed ErrorCode = "PROFILE_UPDATE_FAILED"

	// ErrCodeDatabaseError 数据库相关错误码
	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrCodeRecordNotFound  ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeDuplicateEntry  ErrorCode = "DUPLICATE_ENTRY"
	ErrCodeForeignKeyError ErrorCode = "FOREIGN_KEY_ERROR"

	// ErrCodeBusinessLogicError 业务逻辑错误码
	ErrCodeBusinessLogicError     ErrorCode = "BUSINESS_LOGIC_ERROR"
	ErrCodeInsufficientPermission ErrorCode = "INSUFFICIENT_PERMISSION"
	ErrCodeResourceNotFound       ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeOperationNotAllowed    ErrorCode = "OPERATION_NOT_ALLOWED"

	// ErrCodeExternalServiceError 外部服务错误码
	ErrCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeEmailSendFailed      ErrorCode = "EMAIL_SEND_FAILED"
	ErrCodeSMSSendFailed        ErrorCode = "SMS_SEND_FAILED"
	ErrCodePaymentFailed        ErrorCode = "PAYMENT_FAILED"

	// ErrCodeFileUploadFailed 文件相关错误码
	ErrCodeFileUploadFailed ErrorCode = "FILE_UPLOAD_FAILED"
	ErrCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrCodeInvalidFileType  ErrorCode = "INVALID_FILE_TYPE"
	ErrCodeFileSizeExceeded ErrorCode = "FILE_SIZE_EXCEEDED"
)

// AppError 应用程序错误结构
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Cause      error     `json:"-"`
}

func newAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is 支持errors.Is
func (e *AppError) Is(target error) bool {
	if t, ok := errors.AsType[*AppError](target); ok {
		return e.Code == t.Code
	}
	return false
}

// NewAppError 创建新的应用程序错误
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return newAppError(code, message, statusCode)
}

// NewAppErrorWithDetails 创建带详细信息的应用程序错误
func NewAppErrorWithDetails(code ErrorCode, message, details string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, Details: details, StatusCode: statusCode}
}

// NewAppErrorWithCause 创建带原因的应用程序错误
func NewAppErrorWithCause(code ErrorCode, message string, statusCode int, cause error) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode, Cause: cause}
}

// NewAppErrorWithDetailsAndCause 创建带详细信息和原因的应用程序错误
func NewAppErrorWithDetailsAndCause(code ErrorCode, message, details string, statusCode int, cause error) *AppError {
	return &AppError{Code: code, Message: message, Details: details, StatusCode: statusCode, Cause: cause}
}

// 预定义错误创建函数

// InternalServerError 内部服务器错误
func InternalServerError(message string) *AppError {
	return NewAppError(ErrCodeInternalServer, message, http.StatusInternalServerError)
}

// InternalServerErrorWithCause 带原因的内部服务器错误
func InternalServerErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeInternalServer, message, http.StatusInternalServerError, cause)
}

// BadRequest 请求错误
func BadRequest(message string) *AppError {
	return NewAppError(ErrCodeBadRequest, message, http.StatusBadRequest)
}

// BadRequestWithDetails 带详细信息的请求错误
func BadRequestWithDetails(message, details string) *AppError {
	return NewAppErrorWithDetails(ErrCodeBadRequest, message, details, http.StatusBadRequest)
}

// Unauthorized 未授权错误
func Unauthorized(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden 禁止访问错误
func Forbidden(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden)
}

// NotFound 资源未找到错误
func NotFound(message string) *AppError {
	return NewAppError(ErrCodeNotFound, message, http.StatusNotFound)
}

// Conflict 冲突错误
func Conflict(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

// Validation 验证错误
func Validation(message string) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest)
}

// ValidationWithDetails 带详细信息的验证错误
func ValidationWithDetails(message, details string) *AppError {
	return NewAppErrorWithDetails(ErrCodeValidation, message, details, http.StatusBadRequest)
}

// 认证相关错误创建函数

// InvalidCredentials 无效凭据错误
func InvalidCredentials(message string) *AppError {
	return NewAppError(ErrCodeInvalidCredentials, message, http.StatusUnauthorized)
}

// TokenExpired 令牌过期错误
func TokenExpired(message string) *AppError {
	return NewAppError(ErrCodeTokenExpired, message, http.StatusUnauthorized)
}

// TokenInvalid 无效令牌错误
func TokenInvalid(message string) *AppError {
	return NewAppError(ErrCodeTokenInvalid, message, http.StatusUnauthorized)
}

// UserNotFound 用户未找到错误
func UserNotFound(message string) *AppError {
	return NewAppError(ErrCodeUserNotFound, message, http.StatusNotFound)
}

// UserExists 用户已存在错误
func UserExists(message string) *AppError {
	return NewAppError(ErrCodeUserExists, message, http.StatusConflict)
}

// InvalidPassword 无效密码错误
func InvalidPassword(message string) *AppError {
	return NewAppError(ErrCodeInvalidPassword, message, http.StatusBadRequest)
}

// 业务逻辑错误创建函数

// DatabaseError 数据库错误
func DatabaseError(message string) *AppError {
	return NewAppErrorWithCause(ErrCodeDatabaseError, message, http.StatusInternalServerError, nil)
}

// DatabaseErrorWithCause 带原因的数据库错误
func DatabaseErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeDatabaseError, message, http.StatusInternalServerError, cause)
}

// RecordNotFound 记录未找到错误
func RecordNotFound(message string) *AppError {
	return NewAppError(ErrCodeRecordNotFound, message, http.StatusNotFound)
}

// DuplicateEntry 重复条目错误
func DuplicateEntry(message string) *AppError {
	return NewAppError(ErrCodeDuplicateEntry, message, http.StatusConflict)
}

// BusinessLogicError 业务逻辑错误
func BusinessLogicError(message string) *AppError {
	return NewAppError(ErrCodeBusinessLogicError, message, http.StatusBadRequest)
}

// InsufficientPermission 权限不足错误
func InsufficientPermission(message string) *AppError {
	return NewAppError(ErrCodeInsufficientPermission, message, http.StatusForbidden)
}

// ExternalServiceError 外部服务错误
func ExternalServiceError(message string) *AppError {
	return NewAppError(ErrCodeExternalServiceError, message, http.StatusInternalServerError)
}

// ExternalServiceErrorWithCause 带原因的外部服务错误
func ExternalServiceErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeExternalServiceError, message, http.StatusInternalServerError, cause)
}

// IsAppError 检查是否为应用程序错误
func IsAppError(err error) bool {
	var appError *AppError
	return errors.As(err, &appError)
}

// GetAppError 获取应用程序错误
func GetAppError(err error) (*AppError, bool) {
	if appErr, ok := errors.AsType[*AppError](err); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError 包装普通错误为应用程序错误
func WrapError(err error, code ErrorCode, message string, statusCode int) *AppError {
	return NewAppErrorWithCause(code, message, statusCode, err)
}

// GetHTTPStatusCode 获取HTTP状态码
func GetHTTPStatusCode(err error) int {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
