// Package auth provides pluggable authentication backends for the savetoink application.
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/paseto"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

const (
	authHeader       = "Authorization"
	authHeaderPrefix = "Bearer "
	adminAccountID   = "admin"
)

// NewAccountIDMiddleware returns authentication middleware based on the configured auth backend.
// The returned middleware ensures the accountID is set in the context, adds authentication error
// to the context if any. To subsequently validate authentication use EnsureAutheticatedMiddleware.
func NewAccountIDMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	switch cfg.AuthBackend {
	case consts.AuthBackendSharedAPIKey:
		return sharedAPIKeyMiddleware(cfg.APIKeySecret)
	case consts.AuthBackendAuth0:
		keyStore := newPasetoKeyStore(cfg)
		return pasetoMiddleware(keyStore)
	default:
		return sharedAPIKeyMiddleware(cfg.APIKeySecret)
	}
}

// EnsureAutheticatedMiddleware ensures that the request is authenticated before
// proceeding to the next handler.
func EnsureAutheticatedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth.GetAuthErrorFromCtx(r.Context()); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
			return
		}

		accountID := auth.GetAccountIDFromCtx(r.Context())
		if accountID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func sharedAPIKeyMiddleware(apiKeySecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(authHeader)
			if authHeader == "" {
				handleAuthError(r.Context(), next, w, r, "missing auth header")
				return
			}
			if !strings.HasPrefix(authHeader, authHeaderPrefix) {
				handleAuthError(r.Context(), next, w, r, "malformed auth header")
				return
			}
			token := strings.TrimPrefix(authHeader, authHeaderPrefix)
			if token != apiKeySecret {
				handleAuthError(r.Context(), next, w, r, "invalid API key")
				return
			}
			ctx := auth.AddAccountIDToCtx(r.Context(), adminAccountID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func pasetoMiddleware(keyStore *paseto.KeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeaderVal := r.Header.Get(authHeader)
			if authHeaderVal == "" || !strings.HasPrefix(authHeaderVal, authHeaderPrefix) {
				// No auth header — let EnsureAutheticatedMiddleware handle it
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimPrefix(authHeaderVal, authHeaderPrefix)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, validateErr := keyStore.ValidateToken(token)
			if validateErr != nil {
				handleAuthError(r.Context(), next, w, r, "invalid token: "+validateErr.Error())
				return
			}

			ctx := auth.AddAccountIDToCtx(r.Context(), claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func newPasetoKeyStore(cfg *config.Config) *paseto.KeyStore {
	keyStore, err := paseto.NewKeyStore(paseto.KeyStoreConfig{
		SymmetricKey: cfg.PASETOSymmetricKey,
		KeyVersion:   cfg.PASETOKeyVersion,
	})
	if err != nil {
		panic("failed to create paseto key store: " + err.Error())
	}
	return keyStore
}

func handleAuthError(ctx context.Context, next http.Handler, w http.ResponseWriter, r *http.Request, msg string) {
	ctx = auth.AddAuthErrorToCtx(ctx, msg)
	next.ServeHTTP(w, r.WithContext(ctx))
}
