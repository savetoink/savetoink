package lambda

import (
	"context"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/stretchr/testify/assert"
)

const testFunctionName = "test-function"

type testCaptureHandler struct {
	record *slog.Record
}

func (h *testCaptureHandler) Handle(_ context.Context, r slog.Record) error { //nolint:gocritic
	if h.record != nil {
		*h.record = r
	}
	return nil
}

func (h *testCaptureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *testCaptureHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *testCaptureHandler) WithGroup(_ string) slog.Handler {
	return h
}

func TestNewProcessor(t *testing.T) {
	awsCfg := aws.Config{
		Region: "us-east-1",
	}
	processor := NewProcessor(testFunctionName, &awsCfg)

	assert.NotNil(t, processor)
	assert.Equal(t, testFunctionName, processor.functionName)
	assert.NotNil(t, processor.lambdaClient)
}

func TestLogProcessingStarted(t *testing.T) {
	var capturedRecord slog.Record

	captureHandler := &testCaptureHandler{record: &capturedRecord}
	logger := slog.New(captureHandler)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	processor := NewProcessor(testFunctionName, &aws.Config{})

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &capturedRecord})

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
		slog.String("account_id", "acc-456"),
	}

	processor.logProcessingStarted(ctx, inheritedAttrs, "success")

	assert.Equal(t, "article processing started with "+testFunctionName, capturedRecord.Message)

	var attrs []slog.Attr
	capturedRecord.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, "acc-456", attrMap["account_id"])
	assert.Equal(t, "success", attrMap["status"])
}

func TestLogProcessingStarted_Failure(t *testing.T) {
	var capturedRecord slog.Record

	captureHandler := &testCaptureHandler{record: &capturedRecord}
	logger := slog.New(captureHandler)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	processor := NewProcessor(testFunctionName, &aws.Config{})

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &capturedRecord})

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}

	processor.logProcessingStarted(ctx, inheritedAttrs, "failure")

	assert.Equal(t, "article processing started with "+testFunctionName, capturedRecord.Message)

	var attrs []slog.Attr
	capturedRecord.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, "failure", attrMap["status"])
}

func TestLogProcessingStarted_EmptyAttrs(t *testing.T) {
	var capturedRecord slog.Record

	captureHandler := &testCaptureHandler{record: &capturedRecord}
	logger := slog.New(captureHandler)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	processor := NewProcessor(testFunctionName, &aws.Config{})

	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.LogRecordKey, &logging.LogRecord{Record: &capturedRecord})

	inheritedAttrs := []slog.Attr{}

	processor.logProcessingStarted(ctx, inheritedAttrs, "success")

	assert.Equal(t, "article processing started with "+testFunctionName, capturedRecord.Message)

	var attrs []slog.Attr
	capturedRecord.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "success", attrMap["status"])
}
