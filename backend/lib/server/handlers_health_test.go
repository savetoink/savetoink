package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{}
	h := newHandlers(cfg, nil, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/health", http.NoBody)
	w := httptest.NewRecorder()

	h.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp healthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}
