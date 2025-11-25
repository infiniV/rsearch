package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/infiniv/rsearch/pkg/rsearch"
)

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but don't try to write another response
			return
		}
	}
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondJSON(w, status, rsearch.ErrorResponse{
		Error: rsearch.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// RespondErrorWithDetails sends an error response with details
func RespondErrorWithDetails(w http.ResponseWriter, status int, code, message, query string, details []rsearch.ErrorInfo) {
	RespondJSON(w, status, rsearch.ErrorResponse{
		Error: rsearch.ErrorDetail{
			Code:    code,
			Message: message,
			Query:   query,
			Details: details,
		},
	})
}

// RespondInternalError sends a 500 internal server error
func RespondInternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "An internal error occurred"
	}
	RespondError(w, http.StatusInternalServerError, rsearch.ErrorCodeInternalError, message)
}

// RespondBadRequest sends a 400 bad request error
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, rsearch.ErrorCodeParseError, message)
}

// RespondNotFound sends a 404 not found error
func RespondNotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, rsearch.ErrorCodeSchemaNotFound, message)
}

// RespondRateLimited sends a 429 rate limited error
func RespondRateLimited(w http.ResponseWriter, retryAfter int) {
	if retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	}
	RespondError(w, http.StatusTooManyRequests, rsearch.ErrorCodeRateLimited, "Rate limit exceeded")
}

// RespondUnauthorized sends a 401 unauthorized error
func RespondUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	RespondError(w, http.StatusUnauthorized, rsearch.ErrorCodeUnauthorized, message)
}

// RespondForbidden sends a 403 forbidden error
func RespondForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	RespondError(w, http.StatusForbidden, rsearch.ErrorCodeForbidden, message)
}

// RespondTooManyRequests sends a 429 too many requests error with retry after header
func RespondTooManyRequests(w http.ResponseWriter, message string, retryAfter int) {
	if message == "" {
		message = "Too many requests"
	}
	if retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	}
	RespondError(w, http.StatusTooManyRequests, rsearch.ErrorCodeRateLimited, message)
}

// RespondServiceUnavailable sends a 503 service unavailable error
func RespondServiceUnavailable(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Service unavailable"
	}
	RespondError(w, http.StatusServiceUnavailable, rsearch.ErrorCodeServiceUnavailable, message)
}

// RespondTimeout sends a 504 gateway timeout error
func RespondTimeout(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Request timeout"
	}
	RespondError(w, http.StatusGatewayTimeout, rsearch.ErrorCodeTimeout, message)
}
