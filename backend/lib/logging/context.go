// Package logging provides structured logging helpers for request-scoped wide event logging.
package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const (
	// logRecordKey is the context key for the log record.
	logRecordKey contextKey = "log_record"
	// requestErrorKey is the context key for request errors.
	requestErrorKey contextKey = "request_error"
	// requestIDKey is the context key for request IDs.
	requestIDKey contextKey = "request_id"

	excludeKeyClientIP  = "client_ip"
	excludeKeyUserAgent = "user_agent"
	excludeKeyPath      = "path"
	excludeKeyMethod    = "method"
	excludeKeyURL       = "url"
)

// LogRecord wraps a slog.Record for use in request context.
type LogRecord struct {
	*slog.Record
}

func getLogRecord(ctx context.Context) *LogRecord {
	if record, ok := ctx.Value(logRecordKey).(*LogRecord); ok {
		return record
	}
	return nil
}

// GetRequestError retrieves any request error from context.
func GetRequestError(ctx context.Context) error {
	if errPtr, ok := ctx.Value(requestErrorKey).(*error); ok && errPtr != nil {
		return *errPtr
	}
	return nil
}

func addLogAttr(ctx context.Context, attr slog.Attr) {
	if record := getLogRecord(ctx); record != nil {
		record.AddAttrs(attr)
	}
}

// AddLogAttr adds a single attribute to the log record.
func AddLogAttr(ctx context.Context, attr slog.Attr) {
	addLogAttr(ctx, attr)
}

// AddString adds a string attribute to the log record.
func AddString(ctx context.Context, key, value string) {
	addLogAttr(ctx, slog.String(key, value))
}

// AddInt adds an integer attribute to the log record.
func AddInt(ctx context.Context, key string, value int) {
	addLogAttr(ctx, slog.Int(key, value))
}

// AddBool adds a boolean attribute to the log record.
func AddBool(ctx context.Context, key string, value bool) {
	addLogAttr(ctx, slog.Bool(key, value))
}

// AddTime adds a time attribute to the log record.
func AddTime(ctx context.Context, key string, value time.Time) {
	addLogAttr(ctx, slog.Time(key, value))
}

// AddRequestError adds an error to the request error accumulator.
func AddRequestError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if errPtr, ok := ctx.Value(requestErrorKey).(*error); ok && errPtr != nil {
		if *errPtr != nil {
			*errPtr = errors.Join(*errPtr, err)
		} else {
			*errPtr = err
		}
	}
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetLogRecord retrieves the log record from context.
// This is useful for testing and internal package use.
func GetLogRecord(ctx context.Context) *LogRecord {
	return getLogRecord(ctx)
}

// LogArticleProcessing logs article processing events.
func LogArticleProcessing(ctx context.Context, message string, inheritedAttrs []slog.Attr) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, message, 0)
	for _, attr := range inheritedAttrs {
		record.AddAttrs(attr)
	}

	if logRecord := getLogRecord(ctx); logRecord != nil {
		logRecord.Attrs(func(attr slog.Attr) bool {
			record.AddAttrs(attr)
			return true
		})
	}

	if requestError := GetRequestError(ctx); requestError != nil {
		record.AddAttrs(slog.Int("status", http.StatusInternalServerError))
		if joinedErr, ok := requestError.(interface{ Unwrap() []error }); ok {
			for i, err := range joinedErr.Unwrap() {
				record.AddAttrs(slog.String(fmt.Sprintf("error_%d", i), err.Error()))
			}
		} else {
			record.AddAttrs(slog.String("error", requestError.Error()))
		}
		record.Level = slog.LevelError
	} else {
		record.AddAttrs(slog.Int("status", http.StatusOK))
	}

	if err := slog.Default().Handler().Handle(ctx, record); err != nil {
		slog.Error("failed to log article processing", "error", err)
	}
}

// ExtractInheritedLogAttrs extracts attributes from the log record in the context,
// excluding HTTP metadata keys like client_ip, user_agent, path, method, and url.
func ExtractInheritedLogAttrs(ctx context.Context) []slog.Attr {
	logRecord, ok := ctx.Value(logRecordKey).(*LogRecord)
	if !ok || logRecord == nil {
		return nil
	}

	var attrs []slog.Attr
	excludeKeys := map[string]bool{
		excludeKeyClientIP:  true,
		excludeKeyUserAgent: true,
		excludeKeyPath:      true,
		excludeKeyMethod:    true,
		excludeKeyURL:       true,
	}
	logRecord.Attrs(func(a slog.Attr) bool {
		if !excludeKeys[a.Key] {
			attrs = append(attrs, a)
		}
		return true
	})
	return attrs
}

// WithLogRecord creates a context with a log record for structured logging.
func WithLogRecord(ctx context.Context) context.Context {
	return context.WithValue(ctx, logRecordKey, &LogRecord{Record: &slog.Record{}})
}

// WithLogRecordValue creates a context with the provided log record.
// This is useful for testing.
func WithLogRecordValue(ctx context.Context, record *LogRecord) context.Context {
	return context.WithValue(ctx, logRecordKey, record)
}

// WithRequestError creates a context with a request error accumulator.
func WithRequestError(ctx context.Context) context.Context {
	var err error
	return context.WithValue(ctx, requestErrorKey, &err)
}

// WithRequestID creates a context with a request ID.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}
