package logging

import (
	"log/slog"
	"time"
)

// String creates a string attribute.
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int creates an integer attribute.
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Int64 creates an int64 attribute.
func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

// Bool creates a boolean attribute.
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Time creates a time attribute.
func Time(key string, value time.Time) slog.Attr {
	return slog.Time(key, value)
}

// Duration creates a duration attribute.
func Duration(key string, value time.Duration) slog.Attr {
	return slog.Duration(key, value)
}

// Any creates an any attribute.
func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// Error creates an error attribute.
func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}

// EmailSent returns attributes for email send events.
func EmailSent(messageID, destEmail string) []slog.Attr {
	return []slog.Attr{
		String("email_message_id", messageID),
		String("email_destination", destEmail),
	}
}

// ArticleAction returns attributes for article action events.
func ArticleAction(action, articleID string) []slog.Attr {
	return []slog.Attr{
		String("action", action),
		String("article_id", articleID),
	}
}

// DeviceEmailChanged returns attributes for device email change events.
func DeviceEmailChanged(oldEmail, newEmail string, autoSend bool) []slog.Attr {
	attrs := []slog.Attr{
		Bool("auto_send", autoSend),
	}
	if oldEmail != "" {
		attrs = append(attrs, String("old_device_email", oldEmail))
	}
	attrs = append(attrs, String("new_device_email", newEmail))
	return attrs
}

// BounceHandled returns attributes for bounce handling events.
func BounceHandled(deviceEmail, errorMessage string, hardBounce bool) []slog.Attr {
	attrs := []slog.Attr{
		String("bounced_email", deviceEmail),
		String("bounce_error", errorMessage),
		Bool("hard_bounce", hardBounce),
	}
	return attrs
}

// WebhookProcessed returns attributes for webhook processing events.
func WebhookProcessed(eventCount, processedCount, failedCount int) []slog.Attr {
	return []slog.Attr{
		Int("webhook_event_count", eventCount),
		Int("webhook_processed_count", processedCount),
		Int("webhook_failed_count", failedCount),
	}
}

// Pagination returns attributes for pagination events.
func Pagination(page, pageSize, total int, hasMore bool) []slog.Attr {
	return []slog.Attr{
		Int("page", page),
		Int("page_size", pageSize),
		Int("total", total),
		Bool("has_more", hasMore),
	}
}

// ConvertSlogAttrsToMap converts a slice of slog.Attr to a slice of map[string]any.
func ConvertSlogAttrsToMap(attrs []slog.Attr) []map[string]any {
	if attrs == nil {
		return nil
	}
	result := make([]map[string]any, len(attrs))
	for i, attr := range attrs {
		result[i] = map[string]any{attr.Key: attr.Value.Any()}
	}
	return result
}
