package httperr

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// ErrorResponse defines the structure for JSON error responses
type ErrorResponse struct {
	Error string `json:"error"`
}

// HTTPError represents an error with an associated HTTP status code and user message
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// AppHandler is a custom handler type that returns an error
type AppHandler func(http.ResponseWriter, *http.Request) error

// ErrorHandler adapts an AppHandler to a standard http.HandlerFunc and handles errors
func ErrorHandler(handler AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			// Default error values
			statusCode := http.StatusInternalServerError
			userMessage := "An internal server error occurred"
			logErr := err

			// Check if it's our custom error type
			if err, ok := err.(*HTTPError); ok {
				statusCode = err.StatusCode
				userMessage = err.Message
				if err.Err != nil {
					logErr = err.Err
				}
			}

			// Log the error with details and stack trace
			errMsg := "no underlying error provided"
			if logErr != nil {
				errMsg = logErr.Error()
			}

			log.Printf(`
ERROR: %s %s
Status: %d %s
Handler Error: %v
Underlying Error: %v
Stack trace:
%s`,
				r.Method, r.URL.Path,
				statusCode, http.StatusText(statusCode),
				err,
				errMsg,
				debug.Stack(),
			)

			// Send JSON error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			response := ErrorResponse{Error: userMessage}

			if jsonErr := json.NewEncoder(w).Encode(response); jsonErr != nil {
				log.Printf("ERROR: Could not encode error response JSON: %v", jsonErr)
				http.Error(w, `{"error":"Failed to encode error response"}`, http.StatusInternalServerError)
			}
		}
	}
}

// NewHTTPError creates a new HTTPError instance
func NewHTTPError(statusCode int, message string, err error) *HTTPError {
	if message == "" {
		message = http.StatusText(statusCode)
		if message == "" {
			message = "An unexpected error occurred"
		}
	}
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// NewBadRequest creates a 400 Bad Request error
func NewBadRequest(err error, userMessage string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, userMessage, err)
}

// NewUnauthorized creates a 401 Unauthorized error
func NewUnauthorized(err error, userMessage string) *HTTPError {
	if userMessage == "" {
		userMessage = "Authentication required"
	}
	return NewHTTPError(http.StatusUnauthorized, userMessage, err)
}

// NewForbidden creates a 403 Forbidden error
func NewForbidden(err error, userMessage string) *HTTPError {
	if userMessage == "" {
		userMessage = "You do not have permission to perform this action"
	}
	return NewHTTPError(http.StatusForbidden, userMessage, err)
}

// NewNotFound creates a 404 Not Found error
func NewNotFound(err error, userMessage string) *HTTPError {
	if userMessage == "" {
		userMessage = "The requested resource was not found"
	}
	return NewHTTPError(http.StatusNotFound, userMessage, err)
}

// NewMethodNotAllowed creates a 405 Method Not Allowed error
func NewMethodNotAllowed(err error, userMessage string) *HTTPError {
	if userMessage == "" {
		userMessage = "Method not allowed for this resource"
	}
	return NewHTTPError(http.StatusMethodNotAllowed, userMessage, err)
}

// NewInternalServerError creates a 500 Internal Server Error
func NewInternalServerError(err error, userMessage string) *HTTPError {
	if userMessage == "" {
		userMessage = "An internal server error occurred"
	}
	return NewHTTPError(http.StatusInternalServerError, userMessage, err)
}
