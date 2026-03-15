package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
)

const sentryTimeout = 5 * time.Second

// SetupLogging configures the global slog logger based on the provided configuration.
// It sets up a text handler and optionally integrates Sentry for error tracking.
func SetupLogging(cfg *config.Config) {
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	defaultHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(defaultHandler).With("version", *consts.Version()))

	if cfg.LoggingProvider == consts.LoggingBackendNone {
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Debug:            cfg.Debug,
		Environment:      cfg.SentryEnvironment,
		SampleRate:       cfg.SentrySampleRate,
		AttachStacktrace: true,
		EnableLogs:       true,
		Transport: &sentry.HTTPSyncTransport{
			Timeout: sentryTimeout,
		},
	})

	if err != nil {
		slog.With("version", *consts.Version()).
			Error("failed to initialize Sentry, fall back to default logger", "error", err)
		return
	}

	logLevels := []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	if cfg.Debug {
		logLevels = append(logLevels, slog.LevelDebug)
	}

	sentryHandler := sentryslog.Option{
		EventLevel: []slog.Level{slog.LevelWarn, slog.LevelError, sentryslog.LevelFatal},
		LogLevel:   logLevels,
	}.NewSentryHandler(context.Background())

	multiHandler := slog.NewMultiHandler(sentryHandler, defaultHandler)
	slog.SetDefault(slog.New(multiHandler))
}
