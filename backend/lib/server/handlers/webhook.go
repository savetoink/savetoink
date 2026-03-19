package handlers

import (
	"context"
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
	processor := bounceEventProcessor{handlers: h, ctx: r.Context()}
	return processor.Process(events)
}

type bounceEventProcessor struct {
	handlers *Handlers
	ctx      context.Context
}

func (p *bounceEventProcessor) Process(events []mailjetEvent) error {
	var processedCount int
	var failedCount int
	var err error

	for i := range events {
		event := &events[i]

		eventErr := p.processEvent(event)
		if eventErr != nil {
			failedCount++
			err = errors.Join(err, eventErr)
			continue
		}

		processedCount++
		p.logEventDetails(event)
	}

	logging.AddLogAttr(p.ctx, slog.Int("processed_count", processedCount))
	logging.AddLogAttr(p.ctx, slog.Int("failed_count", failedCount))

	return err
}

func (p *bounceEventProcessor) processEvent(event *mailjetEvent) error {
	if event.Event != mailjetEventBounce {
		return errors.New("unexpected event: " + event.Event)
	}

	if event.Email == "" {
		return errors.New("empty email")
	}

	errorMessage := p.handlers.extractErrorMessage(event)
	if handleErr := p.handlers.service.HandleBounce(p.ctx, event.Email, errorMessage); handleErr != nil {
		return fmt.Errorf("handleErr: %w (email: %s)", handleErr, event.Email)
	}

	return nil
}

func (p *bounceEventProcessor) logEventDetails(event *mailjetEvent) {
	logging.AddLogAttr(p.ctx, slog.String("bounced_email", event.Email))

	accountID, accountErr := p.handlers.service.GetAccountIDByDeviceEmail(p.ctx, event.Email)
	if accountErr == nil && accountID != "" {
		logging.AddLogAttr(p.ctx, slog.String("account_id", accountID))
	}

	errorMessage := p.handlers.extractErrorMessage(event)
	if errorMessage != "" {
		logging.AddLogAttr(p.ctx, slog.String("bounce_error", errorMessage))
	}

	if event.HardBounce {
		logging.AddLogAttr(p.ctx, slog.Bool("hard_bounce", true))
	}

	if event.Time > 0 {
		bounceTime := time.Unix(event.Time, 0).UTC()
		logging.AddLogAttr(p.ctx, slog.Time("bounce_timestamp", bounceTime))
	}
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
