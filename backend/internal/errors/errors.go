package errors

import "fmt"

// Error codes for common scenarios
const (
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeInvalidInput       = "INVALID_INPUT"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeDatabaseError      = "DATABASE_ERROR"
	CodeInternalError      = "INTERNAL_ERROR"
)

// ReminiscerError represents a custom error with a code and message
type ReminiscerError struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *ReminiscerError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new ReminiscerError
func New(code string, message string) *ReminiscerError {
	return &ReminiscerError{
		Code:    code,
		Message: message,
	}
}

// NotFound creates a new not found error
func NotFound(message string) *ReminiscerError {
	return New(CodeNotFound, message)
}

// AlreadyExists creates a new already exists error
func AlreadyExists(message string) *ReminiscerError {
	return New(CodeAlreadyExists, message)
}

// InvalidInput creates a new invalid input error
func InvalidInput(message string) *ReminiscerError {
	return New(CodeInvalidInput, message)
}

// Unauthorized creates a new unauthorized error
func Unauthorized(message string) *ReminiscerError {
	return New(CodeUnauthorized, message)
}

// Forbidden creates a new forbidden error
func Forbidden(message string) *ReminiscerError {
	return New(CodeForbidden, message)
}

// InvalidToken creates a new invalid token error
func InvalidToken(message string) *ReminiscerError {
	return New(CodeInvalidToken, message)
}

// InvalidCredentials creates a new invalid credentials error
func InvalidCredentials(message string) *ReminiscerError {
	return New(CodeInvalidCredentials, message)
}

// DatabaseError creates a new database error
func DatabaseError(message string) *ReminiscerError {
	return New(CodeDatabaseError, message)
}

// InternalError creates a new internal error
func InternalError(message string) *ReminiscerError {
	return New(CodeInternalError, message)
}

// IsCode checks if an error is a ReminiscerError with the given code
func IsCode(err error, code string) bool {
	if reminiscerErr, ok := err.(*ReminiscerError); ok {
		return reminiscerErr.Code == code
	}
	return false
}

// GetCode returns the error code if it's a ReminiscerError, or an empty string if not
func GetCode(err error) string {
	if reminiscerErr, ok := err.(*ReminiscerError); ok {
		return reminiscerErr.Code
	}
	return ""
}

// GetMessage returns the error message if it's a ReminiscerError, or the error string if not
func GetMessage(err error) string {
	if reminiscerErr, ok := err.(*ReminiscerError); ok {
		return reminiscerErr.Message
	}
	return err.Error()
}

// Wrap wraps a standard error with a ReminiscerError
func Wrap(err error, code string) *ReminiscerError {
	return New(code, err.Error())
}
