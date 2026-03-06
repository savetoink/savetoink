package auth

import (
	"encoding/json"
	"net/http"

	"github.com/shaftoe/savetoink/backend/internal/config"
	"github.com/shaftoe/savetoink/backend/internal/consts"
	"github.com/shaftoe/savetoink/backend/internal/model"
)

// NewEmailBackendEnabledMiddleware returns middleware that checks if an email backend is configured.
// If no email backend is enabled, it returns a 400 Bad Request with an error message.
func NewEmailBackendEnabledMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.EmailProvider == "" || cfg.EmailProvider != consts.EmailBackendMailjet {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "email backend not configured"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
