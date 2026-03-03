package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/service"
)

// NewBouncingEmailMiddleware returns middleware that checks if the user's device email is blocked due to bouncing.
// If the device email is bouncing, it returns a 400 Bad Request with an error message.
// If service errors occur, it returns a 500 Internal Server Error.
func NewBouncingEmailMiddleware(svc service.Interface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountID := GetAccountID(r.Context())
			if accountID == "" {
				next.ServeHTTP(w, r)
				return
			}

			if shouldSkip, err := checkDeviceEmail(r.Context(), svc, accountID); shouldSkip || err != nil {
				if err != nil {
					sendError(w, err)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			destEmail, _, _ := svc.GetUserDeviceEmail(r.Context(), accountID)
			if handleBouncingEmail(r.Context(), svc, w, accountID, destEmail) {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func checkDeviceEmail(ctx context.Context, svc service.Interface, accountID string) (bool, error) {
	destEmail, _, err := svc.GetUserDeviceEmail(ctx, accountID)
	if err != nil {
		return false, fmt.Errorf("failed to get user device email: %s", err.Error())
	}

	if destEmail == "" {
		return true, nil
	}

	isBouncing, err := svc.IsEmailBouncing(ctx, accountID, destEmail)
	if err != nil {
		return false, fmt.Errorf("failed to check if email is bouncing: %s", err.Error())
	}

	return !isBouncing, nil
}

func handleBouncingEmail(
	ctx context.Context,
	svc service.Interface,
	w http.ResponseWriter,
	accountID, destEmail string,
) bool {
	if !isEmailBouncing(ctx, svc, accountID, destEmail) {
		return false
	}

	profile, profileErr := svc.GetUserProfile(ctx, accountID)
	if profileErr == nil && profile != nil && profile.BouncedEmails != nil {
		bounceInfo, exists := profile.BouncedEmails[destEmail]
		if exists {
			sendBounceError(w, destEmail, bounceInfo.Error)
			return true
		}
	}

	sendBounceError(w, destEmail, "")
	return true
}

func isEmailBouncing(ctx context.Context, svc service.Interface, accountID, destEmail string) bool {
	isBouncing, err := svc.IsEmailBouncing(ctx, accountID, destEmail)
	return err == nil && isBouncing
}

func sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
}

func sendBounceError(w http.ResponseWriter, email, errorMessage string) {
	if errorMessage != "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: fmt.Sprintf("device email %s is blocked due to previous bounce: %s", email, errorMessage),
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{
		Error: fmt.Sprintf("device email %s is blocked due to previous bounce", email),
	})
}
