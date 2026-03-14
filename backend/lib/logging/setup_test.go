package logging

import (
	"log/slog"
	"os"
	"testing"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogging_TextOnly(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	cfg := &config.Config{
		Debug:             false,
		LoggingProvider:   consts.LoggingBackendNone,
		SentryDSN:         "",
		SentryEnvironment: "",
		SentrySampleRate:  0,
	}

	SetupLogging(cfg)

	testLogger := slog.Default()
	assert.NotNil(t, testLogger)

	assert.NotPanics(t, func() {
		testLogger.Info("test message")
		testLogger.Debug("debug message")
	})
}

func TestSetupLogging_DebugMode(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	cfg := &config.Config{
		Debug:             true,
		LoggingProvider:   consts.LoggingBackendNone,
		SentryDSN:         "",
		SentryEnvironment: "",
		SentrySampleRate:  0,
	}

	SetupLogging(cfg)

	testLogger := slog.Default()
	assert.NotNil(t, testLogger)

	assert.NotPanics(t, func() {
		testLogger.Debug("debug message")
	})
}

func TestSetupLogging_SentryIntegration(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	cfg := &config.Config{
		Debug:             false,
		LoggingProvider:   consts.LoggingBackendSentry,
		SentryDSN:         "https://test@sentry.io/123",
		SentryEnvironment: "test",
		SentrySampleRate:  1.0,
	}

	SetupLogging(cfg)

	testLogger := slog.Default()
	assert.NotNil(t, testLogger)

	assert.NotPanics(t, func() {
		testLogger.Info("info message")
		testLogger.Warn("warn message")
		testLogger.Error("error message")
	})
}

func TestSetupLogging_SentryWithDebug(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	cfg := &config.Config{
		Debug:             true,
		LoggingProvider:   consts.LoggingBackendSentry,
		SentryDSN:         "https://test@sentry.io/123",
		SentryEnvironment: "test",
		SentrySampleRate:  1.0,
	}

	SetupLogging(cfg)

	testLogger := slog.Default()
	assert.NotNil(t, testLogger)

	assert.NotPanics(t, func() {
		testLogger.Debug("debug message")
		testLogger.Info("info message")
	})
}

func TestSetupLogging_SentryFallback(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	cfg := &config.Config{
		Debug:             false,
		LoggingProvider:   consts.LoggingBackendSentry,
		SentryDSN:         "invalid-dsn",
		SentryEnvironment: "test",
		SentrySampleRate:  1.0,
	}

	SetupLogging(cfg)

	testLogger := slog.Default()
	assert.NotNil(t, testLogger)

	assert.NotPanics(t, func() {
		testLogger.Info("test message")
	})
}

func TestSetupLogging_CustomSampleRate(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	cfg := &config.Config{
		Debug:             false,
		LoggingProvider:   consts.LoggingBackendSentry,
		SentryDSN:         "https://test@sentry.io/123",
		SentryEnvironment: "test",
		SentrySampleRate:  0.5,
	}

	SetupLogging(cfg)

	testLogger := slog.Default()
	assert.NotNil(t, testLogger)

	assert.NotPanics(t, func() {
		testLogger.Info("test message")
	})
}

func TestSetupLogging_RespectsOutput(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := &config.Config{
		Debug:             false,
		LoggingProvider:   consts.LoggingBackendNone,
		SentryDSN:         "",
		SentryEnvironment: "",
		SentrySampleRate:  0,
	}

	SetupLogging(cfg)

	slog.Info("test message")

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}
	os.Stderr = oldStderr

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	assert.Greater(t, n, 0)
	assert.Contains(t, string(buf[:n]), "test message")
}
