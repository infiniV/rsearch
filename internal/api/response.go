package api

import (
	"encoding/json"
	"net/http"

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
