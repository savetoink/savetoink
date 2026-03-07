package logging

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	attr := String("key", "value")
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, "value", attr.Value.String())
}

func TestInt(t *testing.T) {
	attr := Int("key", 42)
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, int64(42), attr.Value.Int64())
}

func TestInt64(t *testing.T) {
	attr := Int64("key", int64(1234567890))
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, int64(1234567890), attr.Value.Int64())
}

func TestBool(t *testing.T) {
	attr := Bool("key", true)
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, true, attr.Value.Bool())

	attr2 := Bool("key2", false)
	assert.Equal(t, "key2", attr2.Key)
	assert.Equal(t, false, attr2.Value.Bool())
}

func TestTime(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	attr := Time("key", now)
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, now, attr.Value.Time())
}

func TestDuration(t *testing.T) {
	duration := 5 * time.Second
	attr := Duration("key", duration)
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, duration, attr.Value.Duration())
}

func TestAny(t *testing.T) {
	attr := Any("key", "custom value")
	assert.Equal(t, "key", attr.Key)
	assert.Equal(t, "custom value", attr.Value.Any())

	attr2 := Any("key2", 123)
	assert.Equal(t, "key2", attr2.Key)
	assert.Equal(t, int64(123), attr2.Value.Any())
}

func TestError(t *testing.T) {
	err := errors.New("test error")
	attr := Error(err)
	assert.Equal(t, "error", attr.Key)
	assert.Equal(t, "test error", attr.Value.String())

	attr2 := Error(errors.New("another error"))
	assert.Equal(t, "error", attr2.Key)
	assert.Equal(t, "another error", attr2.Value.String())
}

func TestEmailSent(t *testing.T) {
	attrs := EmailSent("msg-123", "user@example.com")
	assert.Len(t, attrs, 2)
	assert.Equal(t, "email_message_id", attrs[0].Key)
	assert.Equal(t, "msg-123", attrs[0].Value.String())
	assert.Equal(t, "email_destination", attrs[1].Key)
	assert.Equal(t, "user@example.com", attrs[1].Value.String())
}

func TestArticleAction(t *testing.T) {
	attrs := ArticleAction("delete", "article-456")
	assert.Len(t, attrs, 2)
	assert.Equal(t, "action", attrs[0].Key)
	assert.Equal(t, "delete", attrs[0].Value.String())
	assert.Equal(t, "article_id", attrs[1].Key)
	assert.Equal(t, "article-456", attrs[1].Value.String())
}

func TestDeviceEmailChanged_WithOldEmail(t *testing.T) {
	attrs := DeviceEmailChanged("old@test.com", "new@test.com", true)
	assert.Len(t, attrs, 3)
	assert.Equal(t, "auto_send", attrs[0].Key)
	assert.Equal(t, true, attrs[0].Value.Bool())
	assert.Equal(t, "old_device_email", attrs[1].Key)
	assert.Equal(t, "old@test.com", attrs[1].Value.String())
	assert.Equal(t, "new_device_email", attrs[2].Key)
	assert.Equal(t, "new@test.com", attrs[2].Value.String())
}

func TestDeviceEmailChanged_WithoutOldEmail(t *testing.T) {
	attrs := DeviceEmailChanged("", "new@test.com", false)
	assert.Len(t, attrs, 2)
	assert.Equal(t, "auto_send", attrs[0].Key)
	assert.Equal(t, false, attrs[0].Value.Bool())
	assert.Equal(t, "new_device_email", attrs[1].Key)
	assert.Equal(t, "new@test.com", attrs[1].Value.String())
}

func TestBounceHandled(t *testing.T) {
	attrs := BounceHandled("bounced@test.com", "mailbox full", true)
	assert.Len(t, attrs, 3)
	assert.Equal(t, "bounced_email", attrs[0].Key)
	assert.Equal(t, "bounced@test.com", attrs[0].Value.String())
	assert.Equal(t, "bounce_error", attrs[1].Key)
	assert.Equal(t, "mailbox full", attrs[1].Value.String())
	assert.Equal(t, "hard_bounce", attrs[2].Key)
	assert.Equal(t, true, attrs[2].Value.Bool())
}

func TestWebhookProcessed(t *testing.T) {
	attrs := WebhookProcessed(100, 95, 5)
	assert.Len(t, attrs, 3)
	assert.Equal(t, "webhook_event_count", attrs[0].Key)
	assert.Equal(t, int64(100), attrs[0].Value.Int64())
	assert.Equal(t, "webhook_processed_count", attrs[1].Key)
	assert.Equal(t, int64(95), attrs[1].Value.Int64())
	assert.Equal(t, "webhook_failed_count", attrs[2].Key)
	assert.Equal(t, int64(5), attrs[2].Value.Int64())
}

func TestPagination(t *testing.T) {
	attrs := Pagination(2, 25, 100, true)
	assert.Len(t, attrs, 4)
	assert.Equal(t, "page", attrs[0].Key)
	assert.Equal(t, int64(2), attrs[0].Value.Int64())
	assert.Equal(t, "page_size", attrs[1].Key)
	assert.Equal(t, int64(25), attrs[1].Value.Int64())
	assert.Equal(t, "total", attrs[2].Key)
	assert.Equal(t, int64(100), attrs[2].Value.Int64())
	assert.Equal(t, "has_more", attrs[3].Key)
	assert.Equal(t, true, attrs[3].Value.Bool())
}

func TestAttrHelpers_Table(t *testing.T) {
	tests := []struct {
		name string
		fn   func() slog.Attr
		want slog.Attr
	}{
		{
			name: "String",
			fn:   func() slog.Attr { return String("k", "v") },
			want: slog.String("k", "v"),
		},
		{
			name: "Int",
			fn:   func() slog.Attr { return Int("k", 42) },
			want: slog.Int("k", 42),
		},
		{
			name: "Int64",
			fn:   func() slog.Attr { return Int64("k", 123) },
			want: slog.Int64("k", 123),
		},
		{
			name: "Bool",
			fn:   func() slog.Attr { return Bool("k", true) },
			want: slog.Bool("k", true),
		},
		{
			name: "Duration",
			fn:   func() slog.Attr { return Duration("k", time.Second) },
			want: slog.Duration("k", time.Second),
		},
		{
			name: "Any",
			fn:   func() slog.Attr { return Any("k", 3.14) },
			want: slog.Any("k", 3.14),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			assert.Equal(t, tt.want.Key, got.Key)
			assert.Equal(t, tt.want.Value.Kind(), got.Value.Kind())
		})
	}
}
