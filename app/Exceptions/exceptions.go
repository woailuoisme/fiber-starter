package exceptions

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeTimeout        ErrorCode = "TIMEOUT_ERROR"

	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS" //nolint:gosec // Error code constant, not a credential
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserExists         ErrorCode = "USER_EXISTS"
	ErrCodeInvalidPassword    ErrorCode = "INVALID_PASSWORD"
	ErrCodeAccountLocked      ErrorCode = "ACCOUNT_LOCKED"

	ErrCodeUserCreateFailed    ErrorCode = "USER_CREATE_FAILED"
	ErrCodeUserUpdateFailed    ErrorCode = "USER_UPDATE_FAILED"
	ErrCodeUserDeleteFailed    ErrorCode = "USER_DELETE_FAILED"
	ErrCodeProfileUpdateFailed ErrorCode = "PROFILE_UPDATE_FAILED"

	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrCodeRecordNotFound  ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeDuplicateEntry  ErrorCode = "DUPLICATE_ENTRY"
	ErrCodeForeignKeyError ErrorCode = "FOREIGN_KEY_ERROR"

	ErrCodeBusinessLogicError     ErrorCode = "BUSINESS_LOGIC_ERROR"
	ErrCodeInsufficientPermission ErrorCode = "INSUFFICIENT_PERMISSION"
	ErrCodeResourceNotFound       ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeOperationNotAllowed    ErrorCode = "OPERATION_NOT_ALLOWED"

	ErrCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeEmailSendFailed      ErrorCode = "EMAIL_SEND_FAILED"
	ErrCodeSMSSendFailed        ErrorCode = "SMS_SEND_FAILED"
	ErrCodePaymentFailed        ErrorCode = "PAYMENT_FAILED"

	ErrCodeFileUploadFailed ErrorCode = "FILE_UPLOAD_FAILED"
	ErrCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrCodeInvalidFileType  ErrorCode = "INVALID_FILE_TYPE"
	ErrCodeFileSizeExceeded ErrorCode = "FILE_SIZE_EXCEEDED"
)

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

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) Is(target error) bool {
	if t, ok := errors.AsType[*AppError](target); ok {
		return e.Code == t.Code
	}
	return false
}

func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return newAppError(code, message, statusCode)
}

func NewAppErrorWithDetails(code ErrorCode, message, details string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, Details: details, StatusCode: statusCode}
}

func NewAppErrorWithCause(code ErrorCode, message string, statusCode int, cause error) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode, Cause: cause}
}

func NewAppErrorWithDetailsAndCause(code ErrorCode, message, details string, statusCode int, cause error) *AppError {
	return &AppError{Code: code, Message: message, Details: details, StatusCode: statusCode, Cause: cause}
}

func InternalServerError(message string) *AppError {
	return NewAppError(ErrCodeInternalServer, message, http.StatusInternalServerError)
}

func InternalServerErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeInternalServer, message, http.StatusInternalServerError, cause)
}

func BadRequest(message string) *AppError {
	return NewAppError(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func BadRequestWithDetails(message, details string) *AppError {
	return NewAppErrorWithDetails(ErrCodeBadRequest, message, details, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return NewAppError(ErrCodeNotFound, message, http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

func Validation(message string) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest)
}

func ValidationWithDetails(message, details string) *AppError {
	return NewAppErrorWithDetails(ErrCodeValidation, message, details, http.StatusBadRequest)
}

func InvalidCredentials(message string) *AppError {
	return NewAppError(ErrCodeInvalidCredentials, message, http.StatusUnauthorized)
}

func TokenExpired(message string) *AppError {
	return NewAppError(ErrCodeTokenExpired, message, http.StatusUnauthorized)
}

func TokenInvalid(message string) *AppError {
	return NewAppError(ErrCodeTokenInvalid, message, http.StatusUnauthorized)
}

func UserNotFound(message string) *AppError {
	return NewAppError(ErrCodeUserNotFound, message, http.StatusNotFound)
}

func UserExists(message string) *AppError {
	return NewAppError(ErrCodeUserExists, message, http.StatusConflict)
}

func InvalidPassword(message string) *AppError {
	return NewAppError(ErrCodeInvalidPassword, message, http.StatusBadRequest)
}

func DatabaseError(message string) *AppError {
	return NewAppErrorWithCause(ErrCodeDatabaseError, message, http.StatusInternalServerError, nil)
}

func DatabaseErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeDatabaseError, message, http.StatusInternalServerError, cause)
}

func RecordNotFound(message string) *AppError {
	return NewAppError(ErrCodeRecordNotFound, message, http.StatusNotFound)
}

func DuplicateEntry(message string) *AppError {
	return NewAppError(ErrCodeDuplicateEntry, message, http.StatusConflict)
}

func BusinessLogicError(message string) *AppError {
	return NewAppError(ErrCodeBusinessLogicError, message, http.StatusBadRequest)
}

func InsufficientPermission(message string) *AppError {
	return NewAppError(ErrCodeInsufficientPermission, message, http.StatusForbidden)
}

func ExternalServiceError(message string) *AppError {
	return NewAppError(ErrCodeExternalServiceError, message, http.StatusInternalServerError)
}

func ExternalServiceErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeExternalServiceError, message, http.StatusInternalServerError, cause)
}

