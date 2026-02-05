package errors

import "fmt"

// ErrorCode represents error codes
type ErrorCode int

const (
	BAD_REQUEST         ErrorCode = 400
	UNAUTHORIZED        ErrorCode = 401
	FORBIDDEN           ErrorCode = 403
	NOT_FOUND           ErrorCode = 404
	CONFLICT_ERROR      ErrorCode = 409
	INTERNAL_ERROR      ErrorCode = 500
	SERVICE_UNAVAILABLE ErrorCode = 503
)

// BaseError is the base error interface
type BaseError interface {
	Error() string
	GetCode() ErrorCode
}

// baseError implements BaseError
type baseError struct {
	code ErrorCode
	err  error
}

// NewBaseError creates a new base error
func NewBaseError(code ErrorCode, err error) BaseError {
	return &baseError{
		code: code,
		err:  err,
	}
}

// Error returns the error message
func (e *baseError) Error() string {
	return e.err.Error()
}

// GetCode returns the error code
func (e *baseError) GetCode() ErrorCode {
	return e.code
}

// Common error constructors
func BadRequest(message string) BaseError {
	return NewBaseError(BAD_REQUEST, fmt.Errorf(message))
}

func NotFound(message string) BaseError {
	return NewBaseError(NOT_FOUND, fmt.Errorf(message))
}

func Conflict(message string) BaseError {
	return NewBaseError(CONFLICT_ERROR, fmt.Errorf(message))
}

func Internal(err error) BaseError {
	return NewBaseError(INTERNAL_ERROR, err)
}
