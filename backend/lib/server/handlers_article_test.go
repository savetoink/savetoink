package server

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractInheritedLogAttrs(t *testing.T) {
	tests := []struct {
		name         string
		recordAttrs  []slog.Attr
		expectedKeys []string
		excludedKeys []string
	}{
		{
			name: "excludes client_ip, user_agent, path, method, url",
			recordAttrs: []slog.Attr{
				slog.String("client_ip", "192.168.1.1"),
				slog.String("user_agent", "test-agent"),
				slog.String("path", "/test"),
				slog.String("method", "GET"),
				slog.String("request_id", "req-123"),
				slog.String("version", "1.0.0"),
				slog.String("account_id", "acc-456"),
				slog.String("url", "https://example.com"),
				slog.String("article_id", "art-789"),
			},
			expectedKeys: []string{"request_id", "version", "account_id", "article_id"},
			excludedKeys: []string{"client_ip", "user_agent", "path", "method", "url"},
		},
		{
			name: "empty record returns nil",
			recordAttrs: []slog.Attr{
				slog.String("client_ip", "192.168.1.1"),
				slog.String("user_agent", "test-agent"),
				slog.String("path", "/test"),
				slog.String("method", "GET"),
			},
			expectedKeys: []string{},
			excludedKeys: []string{"client_ip", "user_agent", "path", "method"},
		},
		{
			name: "partial attrs with url excluded",
			recordAttrs: []slog.Attr{
				slog.String("request_id", "req-123"),
				slog.String("url", "https://example.com"),
				slog.String("account_id", "acc-456"),
			},
			expectedKeys: []string{"request_id", "account_id"},
			excludedKeys: []string{"url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			for _, attr := range tt.recordAttrs {
				record.AddAttrs(attr)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &record})

			attrs := extractInheritedLogAttrs(ctx)

			attrMap := make(map[string]string)
			for _, attr := range attrs {
				attrMap[attr.Key] = attr.Value.String()
			}

			for _, key := range tt.expectedKeys {
				assert.Contains(t, attrMap, key, "expected key %s to be present", key)
			}
			for _, key := range tt.excludedKeys {
				assert.NotContains(t, attrMap, key, "expected key %s to be excluded", key)
			}
		})
	}
}

func TestExtractInheritedLogAttrs_NoLogRecord(t *testing.T) {
	ctx := context.Background()
	attrs := extractInheritedLogAttrs(ctx)

	assert.Nil(t, attrs)
}

func TestExtractInheritedLogAttrs_NilLogRecord(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, (*logging.LogRecord)(nil))

	attrs := extractInheritedLogAttrs(ctx)

	assert.Nil(t, attrs)
}

type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error { //nolint:gocritic
	h.records = append(h.records, r)
	return nil
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *captureHandler) WithGroup(_ string) slog.Handler {
	return h
}

func TestLogArticleProcessing(t *testing.T) {
	tests := []struct {
		name            string
		inheritedAttrs  []slog.Attr
		extraAttr       slog.Attr
		requestError    error
		expectedLevel   slog.Level
		expectedMessage string
		expectedKeys    []string
	}{
		{
			name: "success case with inherited attrs",
			inheritedAttrs: []slog.Attr{
				slog.String("request_id", "req-123"),
				slog.String("account_id", "acc-456"),
				slog.String("article_id", "art-789"),
			},
			extraAttr:       slog.String("status", "success"),
			requestError:    nil,
			expectedLevel:   slog.LevelInfo,
			expectedMessage: "article processing completed",
			expectedKeys:    []string{"request_id", "account_id", "article_id", "status"},
		},
		{
			name: "failure case with error",
			inheritedAttrs: []slog.Attr{
				slog.String("request_id", "req-123"),
				slog.String("account_id", "acc-456"),
			},
			extraAttr:       slog.String("status", "failed"),
			requestError:    errors.New("fetch error"),
			expectedLevel:   slog.LevelError,
			expectedMessage: "article processing completed",
			expectedKeys:    []string{"request_id", "account_id", "status", "error"},
		},
		{
			name:            "no inherited attrs",
			inheritedAttrs:  []slog.Attr{},
			extraAttr:       slog.String("status", "success"),
			requestError:    nil,
			expectedLevel:   slog.LevelInfo,
			expectedMessage: "article processing completed",
			expectedKeys:    []string{"status"},
		},
		{
			name:            "nil inherited attrs",
			inheritedAttrs:  nil,
			extraAttr:       slog.String("status", "success"),
			requestError:    nil,
			expectedLevel:   slog.LevelInfo,
			expectedMessage: "article processing completed",
			expectedKeys:    []string{"status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := &captureHandler{records: make([]slog.Record, 0)}
			logger := slog.New(capture)
			defaultLogger := slog.Default()
			slog.SetDefault(logger)
			defer slog.SetDefault(defaultLogger)

			var requestError *error
			ctx := context.Background()
			if tt.requestError != nil {
				err := tt.requestError
				requestError = &err
				ctx = context.WithValue(ctx, logging.RequestErrorKey, requestError)
			}

			h := &handlers{}
			h.logArticleProcessing(ctx, tt.inheritedAttrs, tt.extraAttr)

			require.Len(t, capture.records, 1, "expected one log record")

			record := capture.records[0]
			assert.Equal(t, tt.expectedLevel, record.Level)
			assert.Equal(t, tt.expectedMessage, record.Message)

			attrs := make(map[string]string)
			record.Attrs(func(a slog.Attr) bool {
				attrs[a.Key] = a.Value.String()
				return true
			})

			for _, key := range tt.expectedKeys {
				assert.Contains(t, attrs, key, "expected key %s to be present", key)
			}
		})
	}
}

func TestLogArticleProcessing_MultipleErrors(t *testing.T) {
	capture := &captureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	requestError := errors.Join(err1, err2)

	ctx := context.Background()
	errPtr := &requestError
	ctx = context.WithValue(ctx, logging.RequestErrorKey, errPtr)

	h := &handlers{}
	h.logArticleProcessing(ctx, []slog.Attr{
		slog.String("article_id", "art-123"),
	}, slog.String("status", "failed"))

	require.Len(t, capture.records, 1, "expected one log record")

	record := capture.records[0]
	assert.Equal(t, slog.LevelError, record.Level)

	attrs := make(map[string]string)
	record.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})

	assert.Contains(t, attrs, "article_id")
	assert.Contains(t, attrs, "status")
	assert.Contains(t, attrs, "error_0")
	assert.Contains(t, attrs, "error_1")
	assert.Contains(t, attrs["error_0"], "error 1")
	assert.Contains(t, attrs["error_1"], "error 2")
}
