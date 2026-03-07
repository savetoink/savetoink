package logging

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/shaftoe/savetoink/backend/internal/logging"
	"github.com/shaftoe/savetoink/backend/internal/server/auth"
)

func createLogRecord(r *http.Request, accountID string, requestID, version *string) slog.Record {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	record.AddAttrs(
		slog.String("client_ip", remoteAddr(r)),
		slog.String("user_agent", userAgent(r)),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	if requestID != nil {
		record.AddAttrs(slog.String("request_id", *requestID))
	}
	if version != nil {
		record.AddAttrs(slog.String("version", *version))
	}
	if accountID != "" {
		record.AddAttrs(slog.String("account_id", accountID))
	}
	return record
}

func finalizeLogRecord(ctx context.Context, record *slog.Record, start time.Time, statusCode int) {
	authResult := auth.GetAuthError(ctx)
	if authResult != nil {
		record.AddAttrs(slog.String("auth_result", authResult.Error()))
	}

	requestError := logging.GetRequestError(ctx)
	if requestError != nil {
		if joinedErr, ok := requestError.(interface{ Unwrap() []error }); ok {
			for i, err := range joinedErr.Unwrap() {
				record.AddAttrs(slog.String(fmt.Sprintf("error_%d", i), err.Error()))
			}
		} else {
			record.AddAttrs(slog.String("error", requestError.Error()))
		}
	}

	if statusCode >= http.StatusInternalServerError {
		record.Level = slog.LevelError
	} else if statusCode >= http.StatusBadRequest {
		record.Level = slog.LevelInfo
	}

	record.AddAttrs(
		slog.Int("status", statusCode),
		slog.Int64("latency_ms", time.Since(start).Milliseconds()),
	)

	if err := slog.Default().Handler().Handle(ctx, *record); err != nil {
		slog.Error("failed to log request", "error", err)
	}
}

func remoteAddr(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip, _, found := strings.Cut(xff, ","); found {
			return ip
		}
		return xff
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "-"
}

func userAgent(r *http.Request) string {
	if ua := r.Header.Get("User-Agent"); ua != "" {
		return ua
	}
	return "-"
}
