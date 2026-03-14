package logging

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shaftoe/savetoink/backend/lib/auth"
)

func createLogRecord(r *http.Request, accountID string, requestID *string) slog.Record {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "request completed", 0)
	record.AddAttrs(
		slog.String("client_ip", remoteAddr(r)),
		slog.String("user_agent", userAgent(r)),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	forwardedFor := forwardedFor(r)
	if forwardedFor != "" && forwardedFor != remoteAddr(r) {
		record.AddAttrs(slog.String("forwarded_for", forwardedFor))
	}
	if cloudFlareRay := r.Header.Get("CF-Ray"); cloudFlareRay != "" {
		record.AddAttrs(slog.String("cloud_flare_ray", cloudFlareRay))
	}
	if requestID != nil {
		record.AddAttrs(slog.String("request_id", *requestID))
	}
	if accountID != "" {
		record.AddAttrs(slog.String("account_id", accountID))
	}
	return record
}

func finalizeLogRecord(ctx context.Context, record *slog.Record, start time.Time, statusCode int) {
	authResult := auth.GetAuthErrorFromCtx(ctx)
	if authResult != nil {
		record.AddAttrs(slog.String("auth_result", authResult.Error()))
	}

	requestError := GetRequestError(ctx)
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

	if err := slog.Default().Handler().Handle(ctx, *record); err != nil { //nolint:gocritic
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

func forwardedFor(r *http.Request) string {
	return r.Header.Get("X-Forwarded-For")
}

func userAgent(r *http.Request) string {
	if ua := r.Header.Get("User-Agent"); ua != "" {
		return ua
	}
	return "-"
}

// GenerateRunID generates a unique ID for a task execution.
func GenerateRunID() string {
	return uuid.New().String()
}

// LogTaskExecution logs a task execution with timing and result information.
func LogTaskExecution(
	ctx context.Context,
	taskName, runID string,
	start time.Time,
	err error,
	output string,
	scheduledNext *time.Time,
) {
	record := slog.NewRecord(start, slog.LevelInfo, "task execution complete", 0)

	record.AddAttrs(
		String("task_name", taskName),
		String("run_id", runID),
		Duration("latency", time.Since(start)),
	)

	if err != nil {
		record.AddAttrs(Any("error", err))
		record.Level = slog.LevelError
	}

	if output != "" {
		record.AddAttrs(String("output", output))
	}

	if scheduledNext != nil {
		record.AddAttrs(Time("scheduled_next", *scheduledNext))
	}

	if logErr := slog.Default().Handler().Handle(ctx, record); logErr != nil { //nolint:gocritic
		slog.Error("failed to log task execution completion", "error", logErr)
	}
}

// LogSchedulerStarted logs when the background scheduler starts with enabled tasks.
func LogSchedulerStarted(ctx context.Context, tasks map[string]struct{}) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "Save to Ink background scheduler started", 0)
	keys := slices.Sorted(maps.Keys(tasks))
	record.AddAttrs(
		Any("tasks", keys),
	)

	if err := slog.Default().Handler().Handle(ctx, record); err != nil { //nolint:gocritic
		slog.Error("failed to log scheduler start", "error", err)
	}
}

// LogTaskFailed logs when a task execution fails.
func LogTaskFailed(ctx context.Context, taskName string, err error, scheduledNext *time.Time) {
	record := slog.NewRecord(time.Now(), slog.LevelError, "task execution failed", 0)
	record.AddAttrs(
		String("task_name", taskName),
		String("error", err.Error()),
	)

	if scheduledNext != nil {
		record.AddAttrs(Time("scheduled_next", *scheduledNext))
	}

	if logErr := slog.Default().Handler().Handle(ctx, record); logErr != nil { //nolint:gocritic
		slog.Error("failed to log task execution failure", "error", logErr)
	}
}
