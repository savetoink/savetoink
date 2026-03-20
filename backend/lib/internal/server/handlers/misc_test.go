package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRobotsTXTHandler(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/robots.txt", http.NoBody)
	w := httptest.NewRecorder()

	RobotsTXTHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "User-agent: *")
	assert.Contains(t, body, "Allow: /v1/openapi.yaml")
	assert.Contains(t, body, "Disallow: /")
}

func TestOpenAPIHandler(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/v1/openapi.yaml", http.NoBody)
	w := httptest.NewRecorder()

	OpenAPIHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/yaml", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.NotEmpty(t, body, "OpenAPI spec should not be empty")
}
