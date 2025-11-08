package errors

import "fmt"

type ErrorCode string

const (
	ErrorCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists   ErrorCode = "ALREADY_EXISTS"
	ErrorCodeDatabase        ErrorCode = "DATABASE_ERROR"
	ErrorCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrorCodeTimeout         ErrorCode = "TIMEOUT"
	ErrorCodeNetworkError    ErrorCode = "NETWORK_ERROR"
	ErrorCodeInvalidResponse ErrorCode = "INVALID_RESPONSE"
	ErrorCodeRateLimit       ErrorCode = "RATE_LIMIT"
	ErrorCodeServerError     ErrorCode = "SERVER_ERROR"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message string) *AppError {
	return New(ErrorCodeValidation, message)
}

func NewNotFoundError(message string) *AppError {
	return New(ErrorCodeNotFound, message)
}

func NewDatabaseError(err error) *AppError {
	return Wrap(ErrorCodeDatabase, "database operation failed", err)
}

func NewInternalError(err error) *AppError {
	return Wrap(ErrorCodeInternal, "internal server error", err)
}
