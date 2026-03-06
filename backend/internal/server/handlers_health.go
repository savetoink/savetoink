// Package server provides HTTP handlers and middleware for the savetoink application.
package server

import (
	"encoding/json"
	"net/http"
)

func (h *handlers) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
