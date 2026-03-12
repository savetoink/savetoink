package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/logging"
)

const (
	mailjetEventBounce = "bounce"
)

type mailjetEvent struct {
	Event          string `json:"event"`
	Email          string `json:"email"`
	Time           int64  `json:"time"`
	MessageID      int64  `json:"MessageID"`
	CustomID       string `json:"CustomID"`
	Payload        string `json:"Payload"`
	Error          string `json:"error"`
	ErrorRelatedTo string `json:"error_related_to"`
	Comment        string `json:"comment"`
	Blocked        bool   `json:"blocked"`
	HardBounce     bool   `json:"hard_bounce"`
}

type mailjetWebhookResponse struct {
	Status string `json:"status"`
}

// HandleMailjetWebhook handles the Mailjet webhook endpoint.
func (h *Handlers) HandleMailjetWebhook(w http.ResponseWriter, r *http.Request) {
	// https://dev.mailjet.com/email/guides/webhooks/#best-practices
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mailjetWebhookResponse{
		Status: "ok",
	})

	body, readErr := readRequestBody(r)
	if readErr != nil {
		logging.AddRequestError(r.Context(), readErr)
		return
	}

	if verifyErr := h.verifyMailjetSecret(r); verifyErr != nil {
		logging.AddRequestError(r.Context(), verifyErr)
		return
	}

	var events []mailjetEvent
	if unmarshalErr := json.Unmarshal(body, &events); unmarshalErr != nil {
		logging.AddRequestError(r.Context(), unmarshalErr)
		return
	}

	logging.AddLogAttr(r.Context(), slog.Int("webhook_event_count", len(events)))

	processErrors := h.processBounceEvents(r, events)
	if processErrors != nil {
		logging.AddRequestError(r.Context(), processErrors)
		return
	}
}

func (h *Handlers) verifyMailjetSecret(r *http.Request) error {
	if h.cfg.MailjetWebhookSecret == "" {
		return errors.New("mailjet webhook secret not configured")
	}

	secretQueryParam := r.URL.Query().Get("secret")
	if subtle.ConstantTimeCompare([]byte(secretQueryParam), []byte(h.cfg.MailjetWebhookSecret)) != 1 {
		return errors.New("invalid webhook secret")
	}

	return nil
}

func (h *Handlers) processBounceEvents(r *http.Request, events []mailjetEvent) error {
	var processedCount int
	var failedCount int
	var err error

	for i := range events {
		event := &events[i]

		if event.Event != mailjetEventBounce {
			failedCount++
			err = errors.Join(err, errors.New("unexpected event: "+event.Event))
			continue
		}

		if event.Email == "" {
			failedCount++
			err = errors.Join(err, errors.New("empty email"))
			continue
		}

		errorMessage := h.extractErrorMessage(event)
		if handleErr := h.service.HandleBounce(r.Context(), event.Email, errorMessage); handleErr != nil {
			failedCount++
			err = errors.Join(err, fmt.Errorf("handleErr: %w (email: %s)", handleErr, event.Email))
			continue
		}

		processedCount++
		logging.AddLogAttr(r.Context(), slog.String("bounced_email", event.Email))

		accountID, accountErr := h.service.GetAccountIDByDeviceEmail(r.Context(), event.Email)
		if accountErr == nil && accountID != "" {
			logging.AddLogAttr(r.Context(), slog.String("account_id", accountID))
		}

		if errorMessage != "" {
			logging.AddLogAttr(r.Context(), slog.String("bounce_error", errorMessage))
		}

		if event.HardBounce {
			logging.AddLogAttr(r.Context(), slog.Bool("hard_bounce", true))
		}

		if event.Time > 0 {
			bounceTime := time.Unix(event.Time, 0).UTC()
			logging.AddLogAttr(r.Context(), slog.Time("bounce_timestamp", bounceTime))
		}
	}

	logging.AddLogAttr(r.Context(), slog.Int("processed_count", processedCount))
	logging.AddLogAttr(r.Context(), slog.Int("failed_count", failedCount))

	return err
}

func (h *Handlers) extractErrorMessage(event *mailjetEvent) string {
	errorMessage := event.Error
	if errorMessage == "" {
		errorMessage = event.Comment
	}
	if errorMessage == "" {
		errorMessage = "bounce"
	}
	return errorMessage
}

func readRequestBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	return body, nil
}
