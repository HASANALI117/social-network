package helpers

import (
	"encoding/json"
	"errors"
	"log" // Consider using a more structured logger in a real application
	"net/http"

	"github.com/HASANALI117/social-network/pkg/apperrors"
)

// respondJSON writes a JSON response with the given status code and payload.
func RespondJSON(w http.ResponseWriter, code int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			// Log the encoding error, but we can't write another header
			log.Printf("ERROR: Failed to encode JSON response: %v", err)
			return err // Return the error so the wrapper knows something went wrong
		}
	}
	return nil
}

// respondError writes a JSON error response. It checks if the error is an AppError
// to use its specific code and message, otherwise defaults to 500.
func RespondError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		// Log the internal error if it exists
		if appErr.Internal != nil {
			log.Printf("ERROR (internal): %v", appErr.Internal)
		}
		// Use the AppError's code and message for the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.Code)
		json.NewEncoder(w).Encode(map[string]string{"error": appErr.Message})
	} else {
		// Log the unexpected error
		log.Printf("ERROR (unexpected): %v", err)
		// Respond with a generic 500 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
	}
}

// AppHandler defines the signature for application handlers that return an error.
type AppHandler func(http.ResponseWriter, *http.Request) error

// MakeHandler creates an http.HandlerFunc from an AppHandler.
// It executes the AppHandler and handles any returned error using respondError.
func MakeHandler(fn AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r) // Execute the actual handler logic
		if err != nil {
			// If an error occurred, respond appropriately
			RespondError(w, err)
		}
		// If err is nil, the handler likely already wrote a successful response
		// using RespondJSON, or it was a non-error scenario (e.g., redirect).
	}
}
