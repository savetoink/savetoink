package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	apperrors "github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/stretchr/testify/assert"
)

func TestWriteJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("test error")

	WriteJSONError(w, http.StatusBadRequest, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	assert.Equal(t, "test error", response["error"])
}

func TestStatusCodeForError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "not found error",
			err:      apperrors.ErrNotFound,
			expected: http.StatusNotFound,
		},
		{
			name:     "invalid error",
			err:      apperrors.ErrInvalid,
			expected: http.StatusBadRequest,
		},
		{
			name:     "unauthorized error",
			err:      apperrors.ErrUnauthorized,
			expected: http.StatusUnauthorized,
		},
		{
			name:     "conflict error",
			err:      apperrors.ErrConflict,
			expected: http.StatusConflict,
		},
		{
			name:     "quota exceeded error",
			err:      apperrors.ErrQuotaExceeded,
			expected: http.StatusTooManyRequests,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "wrapped not found error",
			err:      fmt.Errorf("wrapped: %w", apperrors.ErrNotFound),
			expected: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statusCodeForError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeAndValidateRequest(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid JSON",
			body: TestRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid JSON",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "valid JSON with null values",
			body: TestRequest{
				Name:  "",
				Email: "",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			var bodyReader io.Reader
			if strBody, ok := tt.body.(string); ok {
				bodyReader = strings.NewReader(strBody)
			} else {
				bodyBytes, err := json.Marshal(tt.body)
				assert.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			r := httptest.NewRequest("POST", "/test", bodyReader)
			req := &TestRequest{}

			err := DecodeAndValidateRequest(w, r, req)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedStatus, w.Code)

				var response map[string]string
				if decodeErr := json.Unmarshal(w.Body.Bytes(), &response); decodeErr == nil {
					assert.Contains(t, response["error"], "failed to decode request body")
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestHandleServiceError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		contextStr     string
		expectedStatus int
	}{
		{
			name:           "not found error",
			err:            apperrors.ErrNotFound,
			contextStr:     "get article",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid error",
			err:            apperrors.ErrInvalid,
			contextStr:     "validate input",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized error",
			err:            apperrors.ErrUnauthorized,
			contextStr:     "check auth",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "conflict error",
			err:            apperrors.ErrConflict,
			contextStr:     "create resource",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "quota exceeded error",
			err:            apperrors.ErrQuotaExceeded,
			contextStr:     "check quota",
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "generic error",
			err:            errors.New("database connection failed"),
			contextStr:     "fetch data",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", http.NoBody)

			HandleServiceError(w, r, tt.err, tt.contextStr)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

func TestGetArticleID(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		paramName string
		paramVal  string
		expected  string
	}{
		{
			name:      "valid article ID in URL",
			url:       "/articles/12345",
			paramName: "id",
			paramVal:  "12345",
			expected:  "12345",
		},
		{
			name:      "empty article ID",
			url:       "/articles/",
			paramName: "id",
			paramVal:  "",
			expected:  "",
		},
		{
			name:      "UUID article ID",
			url:       "/articles/550e8400-e29b-41d4-a716-446655440000",
			paramName: "id",
			paramVal:  "550e8400-e29b-41d4-a716-446655440000",
			expected:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "missing parameter",
			url:       "/articles",
			paramName: "",
			paramVal:  "",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, http.NoBody)

			if tt.paramName != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add(tt.paramName, tt.paramVal)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			}

			result := GetArticleID(r)
			assert.Equal(t, tt.expected, result)
		})
	}
}
