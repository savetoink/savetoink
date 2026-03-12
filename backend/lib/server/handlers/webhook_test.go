package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

type webhookMockService struct {
	handleBounceErr              error
	handleBounceCalled           bool
	handleBounceEmails           []string
	handleBounceErrorMessages    []string
	getAccountIDByDeviceEmail    string
	getAccountIDByDeviceEmailErr error
}

func (m *webhookMockService) Fetch(_ context.Context, _ *url.URL) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) ParseHTML(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) Clean(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) SendArticle(
	_ context.Context, _ string, _ io.ReadCloser, _ string,
) (*email.SendEmailResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) SendArticleByID(
	_ context.Context, _, _ string,
) (*servicetypes.SendArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) CreateArticle(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) UpdateArticle(_ context.Context, _ *model.Article) error {
	return nil
}

func (m *webhookMockService) GetArticle(_ context.Context, _, _ string) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) GetArticlesMetadata(
	_ context.Context, _ string, _, _ int, _ *bool,
) (*servicetypes.GetArticlesResult, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) DeleteArticle(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) DeleteAllArticles(_ context.Context, _ string) (*servicetypes.DeleteArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) GetDBError() error {
	return nil
}

func (m *webhookMockService) GetUserDeviceEmailAndAutoSend(
	_ context.Context, _ string,
) (deviceEmail string, autoSend bool, err error) {
	return "", false, nil
}

func (m *webhookMockService) SetUserDeviceEmailWithAutoSend(_ context.Context, _, _ string, _ bool) error {
	return errors.New("not implemented")
}

func (m *webhookMockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *webhookMockService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	return nil, errors.New("not implemented")
}

func (m *webhookMockService) SetUserEmail(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *webhookMockService) DeleteUserProfile(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *webhookMockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *webhookMockService) CountSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return 0, nil
}

func (m *webhookMockService) HandleBounce(_ context.Context, emailAddress, errorMessage string) error {
	m.handleBounceCalled = true
	m.handleBounceEmails = append(m.handleBounceEmails, emailAddress)
	m.handleBounceErrorMessages = append(m.handleBounceErrorMessages, errorMessage)
	return m.handleBounceErr
}

func (m *webhookMockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *webhookMockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	return m.getAccountIDByDeviceEmail, m.getAccountIDByDeviceEmailErr
}

func newWebhookTestHandlers(cfg *config.Config, svc *webhookMockService) *Handlers {
	return New(cfg, svc, http.DefaultClient, nil)
}

//nolint:gosec // test credentials, not real secrets
func newWebhookTestConfig() *config.Config {
	return &config.Config{
		APIKeySecret:         "test-api-key-secret",
		MailjetWebhookSecret: "test-webhook-secret",
		EmailProvider:        consts.EmailBackendMailjet,
		MailjetAPIKey:        "test-key",
		MailjetAPISecret:     "test-secret",
		SenderEmail:          "sender@example.com",
		Debug:                false,
	}
}

func newWebhookTestContext() context.Context {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	logRecord := &logging.LogRecord{Record: &record}
	ctx := context.Background()
	ctx = logging.WithLogRecordValue(ctx, logRecord)
	ctx = logging.WithRequestError(ctx)

	return ctx
}

func createMailjetEvent(
	emailAddress, eventType, errorMessage, comment string,
	hardBounce bool,
	timestamp int64,
) mailjetEvent {
	return mailjetEvent{
		Event:          eventType,
		Email:          emailAddress,
		Time:           timestamp,
		MessageID:      123456,
		CustomID:       "custom-id",
		Payload:        "payload",
		Error:          errorMessage,
		ErrorRelatedTo: "email",
		Comment:        comment,
		Blocked:        false,
		HardBounce:     hardBounce,
	}
}

func createMailjetEventJSON(events []mailjetEvent) []byte {
	eventsJSON, _ := json.Marshal(events)
	return eventsJSON
}

