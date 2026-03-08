package server

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/stretchr/testify/assert"
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
