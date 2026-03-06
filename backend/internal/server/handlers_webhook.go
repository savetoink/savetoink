package server

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

func (h *handlers) handleMailjetWebhook(w http.ResponseWriter, r *http.Request) {
	// https://dev.mailjet.com/email/guides/webhooks/#best-practices
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mailjetWebhookResponse{
		Status: "ok",
	})

	body, readErr := readRequestBody(r)
	if readErr != nil {
		addRequestError(r.Context(), readErr)
		return
	}

	if verifyErr := h.verifyMailjetSecret(r); verifyErr != nil {
		addRequestError(r.Context(), verifyErr)
		return
	}

	var events []mailjetEvent
	if unmarshalErr := json.Unmarshal(body, &events); unmarshalErr != nil {
		addRequestError(r.Context(), unmarshalErr)
		return
	}

	processErrors := h.processBounceEvents(r, events)
	if processErrors != nil {
		addRequestError(r.Context(), processErrors)
		return
	}
}

func (h *handlers) verifyMailjetSecret(r *http.Request) error {
	if h.cfg.MailjetWebhookSecret == "" {
		return errors.New("mailjet webhook secret not configured")
	}

	secretQueryParam := r.URL.Query().Get("secret")
	if subtle.ConstantTimeCompare([]byte(secretQueryParam), []byte(h.cfg.MailjetWebhookSecret)) != 1 {
		return errors.New("invalid webhook secret")
	}

	return nil
}

func (h *handlers) processBounceEvents(r *http.Request, events []mailjetEvent) error {
	var err error

	for i := range events {
		event := &events[i]

		if event.Event != mailjetEventBounce {
			err = errors.Join(err, errors.New("unexpected event: "+event.Event))
			continue
		}

		if event.Email == "" {
			err = errors.Join(err, errors.New("empty email"))
			continue
		}

		errorMessage := h.extractErrorMessage(event)
		if handleErr := h.service.HandleBounce(r.Context(), event.Email, errorMessage); handleErr != nil {
			err = errors.Join(err, fmt.Errorf("handleErr: %w (email: %s)", handleErr, event.Email))
			continue
		}

		addLogAttr(r.Context(), slog.String("bounced_email", event.Email))
	}

	return err
}

func (h *handlers) extractErrorMessage(event *mailjetEvent) string {
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
