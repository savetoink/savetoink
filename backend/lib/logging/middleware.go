// Package logging provides request-scoped wide event logging middleware.
package logging

import (
	"context"
	"net/http"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/consts"
)

type responseStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseStatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware creates a logging middleware that captures request lifecycle information
// and emits a single wide event log per request.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start     = time.Now()
			accountID = auth.GetAccountIDFromCtx(r.Context())
			requestID = getRequestIDFromContext(r.Context())
			version   = consts.Version()
		)

		record := createLogRecord(r, accountID, requestID, version)

		ctx := WithLogRecord(r.Context())
		// Initialize the log record with the created record
		var logRecord *LogRecord
		if lr := getLogRecord(ctx); lr != nil {
			*lr.Record = record
			logRecord = lr
		}
		ctx = WithRequestError(ctx)

		recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		// Use the log record from the context (which has been modified by handlers)
		// instead of the local copy
		if logRecord != nil {
			finalizeLogRecord(ctx, logRecord.Record, start, recorder.status)
		} else {
			// Fallback to local record if context record is missing (shouldn't happen)
			finalizeLogRecord(ctx, &record, start, recorder.status)
		}
	})
}

func getRequestIDFromContext(ctx context.Context) *string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return nil
	}
	return &requestID
}
