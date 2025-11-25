package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infiniv/rsearch/pkg/rsearch"
)

func TestRespondJSON(t *testing.T) {
	t.Run("responds with JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		RespondJSON(w, http.StatusOK, data)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", contentType)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response["message"] != "success" {
			t.Errorf("message = %v, want success", response["message"])
		}
	})

	t.Run("responds with nil data", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondJSON(w, http.StatusNoContent, nil)

		if w.Code != http.StatusNoContent {
			t.Errorf("status = %v, want %v", w.Code, http.StatusNoContent)
		}
	})
}

func TestRespondError(t *testing.T) {
	t.Run("responds with error", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondError(w, http.StatusBadRequest, rsearch.ErrorCodeParseError, "invalid query")

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeParseError {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeParseError)
		}

		if response.Error.Message != "invalid query" {
			t.Errorf("message = %v, want invalid query", response.Error.Message)
		}
	})
}

func TestRespondErrorWithDetails(t *testing.T) {
	t.Run("responds with error and details", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := []rsearch.ErrorInfo{
			{Position: 5, Message: "unexpected token"},
		}

		RespondErrorWithDetails(w, http.StatusBadRequest, rsearch.ErrorCodeParseError, "parse error", "field:test", details)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeParseError {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeParseError)
		}

		if response.Error.Message != "parse error" {
			t.Errorf("message = %v, want parse error", response.Error.Message)
		}

		if response.Error.Query != "field:test" {
			t.Errorf("query = %v, want field:test", response.Error.Query)
		}

		if len(response.Error.Details) != 1 {
			t.Fatalf("details length = %v, want 1", len(response.Error.Details))
		}

		if response.Error.Details[0].Position != 5 {
			t.Errorf("details[0].position = %v, want 5", response.Error.Details[0].Position)
		}
	})
}

func TestRespondInternalError(t *testing.T) {
	t.Run("responds with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondInternalError(w, "database connection failed")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %v, want %v", w.Code, http.StatusInternalServerError)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeInternalError {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeInternalError)
		}

		if response.Error.Message != "database connection failed" {
			t.Errorf("message = %v, want database connection failed", response.Error.Message)
		}
	})

	t.Run("responds with default message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondInternalError(w, "")

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Message != "An internal error occurred" {
			t.Errorf("message = %v, want An internal error occurred", response.Error.Message)
		}
	})
}

func TestRespondBadRequest(t *testing.T) {
	t.Run("responds with bad request", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondBadRequest(w, "invalid input")

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeParseError {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeParseError)
		}

		if response.Error.Message != "invalid input" {
			t.Errorf("message = %v, want invalid input", response.Error.Message)
		}
	})
}

func TestRespondNotFound(t *testing.T) {
	t.Run("responds with not found", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondNotFound(w, "schema not found")

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeSchemaNotFound {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeSchemaNotFound)
		}

		if response.Error.Message != "schema not found" {
			t.Errorf("message = %v, want schema not found", response.Error.Message)
		}
	})
}

func TestRespondRateLimited(t *testing.T) {
	t.Run("responds with rate limited", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondRateLimited(w, 60)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("status = %v, want %v", w.Code, http.StatusTooManyRequests)
		}

		retryAfter := w.Header().Get("Retry-After")
		if retryAfter == "" {
			t.Error("Retry-After header should be set")
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeRateLimited {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeRateLimited)
		}

		if response.Error.Message != "Rate limit exceeded" {
			t.Errorf("message = %v, want Rate limit exceeded", response.Error.Message)
		}
	})
}

func TestRespondUnauthorized(t *testing.T) {
	t.Run("responds with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondUnauthorized(w, "invalid token")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", w.Code, http.StatusUnauthorized)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeUnauthorized {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeUnauthorized)
		}

		if response.Error.Message != "invalid token" {
			t.Errorf("message = %v, want invalid token", response.Error.Message)
		}
	})

	t.Run("responds with default message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondUnauthorized(w, "")

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Message != "Unauthorized" {
			t.Errorf("message = %v, want Unauthorized", response.Error.Message)
		}
	})
}

func TestRespondForbidden(t *testing.T) {
	t.Run("responds with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondForbidden(w, "access denied")

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %v, want %v", w.Code, http.StatusForbidden)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeForbidden {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeForbidden)
		}

		if response.Error.Message != "access denied" {
			t.Errorf("message = %v, want access denied", response.Error.Message)
		}
	})

	t.Run("responds with default message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondForbidden(w, "")

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Message != "Forbidden" {
			t.Errorf("message = %v, want Forbidden", response.Error.Message)
		}
	})
}

func TestRespondTooManyRequests(t *testing.T) {
	t.Run("responds with custom message and retry after", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondTooManyRequests(w, "too many requests", 30)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("status = %v, want %v", w.Code, http.StatusTooManyRequests)
		}

		retryAfter := w.Header().Get("Retry-After")
		if retryAfter == "" {
			t.Error("Retry-After header should be set")
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeRateLimited {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeRateLimited)
		}

		if response.Error.Message != "too many requests" {
			t.Errorf("message = %v, want too many requests", response.Error.Message)
		}
	})

	t.Run("responds with default message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondTooManyRequests(w, "", 0)

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Message != "Too many requests" {
			t.Errorf("message = %v, want Too many requests", response.Error.Message)
		}
	})

	t.Run("does not set retry after header when zero", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondTooManyRequests(w, "rate limited", 0)

		retryAfter := w.Header().Get("Retry-After")
		if retryAfter != "" {
			t.Error("Retry-After header should not be set when retryAfter is 0")
		}
	})
}

func TestRespondServiceUnavailable(t *testing.T) {
	t.Run("responds with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondServiceUnavailable(w, "maintenance mode")

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %v, want %v", w.Code, http.StatusServiceUnavailable)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeServiceUnavailable {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeServiceUnavailable)
		}

		if response.Error.Message != "maintenance mode" {
			t.Errorf("message = %v, want maintenance mode", response.Error.Message)
		}
	})

	t.Run("responds with default message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondServiceUnavailable(w, "")

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Message != "Service unavailable" {
			t.Errorf("message = %v, want Service unavailable", response.Error.Message)
		}
	})
}

func TestRespondTimeout(t *testing.T) {
	t.Run("responds with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondTimeout(w, "query timeout")

		if w.Code != http.StatusGatewayTimeout {
			t.Errorf("status = %v, want %v", w.Code, http.StatusGatewayTimeout)
		}

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Code != rsearch.ErrorCodeTimeout {
			t.Errorf("code = %v, want %v", response.Error.Code, rsearch.ErrorCodeTimeout)
		}

		if response.Error.Message != "query timeout" {
			t.Errorf("message = %v, want query timeout", response.Error.Message)
		}
	})

	t.Run("responds with default message", func(t *testing.T) {
		w := httptest.NewRecorder()

		RespondTimeout(w, "")

		var response rsearch.ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Error.Message != "Request timeout" {
			t.Errorf("message = %v, want Request timeout", response.Error.Message)
		}
	})
}
