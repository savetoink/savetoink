// Package handlers provides HTTP handlers for the savetoink application.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shaftoe/savetoink/backend/lib/server/types"
)

// HandleHealth handles the health check endpoint.
func (h *Handlers) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.HealthResponse{Status: "ok"})
}