func IsAppError(err error) bool {
	var appError *AppError
	return errors.As(err, &appError)
}

func GetAppError(err error) (*AppError, bool) {
	if appErr, ok := errors.AsType[*AppError](err); ok {
		return appErr, true
	}
	return nil, false
}

func GetAPIException(err error) (*APIException, bool) {
	if apiErr, ok := errors.AsType[*ValidationException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*AuthenticationException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*AuthorizationException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*NotFoundException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*BadRequestException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*ConflictException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*ServerException](err); ok {
		return apiErr.APIException, true
	}
	if apiErr, ok := errors.AsType[*APIException](err); ok {
		return apiErr, true
	}
	return nil, false
}

func WrapError(err error, code ErrorCode, message string, statusCode int) *AppError {
	return NewAppErrorWithCause(code, message, statusCode, err)
}

func GetHTTPStatusCode(err error) int {
	if appErr, ok := GetAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

type APIException struct {
	Message string
	Code    int
	Errors  map[string][]string
}

func newAPIException(message string, code int) *APIException {
	return &APIException{
		Message: message,
		Code:    code,
		Errors:  make(map[string][]string),
	}
}

func (e *APIException) Error() string {
	return e.Message
}

func NewAPIException(message string, code int) *APIException {
	return newAPIException(message, code)
}

func (e *APIException) WithMessage(message string) *APIException {
	e.Message = message
	return e
}

func (e *APIException) WithErrors(errors map[string][]string) *APIException {
	e.Errors = errors
	return e
}

func (e *APIException) WithCode(code int) *APIException {
	e.Code = code
	return e
}

func (e *APIException) AddError(field string, message string) *APIException {
	if e.Errors == nil {
		e.Errors = make(map[string][]string)
	}
	e.Errors[field] = append(e.Errors[field], message)
	return e
}

type ValidationException struct {
	*APIException
}

func NewValidationException(message string) *ValidationException {
	if message == "" {
		message = "Validation failed"
	}
	return &ValidationException{
		APIException: newAPIException(message, 422),
	}
}

func NewValidationExceptionWithErrors(message string, errors map[string][]string) *ValidationException {
	if message == "" {
		message = "Validation failed"
	}
	return &ValidationException{
		APIException: &APIException{Message: message, Code: 422, Errors: errors},
	}
}

type AuthenticationException struct {
	*APIException
}

func NewAuthenticationException(message string) *AuthenticationException {
	if message == "" {
		message = "Unauthenticated"
	}
	return &AuthenticationException{
		APIException: newAPIException(message, 401),
	}
}

type AuthorizationException struct {
	*APIException
}

func NewAuthorizationException(message string) *AuthorizationException {
	if message == "" {
		message = "Forbidden"
	}
	return &AuthorizationException{
		APIException: newAPIException(message, 403),
	}
}

type NotFoundException struct {
	*APIException
}

func NewNotFoundException(message string) *NotFoundException {
	if message == "" {
		message = "Resource not found"
	}
	return &NotFoundException{
		APIException: newAPIException(message, 404),
	}
}

func NewNotFoundExceptionf(format string, args ...interface{}) *NotFoundException {
	return NewNotFoundException(fmt.Sprintf(format, args...))
}

type BadRequestException struct {
	*APIException
}

func NewBadRequestException(message string) *BadRequestException {
	if message == "" {
		message = "Bad request"
	}
	return &BadRequestException{
		APIException: newAPIException(message, 400),
	}
}

type ConflictException struct {
	*APIException
}

func NewConflictException(message string) *ConflictException {
	if message == "" {
		message = "Conflict"
	}
	return &ConflictException{
		APIException: newAPIException(message, 409),
	}
}

type ServerException struct {
	*APIException
}

func NewServerException(message string) *ServerException {
	if message == "" {
		message = "Internal server error"
	}
	return &ServerException{
		APIException: newAPIException(message, 500),
	}
}

func NewServerExceptionFromError(err error) *ServerException {
	return &ServerException{
		APIException: newAPIException(err.Error(), 500),
	}
}
