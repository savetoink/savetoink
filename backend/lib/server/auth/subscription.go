package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

func noOpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// NewActiveSubscriptionMiddleware returns authorization middleware based on the configured auth backend.
// The returned middleware ensures the accountID subscription is active.
func NewActiveSubscriptionMiddleware(cfg *config.Config, svc service.Interface) func(http.Handler) http.Handler {
	switch cfg.AuthBackend {
	case consts.AuthBackendSharedAPIKey:
		return noOpMiddleware
	case consts.AuthBackendAuth0:
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				accountID := GetAccountID(r.Context())
				startDate := time.Now().AddDate(0, 0, -consts.FreeTierSendPeriodDays)

				count, err := svc.CountSendsByAccountDateRange(r.Context(), accountID, startDate, time.Now())
				if err != nil {
					r = r.WithContext(context.WithValue(r.Context(), authErrorKey, "failed to check subscription limit: "+err.Error()))
					next.ServeHTTP(w, r)
					return
				}

				if count >= consts.MaxFreeTierSendsPerPeriod {
					w.WriteHeader(http.StatusTooManyRequests)
					_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "free tier limit exceeded"})
					return
				}

				ctx := context.WithValue(r.Context(), sendsCountKey, count)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		}
	default:
		return noOpMiddleware
	}
}
