package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/server/auth"
)

type responseStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseStatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

const (
	requestIDKey = "request_id"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := ""

		if lc, ok := lambdacontext.FromContext(r.Context()); ok {
			requestID = lc.AwsRequestID
		}

		if requestID == "" {
			requestID = r.Header.Get("X-Request-ID")
		}

		if requestID == "" {
			requestID = r.Header.Get("x-amzn-request-id")
		}

		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), contextKey(requestIDKey), requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start     = time.Now()
			accountID = auth.GetAccountID(r.Context())
			requestID = getRequestIDFromContext(r.Context())
			version   = consts.Version()
		)

		record := createLogRecord(r, accountID, requestID, version)

		var requestError error
		ctx := context.WithValue(r.Context(), logRecordKey, &logRecord{Record: &record})
		ctx = context.WithValue(ctx, requestErrorKey, &requestError)

		recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		finalizeLogRecord(ctx, &record, start, recorder.status)
	})
}

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

	requestError := getRequestError(ctx)
	if requestError != nil {
		record.Level = slog.LevelError
		record.AddAttrs(slog.String("error", requestError.Error()))
	}

	if statusCode >= http.StatusInternalServerError {
		record.Level = slog.LevelError
	}

	record.AddAttrs(
		slog.Int("status", statusCode),
		slog.Int64("latency_ms", time.Since(start).Milliseconds()),
	)

	if err := slog.Default().Handler().Handle(ctx, *record); err != nil {
		slog.Error("failed to log request", "error", err)
	}
}

func generateRequestID() string {
	return strings.ReplaceAll(time.Now().Format("20060102-150405.000"), ".", "")
}

func getRequestIDFromContext(ctx context.Context) *string {
	requestID, ok := ctx.Value(contextKey(requestIDKey)).(string)
	if !ok {
		return nil
	}
	return &requestID
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

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