func createWebhookRequest(body []byte, secret string) *http.Request {
	path := "/v1/webhooks/mailjet"
	if secret != "" {
		path = path + "?secret=" + secret
	}
	req := httptest.NewRequestWithContext(context.Background(), "POST", path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createTestHTTPRequestWithErrorBody() *http.Request {
	path := "/v1/webhooks/mailjet?secret=test-webhook-secret"
	req := httptest.NewRequestWithContext(context.Background(), "POST", path, &errorReadCloser{})
	req.Header.Set("Content-Type", "application/json")
	return req
}

type errorReadCloser struct{}

func (e *errorReadCloser) Read([]byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (e *errorReadCloser) Close() error {
	return nil
}

const (
	logKeyAccountID       = "account_id"
	logKeyBounceError     = "bounce_error"
	logKeyBouncedEmail    = "bounced_email"
	logKeyHardBounce      = "hard_bounce"
	logKeyBounceTimestamp = "bounce_timestamp"
	logKeyProcessedCount  = "processed_count"
	logKeyFailedCount     = "failed_count"
)

func TestExtractErrorMessage(t *testing.T) {
	t.Run("returns event.Error when present", func(t *testing.T) {
		event := &mailjetEvent{
			Error:   "bounce error",
			Comment: "comment",
		}

		h := &Handlers{}
		result := h.extractErrorMessage(event)

		assert.Equal(t, "bounce error", result)
	})

	t.Run("returns event.Comment when Error is empty but Comment is present", func(t *testing.T) {
		event := &mailjetEvent{
			Error:   "",
			Comment: "bounce comment",
		}

		h := &Handlers{}
		result := h.extractErrorMessage(event)

		assert.Equal(t, "bounce comment", result)
	})

	t.Run("returns default 'bounce' when both Error and Comment are empty", func(t *testing.T) {
		event := &mailjetEvent{
			Error:   "",
			Comment: "",
		}

		h := &Handlers{}
		result := h.extractErrorMessage(event)

		assert.Equal(t, "bounce", result)
	})
}

func TestReadRequestBody(t *testing.T) {
	t.Run("successfully reads request body", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/test", bytes.NewReader(body))

		result, err := readRequestBody(req)

		require.NoError(t, err)
		assert.Equal(t, body, result)
	})

	t.Run("returns error when body read fails", func(t *testing.T) {
		req := &http.Request{
			Body: &errorReadCloser{},
		}

		result, err := readRequestBody(req)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to read request body")
	})
}

func TestVerifyMailjetSecret(t *testing.T) {
	t.Run("returns error when MailjetWebhookSecret is not configured", func(t *testing.T) {
		cfg := &config.Config{
			MailjetWebhookSecret: "",
		}
		h := New(cfg, &webhookMockService{}, http.DefaultClient, nil)

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test?secret=some-secret", http.NoBody)
		err := h.verifyMailjetSecret(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "mailjet webhook secret not configured")
	})

	t.Run("returns error when secret query param doesn't match", func(t *testing.T) {
		cfg := &config.Config{
			MailjetWebhookSecret: "correct-secret",
		}
		h := New(cfg, &webhookMockService{}, http.DefaultClient, nil)

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test?secret=wrong-secret", http.NoBody)
		err := h.verifyMailjetSecret(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid webhook secret")
	})

	t.Run("returns nil when secret is valid and matches", func(t *testing.T) {
		cfg := &config.Config{
			MailjetWebhookSecret: "correct-secret",
		}
		h := New(cfg, &webhookMockService{}, http.DefaultClient, nil)

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test?secret=correct-secret", http.NoBody)
		err := h.verifyMailjetSecret(req)

		require.NoError(t, err)
	})
}

func TestProcessBounceEvents(t *testing.T) {
	t.Run("successfully processes single valid bounce event", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "bounce error", "comment", true, 1234567890),
		}

		err := h.processBounceEvents(req, events)

		require.NoError(t, err)
		assert.True(t, svc.handleBounceCalled)
		assert.Len(t, svc.handleBounceEmails, 1)
		assert.Equal(t, "test@example.com", svc.handleBounceEmails[0])
		assert.Equal(t, "bounce error", svc.handleBounceErrorMessages[0])
	})

	t.Run("successfully processes multiple valid bounce events", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test1@example.com", "bounce", "error1", "", false, 1234567890),
			createMailjetEvent("test2@example.com", "bounce", "error2", "", true, 1234567891),
			createMailjetEvent("test3@example.com", "bounce", "error3", "", false, 1234567892),
		}

		err := h.processBounceEvents(req, events)

		require.NoError(t, err)
		assert.True(t, svc.handleBounceCalled)
		assert.Len(t, svc.handleBounceEmails, 3)
	})

	t.Run("handles events where HandleBounce service returns error", func(t *testing.T) {
		svc := &webhookMockService{
			handleBounceErr: errors.New("handle bounce failed"),
		}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "bounce error", "", true, 1234567890),
		}

		err := h.processBounceEvents(req, events)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "handle bounce failed")
		assert.Contains(t, err.Error(), "test@example.com")
	})

	t.Run("handles events with non-bounce event type", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "open", "", "", false, 1234567890),
		}

		err := h.processBounceEvents(req, events)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected event: open")
		assert.False(t, svc.handleBounceCalled)
	})

	t.Run("handles events with empty email field", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("", "bounce", "error", "", true, 1234567890),
		}

		err := h.processBounceEvents(req, events)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty email")
		assert.False(t, svc.handleBounceCalled)
	})

	t.Run("handles mixed valid and invalid events", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("valid@example.com", "bounce", "error", "", true, 1234567890),
			createMailjetEvent("", "bounce", "error", "", true, 1234567891),
			createMailjetEvent("another@example.com", "open", "", "", false, 1234567892),
		}

		err := h.processBounceEvents(req, events)

		require.Error(t, err)
		assert.True(t, svc.handleBounceCalled)
		assert.Len(t, svc.handleBounceEmails, 1)
		assert.Equal(t, "valid@example.com", svc.handleBounceEmails[0])
	})

	t.Run("logs bounced_email attribute", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyBouncedEmail && a.Value.String() == "test@example.com" {
				found = true
			}
			return !found
		})
		assert.True(t, found, "bounced_email attribute should be logged")
	})

	t.Run("logs account_id attribute when GetAccountIDByDeviceEmail succeeds", func(t *testing.T) {
		svc := &webhookMockService{
			getAccountIDByDeviceEmail: "account-123",
		}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyAccountID && a.Value.String() == "account-123" {
				found = true
			}
			return !found
		})
		assert.True(t, found, "account_id attribute should be logged")
	})

	t.Run("does not log account_id when GetAccountIDByDeviceEmail returns empty", func(t *testing.T) {
		svc := &webhookMockService{
			getAccountIDByDeviceEmail: "",
		}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyAccountID {
				found = true
			}
			return !found
		})
		assert.False(t, found, "account_id attribute should not be logged when empty")
	})

	t.Run("logs bounce_error attribute when errorMessage present", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "bounce error message", "", true, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyBounceError && a.Value.String() == "bounce error message" {
				found = true
			}
			return !found
		})
		assert.True(t, found, "bounce_error attribute should be logged")
	})

	t.Run("logs bounce_error with default value when both error and comment empty", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "", "", true, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyBounceError && a.Value.String() == mailjetEventBounce {
				found = true
			}
			return !found
		})
		assert.True(t, found, "bounce_error attribute should be logged with default value")
	})

	t.Run("logs hard_bounce attribute when HardBounce is true", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "", "", true, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyHardBounce && a.Value.Bool() {
				found = true
			}
			return !found
		})
		assert.True(t, found, "hard_bounce attribute should be logged")
	})

	t.Run("does not log hard_bounce when HardBounce is false", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "", "", false, 1234567890),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyHardBounce {
				found = true
			}
			return !found
		})
		assert.False(t, found, "hard_bounce attribute should not be logged when false")
	})

	t.Run("logs bounce_timestamp attribute when Time > 0", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		timestamp := int64(1234567890)
		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "", "", true, timestamp),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		expectedTime := time.Unix(timestamp, 0).UTC()
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyBounceTimestamp {
				found = true
				assert.Equal(t, expectedTime, a.Value.Time())
			}
			return !found
		})
		assert.True(t, found, "bounce_timestamp attribute should be logged")
	})

	t.Run("does not log bounce_timestamp when Time <= 0", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "", "", true, 0),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var found bool
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyBounceTimestamp {
				found = true
			}
			return !found
		})
		assert.False(t, found, "bounce_timestamp attribute should not be logged when Time <= 0")
	})

	t.Run("logs processed_count and failed_count correctly", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("valid1@example.com", "bounce", "error1", "", true, 1234567890),
			createMailjetEvent("valid2@example.com", "bounce", "error2", "", false, 1234567891),
			createMailjetEvent("valid3@example.com", "bounce", "error3", "", true, 1234567892),
			createMailjetEvent("", "bounce", "error", "", true, 1234567893),
			createMailjetEvent("test@example.com", "open", "", "", false, 1234567894),
		}

		_ = h.processBounceEvents(req, events)

		logRecord := logging.GetLogRecord(ctx)
		var processedCount, failedCount int
		logRecord.Attrs(func(a slog.Attr) bool {
			if a.Key == logKeyProcessedCount {
				processedCount = int(a.Value.Int64())
			}
			if a.Key == logKeyFailedCount {
				failedCount = int(a.Value.Int64())
			}
			return true
		})
		assert.Equal(t, 3, processedCount)
		assert.Equal(t, 2, failedCount)
	})

	t.Run("returns combined errors from multiple failed events", func(t *testing.T) {
		svc := &webhookMockService{
			handleBounceErr: errors.New("handle bounce failed"),
		}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{
			createMailjetEvent("", "bounce", "error", "", true, 1234567890),
			createMailjetEvent("", "bounce", "error", "", true, 1234567891),
		}

		err := h.processBounceEvents(req, events)

		require.Error(t, err)
		errStr := err.Error()
		assert.Contains(t, errStr, "empty email")
	})

	t.Run("handles all optional event fields populated", func(t *testing.T) {
		svc := &webhookMockService{
			getAccountIDByDeviceEmail: "account-123",
		}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		event := createMailjetEvent("test@example.com", "bounce", "error message", "comment", true, 1234567890)
		event.MessageID = 123456
		event.CustomID = "custom-123"
		event.Payload = "payload-data"
		event.ErrorRelatedTo = "email-related"
		event.Blocked = true

		events := []mailjetEvent{event}

		err := h.processBounceEvents(req, events)

		require.NoError(t, err)
		assert.True(t, svc.handleBounceCalled)
		assert.Equal(t, "error message", svc.handleBounceErrorMessages[0])
	})

	t.Run("handles minimal required event fields", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		event := mailjetEvent{
			Event: "bounce",
			Email: "test@example.com",
		}

		events := []mailjetEvent{event}

		err := h.processBounceEvents(req, events)

		require.NoError(t, err)
		assert.True(t, svc.handleBounceCalled)
		assert.Equal(t, "bounce", svc.handleBounceErrorMessages[0])
	})

	t.Run("handles empty events array", func(t *testing.T) {
		svc := &webhookMockService{}
		h := New(newWebhookTestConfig(), svc, http.DefaultClient, nil)

		ctx := newWebhookTestContext()
		req := httptest.NewRequestWithContext(ctx, "POST", "/test", http.NoBody)

		events := []mailjetEvent{}

		err := h.processBounceEvents(req, events)

		require.NoError(t, err)
		assert.False(t, svc.handleBounceCalled)
	})
}

