package helpers

import (
	"net/http"
	"strconv"
)

const (
	DefaultLimit  = 10
	DefaultOffset = 0
	MaxLimit      = 100 // Optional: Set a maximum limit
)

// GetPaginationParams extracts limit and offset from query parameters.
// It applies default values and potentially a maximum limit.
func GetPaginationParams(r *http.Request) (limit, offset int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit = DefaultLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			// Optional: Enforce a maximum limit
			// if limit > MaxLimit {
			// 	limit = MaxLimit
			// }
		}
	}

	offset = DefaultOffset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}
