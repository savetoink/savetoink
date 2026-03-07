// Package logging provides structured logging helpers for request-scoped wide event logging.
package logging

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type contextKey string

const (
	// LogRecordKey is the context key for the log record.
	LogRecordKey contextKey = "log_record"
	// RequestErrorKey is the context key for request errors.
	RequestErrorKey contextKey = "request_error"
	// RequestIDKey is the context key for request IDs.
	RequestIDKey contextKey = "request_id"
)

// LogRecord wraps a slog.Record for use in request context.
type LogRecord struct {
	*slog.Record
}

func getLogRecord(ctx context.Context) *LogRecord {
	if record, ok := ctx.Value(LogRecordKey).(*LogRecord); ok {
		return record
	}
	return nil
}

// GetRequestError retrieves any request error from context.
func GetRequestError(ctx context.Context) error {
	if errPtr, ok := ctx.Value(RequestErrorKey).(*error); ok && errPtr != nil {
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
	if errPtr, ok := ctx.Value(RequestErrorKey).(*error); ok && errPtr != nil {
		if *errPtr != nil {
			*errPtr = errors.Join(*errPtr, err)
		} else {
			*errPtr = err
		}
	}
}