func TestHandleMailjetWebhook(t *testing.T) {
	t.Run("returns OK 200 response with status ok", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("handles request body read error", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		req := createTestHTTPRequestWithErrorBody()
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("handles verifyMailjetSecret error with missing query param", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("handles verifyMailjetSecret error with invalid secret", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "wrong-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("handles invalid JSON in request body", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		body := []byte(`invalid json`)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("successfully processes valid bounce events array", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test1@example.com", "bounce", "error1", "", true, 1234567890),
			createMailjetEvent("test2@example.com", "bounce", "error2", "", false, 1234567891),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.True(t, svc.handleBounceCalled)
		assert.Len(t, svc.handleBounceEmails, 2)
	})

	t.Run("handles errors from processBounceEvents but still returns OK", func(t *testing.T) {
		svc := &webhookMockService{
			handleBounceErr: errors.New("handle bounce failed"),
		}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("verifies response body contains ok status", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test@example.com", "bounce", "error", "", true, 1234567890),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("handles empty events array", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.False(t, svc.handleBounceCalled)
	})

	t.Run("handles large events array", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := make([]mailjetEvent, 0, 100)
		for i := range 100 {
			emailAddress := strings.ToLower(string(rune('a'+i%26))) + "@example.com"
			events = append(events, createMailjetEvent(emailAddress, "bounce", "error", "", true, int64(1234567890+i)))
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, svc.handleBounceEmails, 100)
	})

	t.Run("always returns OK regardless of processing outcome (Mailjet best practice)", func(t *testing.T) {
		svc := &webhookMockService{
			handleBounceErr:              errors.New("service error"),
			getAccountIDByDeviceEmailErr: errors.New("db error"),
		}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("", "bounce", "", "", true, 0),
			createMailjetEvent("test@example.com", "open", "", "", false, 0),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp mailjetWebhookResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})

	t.Run("logs webhook_event_count", func(t *testing.T) {
		svc := &webhookMockService{}
		h := newWebhookTestHandlers(newWebhookTestConfig(), svc)

		events := []mailjetEvent{
			createMailjetEvent("test1@example.com", "bounce", "error1", "", true, 1234567890),
			createMailjetEvent("test2@example.com", "bounce", "error2", "", false, 1234567891),
			createMailjetEvent("test3@example.com", "bounce", "error3", "", true, 1234567892),
		}
		body := createMailjetEventJSON(events)
		req := createWebhookRequest(body, "test-webhook-secret")
		w := httptest.NewRecorder()

		h.HandleMailjetWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, svc.handleBounceEmails, 3)
	})
}
