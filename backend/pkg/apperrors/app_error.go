package apperrors

import "fmt"

// AppError represents a custom application error with an HTTP status code
// and user-friendly message. It can optionally wrap an internal error.
type AppError struct {
	Code     int    // HTTP status code (e.g., http.StatusBadRequest)
	Message  string // User-facing error message
	Internal error  // Original underlying error (for logging)
}

// Error returns the user-facing error message.
// This makes AppError satisfy the standard Go error interface.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the internal error, allowing for error chaining checks
// (e.g., using errors.Is or errors.As).
func (e *AppError) Unwrap() error {
	return e.Internal
}

// New creates a new AppError.
// The internal error is optional and can be nil.
func New(code int, message string, internal error) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		Internal: internal,
	}
}

// Newf creates a new AppError with formatted message.
// The internal error is optional and can be nil.
func Newf(code int, internal error, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:     code,
		Message:  fmt.Sprintf(format, args...),
		Internal: internal,
	}
}

// --- Predefined Errors (Examples - Add more as needed) ---

func ErrBadRequest(message string, internal error) *AppError {
	if message == "" {
		message = "Bad Request"
	}
	return New(400, message, internal)
}

func ErrUnauthorized(message string, internal error) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return New(401, message, internal)
}

func ErrForbidden(message string, internal error) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return New(403, message, internal)
}

func ErrNotFound(message string, internal error) *AppError {
	if message == "" {
		message = "Resource Not Found"
	}
	return New(404, message, internal)
}

func ErrMethodNotAllowed(message string, internal error) *AppError {
	if message == "" {
		message = "Method Not Allowed"
	}
	return New(405, message, internal)
}

func ErrInternalServer(message string, internal error) *AppError {
	if message == "" {
		message = "Internal Server Error"
	}
	return New(500, message, internal)
}
