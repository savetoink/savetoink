package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/server/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{}
	h := New(cfg, nil, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/health", http.NoBody)
	w := httptest.NewRecorder()

	h.HandleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}
