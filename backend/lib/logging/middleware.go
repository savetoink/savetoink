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
			start        = time.Now()
			accountID    = auth.GetAccountIDFromCtx(r.Context())
			requestID    = getRequestIDFromContext(r.Context())
			version      = consts.Version()
			requestError error
		)

		record := createLogRecord(r, accountID, requestID, version)

		ctx := context.WithValue(r.Context(), LogRecordKey, &LogRecord{Record: &record})
		ctx = context.WithValue(ctx, RequestErrorKey, &requestError)

		recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		finalizeLogRecord(ctx, &record, start, recorder.status)
	})
}

func getRequestIDFromContext(ctx context.Context) *string {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return nil
	}
	return &requestID
}
