// Package exceptions 定义应用程序中使用的异常类型
package exceptions

import (
	"fmt"
)

func newAPIException(message string, code int) *APIException {
	return &APIException{
		Message: message,
		Code:    code,
		Errors:  make(map[string][]string),
	}
}

// APIException 基础异常类型
// Requirements: 11.1, 11.3, 11.11, 11.12, 11.13
type APIException struct {
	Message string
	Code    int
	Errors  map[string][]string
}

// Error 实现 error 接口
func (e *APIException) Error() string {
	return e.Message
}

// NewAPIException 创建新的 API 异常
func NewAPIException(message string, code int) *APIException {
	return newAPIException(message, code)
}

// WithMessage 设置异常消息（链式调用）
// Requirements: 11.12
func (e *APIException) WithMessage(message string) *APIException {
	e.Message = message
	return e
}

// WithErrors 设置错误详情（链式调用）
// Requirements: 11.12
func (e *APIException) WithErrors(errors map[string][]string) *APIException {
	e.Errors = errors
	return e
}

// WithCode 设置状态码（链式调用）
// Requirements: 11.12
func (e *APIException) WithCode(code int) *APIException {
	e.Code = code
	return e
}

// AddError 添加单个字段错误
func (e *APIException) AddError(field string, message string) *APIException {
	if e.Errors == nil {
		e.Errors = make(map[string][]string)
	}
	e.Errors[field] = append(e.Errors[field], message)
	return e
}

// ValidationException 验证异常 (422)
// Requirements: 11.2, 11.4
type ValidationException struct {
	*APIException
}

// NewValidationException 创建验证异常
func NewValidationException(message string) *ValidationException {
	if message == "" {
		message = "Validation failed"
	}
	return &ValidationException{
		APIException: newAPIException(message, 422),
	}
}

// NewValidationExceptionWithErrors 创建带错误详情的验证异常
func NewValidationExceptionWithErrors(message string, errors map[string][]string) *ValidationException {
	if message == "" {
		message = "Validation failed"
	}
	return &ValidationException{
		APIException: &APIException{Message: message, Code: 422, Errors: errors},
	}
}

// AuthenticationException 认证异常 (401)
// Requirements: 11.2, 11.5
type AuthenticationException struct {
	*APIException
}

// NewAuthenticationException 创建认证异常
func NewAuthenticationException(message string) *AuthenticationException {
	if message == "" {
		message = "Unauthenticated"
	}
	return &AuthenticationException{
		APIException: newAPIException(message, 401),
	}
}

// AuthorizationException 授权异常 (403)
// Requirements: 11.2, 11.6
type AuthorizationException struct {
	*APIException
}

// NewAuthorizationException 创建授权异常
func NewAuthorizationException(message string) *AuthorizationException {
	if message == "" {
		message = "Forbidden"
	}
	return &AuthorizationException{
		APIException: newAPIException(message, 403),
	}
}

// NotFoundException 未找到异常 (404)
// Requirements: 11.2, 11.7
type NotFoundException struct {
	*APIException
}

// NewNotFoundException 创建未找到异常
func NewNotFoundException(message string) *NotFoundException {
	if message == "" {
		message = "Resource not found"
	}
	return &NotFoundException{
		APIException: newAPIException(message, 404),
	}
}

// NewNotFoundExceptionf 创建格式化的未找到异常
func NewNotFoundExceptionf(format string, args ...interface{}) *NotFoundException {
	return NewNotFoundException(fmt.Sprintf(format, args...))
}

// BadRequestException 错误请求异常 (400)
// Requirements: 11.2, 11.8
type BadRequestException struct {
	*APIException
}

// NewBadRequestException 创建错误请求异常
func NewBadRequestException(message string) *BadRequestException {
	if message == "" {
		message = "Bad request"
	}
	return &BadRequestException{
		APIException: newAPIException(message, 400),
	}
}

// ConflictException 冲突异常 (409)
// Requirements: 11.2, 11.9
type ConflictException struct {
	*APIException
}

// NewConflictException 创建冲突异常
func NewConflictException(message string) *ConflictException {
	if message == "" {
		message = "Conflict"
	}
	return &ConflictException{
		APIException: newAPIException(message, 409),
	}
}

// ServerException 服务器异常 (500)
// Requirements: 11.2, 11.10
type ServerException struct {
	*APIException
}

// NewServerException 创建服务器异常
func NewServerException(message string) *ServerException {
	if message == "" {
		message = "Internal server error"
	}
	return &ServerException{
		APIException: newAPIException(message, 500),
	}
}

// NewServerExceptionFromError 从 error 创建服务器异常
func NewServerExceptionFromError(err error) *ServerException {
	return &ServerException{
		APIException: newAPIException(err.Error(), 500),
	}
}
