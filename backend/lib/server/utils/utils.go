package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service"
)

// WriteJSONError writes an error response with given status code and error message.
func WriteJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
}

// statusCodeForError returns the appropriate HTTP status code for the given error.
func statusCodeForError(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperrors.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperrors.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// DecodeAndValidateRequest decodes JSON from request body and handles errors.
func DecodeAndValidateRequest(w http.ResponseWriter, r *http.Request, req any) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		decodeErr := fmt.Errorf("failed to decode request body: %w", err)
		WriteJSONError(w, http.StatusBadRequest, decodeErr)
		return decodeErr
	}
	return nil
}

// HandleServiceError logs error and writes appropriate response.
func HandleServiceError(w http.ResponseWriter, r *http.Request, err error, contextStr string) {
	logging.AddRequestError(r.Context(), fmt.Errorf("%s: %w", contextStr, err))
	WriteJSONError(w, statusCodeForError(err), err)
}

// CheckEmailBackendEnabled checks if email backend is configured.
func CheckEmailBackendEnabled(w http.ResponseWriter, r *http.Request, emailProvider consts.EmailProvider) error {
	if emailProvider == "" || emailProvider != consts.EmailBackendMailjet {
		backendErr := fmt.Errorf("email backend not configured: %w", apperrors.ErrInvalid)
		HandleServiceError(w, r, backendErr, "check email backend")
		return backendErr
	}
	return nil
}

// CheckQuotaAndDeviceEmail checks if user has exceeded quota and if device email is bouncing.
func CheckQuotaAndDeviceEmail(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	svc service.Interface,
	authBackend consts.AuthBackend,
	accountID string,
) (int, error) {
	sendsCount, err := checkQuota(ctx, w, r, svc, authBackend, accountID)
	if err != nil {
		return sendsCount, err
	}

	if emailErr := checkDeviceEmail(ctx, w, r, svc, accountID); emailErr != nil {
		return sendsCount, emailErr
	}

	return sendsCount, nil
}

func checkQuota(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	svc service.Interface,
	authBackend consts.AuthBackend,
	accountID string,
) (int, error) {
	if authBackend != consts.AuthBackendAuth0 {
		return 0, nil
	}

	startDate := time.Now().AddDate(0, 0, -consts.FreeTierSendPeriodDays)

	count, err := svc.CountSendsByAccountDateRange(ctx, accountID, startDate, time.Now())
	if err != nil {
		countErr := fmt.Errorf("failed to check subscription limit: %w", err)
		HandleServiceError(w, r, countErr, "check quota")
		return 0, countErr
	}

	logging.AddLogAttr(r.Context(), slog.Int("sends_count", count))

	if count >= consts.MaxFreeTierSendsPerPeriod {
		quotaErr := fmt.Errorf("free tier limit exceeded: %w", apperrors.ErrInvalid)
		HandleServiceError(w, r, quotaErr, "check quota")
		return count, quotaErr
	}

	return count, nil
}

func checkDeviceEmail(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	svc service.Interface,
	accountID string,
) error {
	destEmail, _, err := svc.GetUserDeviceEmailAndAutoSend(ctx, accountID)
	if err != nil {
		emailErr := fmt.Errorf("failed to get user device email: %w", err)
		HandleServiceError(w, r, emailErr, "get device email")
		return emailErr
	}

	if destEmail == "" {
		return nil
	}

	isBouncing, err := svc.IsEmailBouncing(ctx, accountID, destEmail)
	if err != nil {
		bounceErr := fmt.Errorf("failed to check if email is bouncing: %w", err)
		HandleServiceError(w, r, bounceErr, "check bouncing email")
		return bounceErr
	}

	if isBouncing {
		return handleBouncingEmail(ctx, w, r, svc, accountID, destEmail)
	}

	return nil
}

func handleBouncingEmail(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	svc service.Interface,
	accountID string,
	destEmail string,
) error {
	bounceMsg := fmt.Sprintf("device email %s is blocked due to previous bounce", destEmail)

	profile, profileErr := svc.GetUserProfile(ctx, accountID)
	if profileErr == nil && profile != nil && profile.BouncedEmails != nil {
		bounceInfo, exists := profile.BouncedEmails[destEmail]
		if exists && bounceInfo.Error != "" {
			bounceMsg = fmt.Sprintf("%s: %s", bounceMsg, bounceInfo.Error)
		}
	}

	bounceErr := fmt.Errorf("%s: %w", bounceMsg, apperrors.ErrInvalid)
	HandleServiceError(w, r, bounceErr, "check bouncing email")
	return bounceErr
}
