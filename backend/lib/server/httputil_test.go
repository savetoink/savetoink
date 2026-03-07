package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
)

const (
	invalidJSON = `{invalid json}`
)

func TestStatusCodeForError(t *testing.T) {
	t.Run("returns 404 for ErrNotFound", func(t *testing.T) {
		statusCode := statusCodeForError(apperrors.ErrNotFound)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("returns 400 for ErrInvalid", func(t *testing.T) {
		statusCode := statusCodeForError(apperrors.ErrInvalid)
		assert.Equal(t, http.StatusBadRequest, statusCode)
	})

	t.Run("returns 401 for ErrUnauthorized", func(t *testing.T) {
		statusCode := statusCodeForError(apperrors.ErrUnauthorized)
		assert.Equal(t, http.StatusUnauthorized, statusCode)
	})

	t.Run("returns 409 for ErrConflict", func(t *testing.T) {
		statusCode := statusCodeForError(apperrors.ErrConflict)
		assert.Equal(t, http.StatusConflict, statusCode)
	})

	t.Run("returns 500 for unknown errors", func(t *testing.T) {
		unknownErr := errors.New("unknown error")
		statusCode := statusCodeForError(unknownErr)
		assert.Equal(t, http.StatusInternalServerError, statusCode)
	})

	t.Run("returns 500 for nil error", func(t *testing.T) {
		statusCode := statusCodeForError(nil)
		assert.Equal(t, http.StatusInternalServerError, statusCode)
	})

	t.Run("returns 500 for generic error", func(t *testing.T) {
		genericErr := fmt.Errorf("generic error: %s", "something went wrong")
		statusCode := statusCodeForError(genericErr)
		assert.Equal(t, http.StatusInternalServerError, statusCode)
	})

	t.Run("errors.Is works correctly with wrapped errors", func(t *testing.T) {
		wrappedErr := fmt.Errorf("wrapped: %w", apperrors.ErrNotFound)
		statusCode := statusCodeForError(wrappedErr)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("handles wrapped ErrInvalid", func(t *testing.T) {
		wrappedErr := fmt.Errorf("validation failed: %w", apperrors.ErrInvalid)
		statusCode := statusCodeForError(wrappedErr)
		assert.Equal(t, http.StatusBadRequest, statusCode)
	})

	t.Run("handles wrapped ErrUnauthorized", func(t *testing.T) {
		wrappedErr := fmt.Errorf("auth failed: %w", apperrors.ErrUnauthorized)
		statusCode := statusCodeForError(wrappedErr)
		assert.Equal(t, http.StatusUnauthorized, statusCode)
	})

	t.Run("handles wrapped ErrConflict", func(t *testing.T) {
		wrappedErr := fmt.Errorf("duplicate entry: %w", apperrors.ErrConflict)
		statusCode := statusCodeForError(wrappedErr)
		assert.Equal(t, http.StatusConflict, statusCode)
	})

	t.Run("wrapped unknown error returns 500", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := fmt.Errorf("wrapper: %w", baseErr)
		statusCode := statusCodeForError(wrappedErr)
		assert.Equal(t, http.StatusInternalServerError, statusCode)
	})
}

func TestWriteJSONError(t *testing.T) {
	t.Run("writes correct HTTP status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("test error")

		writeJSONError(w, http.StatusBadRequest, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("response body contains error message", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("test error message")

		writeJSONError(w, http.StatusInternalServerError, err)

		assert.Contains(t, w.Body.String(), "test error message")
	})

	t.Run("response body is valid JSON with error field", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("test error")

		writeJSONError(w, http.StatusNotFound, err)

		var resp model.ErrorResponse
		err2 := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err2)
		assert.Equal(t, "test error", resp.Error)
	})

	t.Run("handles special characters in error message", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("error with \"quotes\" and 'apostrophes'")

		writeJSONError(w, http.StatusBadRequest, err)

		var resp model.ErrorResponse
		err2 := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err2)
		assert.Contains(t, resp.Error, "quotes")
	})

	t.Run("handles empty error message", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := errors.New("")

		writeJSONError(w, http.StatusBadRequest, err)

		var resp model.ErrorResponse
		err2 := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err2)
		assert.Equal(t, "", resp.Error)
	})

	t.Run("handles long error messages", func(t *testing.T) {
		longMsg := strings.Repeat("a", 10000)
		w := httptest.NewRecorder()
		err := errors.New(longMsg)

		writeJSONError(w, http.StatusInternalServerError, err)

		var resp model.ErrorResponse
		err2 := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err2)
		assert.Equal(t, longMsg, resp.Error)
	})

	t.Run("handles all HTTP status codes", func(t *testing.T) {
		statusCodes := []int{
			http.StatusContinue,
			http.StatusOK,
			http.StatusMultipleChoices,
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusConflict,
			http.StatusInternalServerError,
			http.StatusServiceUnavailable,
		}

		for _, statusCode := range statusCodes {
			t.Run(fmt.Sprintf("status %d", statusCode), func(t *testing.T) {
				w := httptest.NewRecorder()
				err := errors.New("test")

				writeJSONError(w, statusCode, err)

				assert.Equal(t, statusCode, w.Code)
			})
		}
	})
}

