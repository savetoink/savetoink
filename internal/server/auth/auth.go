// Package auth provides pluggable authentication backends for the savetoink application.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

const (
	authHeader       = "Authorization"
	authHeaderPrefix = "Bearer "
	adminAccountID   = "admin"
	allowedClockSkew = 30 * time.Second
)

type contextKey string

const (
	accountIDKey contextKey = "account_id"
	authErrorKey contextKey = "auth_error"
)

// NewAccountIDMiddleware returns authentication middleware based on the configured auth backend.
// The returned middleware ensures the accountID is set in the context, adds authentication error
// to the context if any. To subsequently validate authentication use EnsureAutheticatedMiddleware.
func NewAccountIDMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	switch cfg.AuthBackend {
	case consts.AuthBackendSharedAPIKey:
		return sharedAPIKeyMiddleware(cfg.APIKeySecret)
	case consts.AuthBackendAuth0:
		return auth0Middleware(cfg.Auth0Domain, cfg.Auth0Audience)
	default:
		return sharedAPIKeyMiddleware(cfg.APIKeySecret)
	}
}

// EnsureAutheticatedMiddleware ensures that the request is authenticated before
// proceeding to the next handler.
func EnsureAutheticatedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := GetAuthError(r.Context()); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
			return
		}

		accountID := GetAccountID(r.Context())
		if accountID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetAccountID retrieves the authenticated account ID from the context.
func GetAccountID(ctx context.Context) string {
	accountID, _ := ctx.Value(accountIDKey).(string)
	return accountID
}

// GetAuthError retrieves the authentication error from the context, if any.
func GetAuthError(ctx context.Context) error {
	authError, found := ctx.Value(authErrorKey).(string)
	if found {
		return errors.New(authError)
	}
	return nil
}

func sharedAPIKeyMiddleware(apiKeySecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get(authHeader)
			if auth == "" || !strings.HasPrefix(auth, authHeaderPrefix) {
				handleAuthError(r.Context(), next, w, r, "missing or malformed auth header")
				return
			}
			token := strings.TrimPrefix(auth, authHeaderPrefix)
			if token != apiKeySecret {
				handleAuthError(r.Context(), next, w, r, "invalid API key")
				return
			}
			ctx := addAccountIDToContext(r.Context(), adminAccountID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func auth0Middleware(domain, audience string) func(http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil {
		panic("failed to parse issuer URL: " + err.Error())
	}

	provider, err := jwks.NewCachingProvider(
		jwks.WithIssuerURL(issuerURL),
	)
	if err != nil {
		panic("failed to create JWKS provider: " + err.Error())
	}

	jwtValidator, err := validator.New(
		validator.WithKeyFunc(provider.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
		validator.WithIssuer(issuerURL.String()),
		validator.WithAudience(audience),
		validator.WithAllowedClockSkew(allowedClockSkew),
	)
	if err != nil {
		panic("failed to create JWT validator: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get(authHeader)
			if auth == "" || !strings.HasPrefix(auth, authHeaderPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimPrefix(auth, authHeaderPrefix)
			claims, validateErr := jwtValidator.ValidateToken(r.Context(), token)
			if validateErr != nil {
				handleAuthError(r.Context(), next, w, r, "invalid JWT token: "+validateErr.Error())
				return
			}

			validatedClaims, ok := claims.(*validator.ValidatedClaims)
			if !ok {
				handleAuthError(r.Context(), next, w, r, "failed to parse JWT claims")
				return
			}

			ctx := addAccountIDToContext(r.Context(), validatedClaims.RegisteredClaims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func addAccountIDToContext(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDKey, accountID)
}

func handleAuthError(ctx context.Context, next http.Handler, w http.ResponseWriter, r *http.Request, msg string) {
	ctx = context.WithValue(ctx, authErrorKey, msg)
	next.ServeHTTP(w, r.WithContext(ctx))
}
