package exceptions

import (
	"fmt"
)

// ApiException 基础异常类型
// Requirements: 11.1, 11.3, 11.11, 11.12, 11.13
type ApiException struct {
	Message string
	Code    int
	Errors  map[string][]string
}

// Error 实现 error 接口
func (e *ApiException) Error() string {
	return e.Message
}

// NewApiException 创建新的 API 异常
func NewApiException(message string, code int) *ApiException {
	return &ApiException{
		Message: message,
		Code:    code,
		Errors:  make(map[string][]string),
	}
}

// WithMessage 设置异常消息（链式调用）
// Requirements: 11.12
func (e *ApiException) WithMessage(message string) *ApiException {
	e.Message = message
	return e
}

// WithErrors 设置错误详情（链式调用）
// Requirements: 11.12
func (e *ApiException) WithErrors(errors map[string][]string) *ApiException {
	e.Errors = errors
	return e
}

// WithCode 设置状态码（链式调用）
// Requirements: 11.12
func (e *ApiException) WithCode(code int) *ApiException {
	e.Code = code
	return e
}

// AddError 添加单个字段错误
func (e *ApiException) AddError(field string, message string) *ApiException {
	if e.Errors == nil {
		e.Errors = make(map[string][]string)
	}
	e.Errors[field] = append(e.Errors[field], message)
	return e
}

// ValidationException 验证异常 (422)
// Requirements: 11.2, 11.4
type ValidationException struct {
	*ApiException
}

// NewValidationException 创建验证异常
func NewValidationException(message string) *ValidationException {
	if message == "" {
		message = "Validation failed"
	}
	return &ValidationException{
		ApiException: &ApiException{
			Message: message,
			Code:    422,
			Errors:  make(map[string][]string),
		},
	}
}

// NewValidationExceptionWithErrors 创建带错误详情的验证异常
func NewValidationExceptionWithErrors(message string, errors map[string][]string) *ValidationException {
	if message == "" {
		message = "Validation failed"
	}
	return &ValidationException{
		ApiException: &ApiException{
			Message: message,
			Code:    422,
			Errors:  errors,
		},
	}
}

// AuthenticationException 认证异常 (401)
// Requirements: 11.2, 11.5
type AuthenticationException struct {
	*ApiException
}

// NewAuthenticationException 创建认证异常
func NewAuthenticationException(message string) *AuthenticationException {
	if message == "" {
		message = "Unauthenticated"
	}
	return &AuthenticationException{
		ApiException: &ApiException{
			Message: message,
			Code:    401,
			Errors:  make(map[string][]string),
		},
	}
}

// AuthorizationException 授权异常 (403)
// Requirements: 11.2, 11.6
type AuthorizationException struct {
	*ApiException
}

// NewAuthorizationException 创建授权异常
func NewAuthorizationException(message string) *AuthorizationException {
	if message == "" {
		message = "Forbidden"
	}
	return &AuthorizationException{
		ApiException: &ApiException{
			Message: message,
			Code:    403,
			Errors:  make(map[string][]string),
		},
	}
}

// NotFoundException 未找到异常 (404)
// Requirements: 11.2, 11.7
type NotFoundException struct {
	*ApiException
}

// NewNotFoundException 创建未找到异常
func NewNotFoundException(message string) *NotFoundException {
	if message == "" {
		message = "Resource not found"
	}
	return &NotFoundException{
		ApiException: &ApiException{
			Message: message,
			Code:    404,
			Errors:  make(map[string][]string),
		},
	}
}

// NewNotFoundExceptionf 创建格式化的未找到异常
func NewNotFoundExceptionf(format string, args ...interface{}) *NotFoundException {
	return NewNotFoundException(fmt.Sprintf(format, args...))
}

// BadRequestException 错误请求异常 (400)
// Requirements: 11.2, 11.8
type BadRequestException struct {
	*ApiException
}

// NewBadRequestException 创建错误请求异常
func NewBadRequestException(message string) *BadRequestException {
	if message == "" {
		message = "Bad request"
	}
	return &BadRequestException{
		ApiException: &ApiException{
			Message: message,
			Code:    400,
			Errors:  make(map[string][]string),
		},
	}
}

// ConflictException 冲突异常 (409)
// Requirements: 11.2, 11.9
type ConflictException struct {
	*ApiException
}

// NewConflictException 创建冲突异常
func NewConflictException(message string) *ConflictException {
	if message == "" {
		message = "Conflict"
	}
	return &ConflictException{
		ApiException: &ApiException{
			Message: message,
			Code:    409,
			Errors:  make(map[string][]string),
		},
	}
}

// ServerException 服务器异常 (500)
// Requirements: 11.2, 11.10
type ServerException struct {
	*ApiException
}

// NewServerException 创建服务器异常
func NewServerException(message string) *ServerException {
	if message == "" {
		message = "Internal server error"
	}
	return &ServerException{
		ApiException: &ApiException{
			Message: message,
			Code:    500,
			Errors:  make(map[string][]string),
		},
	}
}

// NewServerExceptionFromError 从 error 创建服务器异常
func NewServerExceptionFromError(err error) *ServerException {
	return &ServerException{
		ApiException: &ApiException{
			Message: err.Error(),
			Code:    500,
			Errors:  make(map[string][]string),
		},
	}
}