func TestDecodeAndValidateRequest(t *testing.T) {
	t.Run("successfully decodes valid JSON", func(t *testing.T) {
		type testRequest struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		body := `{"name":"test","value":123}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.Equal(t, "test", testReq.Name)
		assert.Equal(t, 123, testReq.Value)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		type testRequest struct {
			Name string `json:"name"`
		}

		body := invalidJSON
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode request body")
	})

	t.Run("writes 400 status on decode failure", func(t *testing.T) {
		type testRequest struct {
			Name string `json:"name"`
		}

		body := invalidJSON
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		_ = decodeAndValidateRequest(w, req, &testReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("includes decode error in response", func(t *testing.T) {
		type testRequest struct {
			Name string `json:"name"`
		}

		body := invalidJSON
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		_ = decodeAndValidateRequest(w, req, &testReq)

		var resp model.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Error, "failed to decode request body")
	})

	t.Run("handles empty JSON object", func(t *testing.T) {
		type testRequest struct {
			Name string `json:"name"`
		}

		body := `{}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.Equal(t, "", testReq.Name)
	})

	t.Run("handles JSON with null values", func(t *testing.T) {
		type testRequest struct {
			Name *string `json:"name"`
		}

		body := `{"name":null}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.Nil(t, testReq.Name)
	})

	t.Run("handles JSON with extra fields", func(t *testing.T) {
		type testRequest struct {
			Name string `json:"name"`
		}

		body := `{"name":"test","extra":"field"}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.Equal(t, "test", testReq.Name)
	})

	t.Run("handles JSON with numbers", func(t *testing.T) {
		type testRequest struct {
			Count int     `json:"count"`
			Price float64 `json:"price"`
		}

		body := `{"count":42,"price":19.99}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.Equal(t, 42, testReq.Count)
		assert.Equal(t, 19.99, testReq.Price)
	})

	t.Run("handles JSON with booleans", func(t *testing.T) {
		type testRequest struct {
			Active  bool `json:"active"`
			Deleted bool `json:"deleted"`
		}

		body := `{"active":true,"deleted":false}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.True(t, testReq.Active)
		assert.False(t, testReq.Deleted)
	})

	t.Run("handles JSON with arrays", func(t *testing.T) {
		type testRequest struct {
			Tags []string `json:"tags"`
		}

		body := `{"tags":["tag1","tag2","tag3"]}`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.NoError(t, err)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, testReq.Tags)
	})

	t.Run("handles malformed JSON - missing closing brace", func(t *testing.T) {
		type testRequest struct {
			Name string `json:"name"`
		}

		body := `{"name":"test"`
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", strings.NewReader(body))
		w := httptest.NewRecorder()

		var testReq testRequest
		err := decodeAndValidateRequest(w, req, &testReq)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleServiceError(t *testing.T) {
	t.Run("logs error with context string", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		err := apperrors.ErrNotFound

		handleServiceError(w, req, err, "article not found")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("writes appropriate status code for ErrNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)

		handleServiceError(w, req, apperrors.ErrNotFound, "resource not found")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("writes appropriate status code for ErrInvalid", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)

		handleServiceError(w, req, apperrors.ErrInvalid, "invalid input")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("writes appropriate status code for ErrUnauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)

		handleServiceError(w, req, apperrors.ErrUnauthorized, "unauthorized access")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("writes appropriate status code for ErrConflict", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)

		handleServiceError(w, req, apperrors.ErrConflict, "duplicate entry")

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("writes 500 for unknown errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		err := errors.New("unknown error")

		handleServiceError(w, req, err, "internal error")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("passes error to writeJSONError", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		err := errors.New("service error")

		handleServiceError(w, req, err, "operation failed")

		var resp model.ErrorResponse
		decodeErr := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, decodeErr)
		assert.Equal(t, "service error", resp.Error)
	})

	t.Run("handles wrapped errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		err := fmt.Errorf("wrapped: %w", apperrors.ErrNotFound)

		handleServiceError(w, req, err, "resource missing")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("context string is included in logged error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		err := errors.New("test error")

		handleServiceError(w, req, err, "context: operation failed")

		var resp model.ErrorResponse
		decodeErr := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, decodeErr)
		assert.Equal(t, "test error", resp.Error)
	})

	t.Run("handles errors with special characters", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		err := errors.New("error with special chars: \n\t\"quotes\"")

		handleServiceError(w, req, err, "context")

		var resp model.ErrorResponse
		decodeErr := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, decodeErr)
		assert.Contains(t, resp.Error, "special chars")
	})
}
