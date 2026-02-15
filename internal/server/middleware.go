package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/internal/auth"
	"github.com/shaftoe/savetoink/internal/consts"
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
			level     = slog.LevelInfo
			record    = slog.NewRecord(time.Now(), level, "request completed", 0)
			requestID = getRequestIDFromContext(r.Context())
			userAgent = userAgent(r)
			clientIP  = remoteAddr(r)
			version   = consts.Version()
		)

		record.AddAttrs(
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("version", version),
		)
		if requestID != nil {
			record.AddAttrs(slog.String("request_id", *requestID))
		}
		if accountID != "" {
			record.AddAttrs(slog.String("account_id", accountID))
		}

		ctx := context.WithValue(r.Context(), logRecordKey, &logRecord{Record: &record})

		recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		latency := time.Since(start)
		statusCode := recorder.status
		authErr := auth.GetAuthError(r.Context())

		if authErr != nil {
			record.AddAttrs(slog.String("auth_error", authErr.Error()))
		}

		if statusCode >= http.StatusInternalServerError {
			record.Level = slog.LevelError
		}

		record.AddAttrs(
			slog.Int("status", statusCode),
			slog.Int64("latency_ms", latency.Milliseconds()),
		)

		if err := slog.Default().Handler().Handle(ctx, record); err != nil {
			slog.Error("failed to log request", "error", err)
		}
	})
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
