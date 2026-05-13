package logging

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccountIDKey = "account_id"

const testRequestIDKey = "request_id"

func TestGetRequestError_NoError(t *testing.T) {
	ctx := context.Background()
	err := GetRequestError(ctx)
	assert.Nil(t, err)
}

func TestGetRequestError_WithNilError(t *testing.T) {
	var nilErr error
	ctx := context.WithValue(context.Background(), requestErrorKey, &nilErr)
	err := GetRequestError(ctx)
	assert.Nil(t, err)
}

func TestGetRequestError_WithError(t *testing.T) {
	testErr := errors.New("test error")
	ctx := context.WithValue(context.Background(), requestErrorKey, &testErr)
	err := GetRequestError(ctx)
	assert.Equal(t, testErr, err)
}

func TestGetRequestError_JoinedErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	joinedErr := errors.Join(err1, err2)
	ctx := context.WithValue(context.Background(), requestErrorKey, &joinedErr)
	err := GetRequestError(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error 1")
	assert.Contains(t, err.Error(), "error 2")
}

func TestAddLogAttr_NoRecord(t *testing.T) {
	ctx := context.Background()
	attr := String("key", "value")

	assert.NotPanics(t, func() {
		AddLogAttr(ctx, attr)
	})
}

func TestAddLogAttr_WithRecord(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddLogAttr(ctx, String("key1", "value1"))

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	assert.Len(t, attrs, 1)
	assert.Equal(t, "key1", attrs[0].Key)
	assert.Equal(t, "value1", attrs[0].Value.String())
}

func TestAddLogAttr_MultipleAttrs(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddLogAttr(ctx, String("key1", "value1"))
	AddLogAttr(ctx, Int("key2", 42))
	AddLogAttr(ctx, Bool("key3", true))

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 3)
	assert.Equal(t, "key1", attrs[0].Key)
	assert.Equal(t, "key2", attrs[1].Key)
	assert.Equal(t, "key3", attrs[2].Key)
}

func TestAddString(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddString(ctx, "test_key", "test_value")

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 1)
	assert.Equal(t, "test_key", attrs[0].Key)
	assert.Equal(t, "test_value", attrs[0].Value.String())
}

func TestAddInt(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddInt(ctx, "count", 100)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 1)
	assert.Equal(t, "count", attrs[0].Key)
	assert.Equal(t, int64(100), attrs[0].Value.Int64())
}

func TestAddBool(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddBool(ctx, "success", true)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 1)
	assert.Equal(t, "success", attrs[0].Key)
	assert.Equal(t, true, attrs[0].Value.Bool())
}

func TestAddTime(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddTime(ctx, "timestamp", now)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 1)
	assert.Equal(t, "timestamp", attrs[0].Key)
	assert.Equal(t, now, attrs[0].Value.Time())
}

func TestAddRequestError_NilError(t *testing.T) {
	var nilErr *error
	ctx := context.WithValue(context.Background(), requestErrorKey, nilErr)

	assert.NotPanics(t, func() {
		AddRequestError(ctx, nil)
	})

	err := GetRequestError(ctx)
	assert.Nil(t, err)
}

func TestAddRequestError_NoContextKey(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")

	assert.NotPanics(t, func() {
		AddRequestError(ctx, testErr)
	})
}

func TestAddRequestError_FirstError(t *testing.T) {
	var errPtr error
	ctx := context.WithValue(context.Background(), requestErrorKey, &errPtr)

	testErr1 := errors.New("first error")
	AddRequestError(ctx, testErr1)

	retrievedErr := GetRequestError(ctx)
	assert.Equal(t, testErr1, retrievedErr)
}

func TestAddRequestError_MultipleErrors(t *testing.T) {
	var errPtr error
	ctx := context.WithValue(context.Background(), requestErrorKey, &errPtr)

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	AddRequestError(ctx, err1)
	AddRequestError(ctx, err2)
	AddRequestError(ctx, err3)

	retrievedErr := GetRequestError(ctx)
	require.Error(t, retrievedErr)

	assert.Contains(t, retrievedErr.Error(), "error 1")
	assert.Contains(t, retrievedErr.Error(), "error 2")
	assert.Contains(t, retrievedErr.Error(), "error 3")
}

func TestAddRequestError_AfterExistingJoinedError(t *testing.T) {
	initialErr := errors.Join(errors.New("initial"), errors.New("error"))
	errPtr := initialErr
	ctx := context.WithValue(context.Background(), requestErrorKey, &errPtr)

	newErr := errors.New("new error")
	AddRequestError(ctx, newErr)

	retrievedErr := GetRequestError(ctx)
	require.Error(t, retrievedErr)
	assert.Contains(t, retrievedErr.Error(), "initial")
	assert.Contains(t, retrievedErr.Error(), "new error")
}

func TestContextKeys(t *testing.T) {
	// These keys are now private and only accessible within the logging package
	// This test verifies the internal key values are correct
	assert.Equal(t, contextKey("log_record"), logRecordKey)
	assert.Equal(t, contextKey("request_error"), requestErrorKey)
	assert.Equal(t, contextKey("request_id"), requestIDKey)
}

func TestLogRecord(t *testing.T) {
	now := time.Now()
	baseRecord := slog.NewRecord(now, slog.LevelInfo, "test message", 0)

	logRecord := &LogRecord{&baseRecord}

	assert.NotNil(t, logRecord)
	assert.NotNil(t, logRecord.Record)
	assert.Equal(t, now, logRecord.Time)
	assert.Equal(t, slog.LevelInfo, logRecord.Level)
	assert.Equal(t, "test message", logRecord.Message)
}

func TestAddLogAttr_Integration(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	logRecord := &LogRecord{&record}
	ctx := context.WithValue(context.Background(), logRecordKey, logRecord)

	AddString(ctx, "user_id", "user-123")
	AddInt(ctx, "items_count", 42)
	AddBool(ctx, "success", true)
	AddTime(ctx, "created_at", time.Now().UTC())

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 4)

	attrKeys := make(map[string]bool)
	for _, attr := range attrs {
		attrKeys[attr.Key] = true
	}

	assert.True(t, attrKeys["user_id"])
	assert.True(t, attrKeys["items_count"])
	assert.True(t, attrKeys["success"])
	assert.True(t, attrKeys["created_at"])
}

func TestGetRequestID_Present(t *testing.T) {
	requestID := "req-12345"
	ctx := context.WithValue(context.Background(), requestIDKey, requestID)

	result := GetRequestID(ctx)

	assert.Equal(t, requestID, result)
}

func TestGetRequestID_Missing(t *testing.T) {
	ctx := context.Background()

	result := GetRequestID(ctx)

	assert.Empty(t, result)
}

func TestGetRequestID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey, 12345)

	result := GetRequestID(ctx)

	assert.Empty(t, result)
}

func TestExtractInheritedLogAttrs(t *testing.T) {
	tests := []struct {
		name         string
		recordAttrs  []slog.Attr
		expectedKeys []string
		excludedKeys []string
	}{
		{
			name: "excludes client_ip, user_agent, path, method, url",
			recordAttrs: []slog.Attr{
				slog.String(excludeKeyClientIP, "192.168.1.1"),
				slog.String(excludeKeyUserAgent, "test-agent"),
				slog.String(excludeKeyPath, "/test"),
				slog.String(excludeKeyMethod, "GET"),
				slog.String("request_id", "req-123"),
				slog.String("version", "1.0.0"),
				slog.String(testAccountIDKey, "acc-456"),
				slog.String(excludeKeyURL, "https://example.com"),
				slog.String("article_id", "art-789"),
			},
			expectedKeys: []string{testRequestIDKey, "version", testAccountIDKey, "article_id"},
			excludedKeys: []string{excludeKeyClientIP, excludeKeyUserAgent, excludeKeyPath, excludeKeyMethod, excludeKeyURL},
		},
		{
			name: "empty record returns nil",
			recordAttrs: []slog.Attr{
				slog.String(excludeKeyClientIP, "192.168.1.1"),
				slog.String(excludeKeyUserAgent, "test-agent"),
				slog.String(excludeKeyPath, "/test"),
				slog.String(excludeKeyMethod, "GET"),
			},
			expectedKeys: []string{},
			excludedKeys: []string{excludeKeyClientIP, excludeKeyUserAgent, excludeKeyPath, excludeKeyMethod},
		},
		{
			name: "partial attrs with url excluded",
			recordAttrs: []slog.Attr{
				slog.String("request_id", "req-123"),
				slog.String(excludeKeyURL, "https://example.com"),
				slog.String(testAccountIDKey, "acc-456"),
			},
			expectedKeys: []string{testRequestIDKey, testAccountIDKey},
			excludedKeys: []string{excludeKeyURL},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			for _, attr := range tt.recordAttrs {
				record.AddAttrs(attr)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, logRecordKey, &LogRecord{Record: &record})

			attrs := ExtractInheritedLogAttrs(ctx)

			attrMap := make(map[string]string)
			for _, attr := range attrs {
				attrMap[attr.Key] = attr.Value.String()
			}

			for _, key := range tt.expectedKeys {
				assert.Contains(t, attrMap, key, "expected key %s to be present", key)
			}
			for _, key := range tt.excludedKeys {
				assert.NotContains(t, attrMap, key, "expected key %s to be excluded", key)
			}
		})
	}
}

func TestExtractInheritedLogAttrs_NoLogRecord(t *testing.T) {
	ctx := context.Background()
	attrs := ExtractInheritedLogAttrs(ctx)

	assert.Nil(t, attrs)
}

func TestExtractInheritedLogAttrs_NilLogRecord(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logRecordKey, (*LogRecord)(nil))

	attrs := ExtractInheritedLogAttrs(ctx)

	assert.Nil(t, attrs)
}

type logCaptureHandler struct {
	records []slog.Record
}

func (h *logCaptureHandler) Handle(_ context.Context, r slog.Record) error { //nolint:gocritic
	h.records = append(h.records, r)
	return nil
}

func (h *logCaptureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *logCaptureHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *logCaptureHandler) WithGroup(_ string) slog.Handler {
	return h
}

func TestLogArticleProcessing_Success(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
		slog.String(testAccountIDKey, "acc-456"),
	}

	ctx := context.Background()

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, "acc-456", attrMap[testAccountIDKey])
	assert.Equal(t, int64(http.StatusOK), attrMap["status"])
	assert.NotContains(t, attrMap, "error")
}

func TestLogArticleProcessing_SingleError(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	testErr := errors.New("fetch failed")
	ctx := context.WithValue(context.Background(), requestErrorKey, &testErr)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, int64(http.StatusInternalServerError), attrMap["status"])
	assert.Equal(t, "fetch failed", attrMap["error"])
}

func TestLogArticleProcessing_JoinedErrors(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	joinedErr := errors.Join(err1, err2)
	ctx := context.WithValue(context.Background(), requestErrorKey, &joinedErr)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, int64(http.StatusInternalServerError), attrMap["status"])
	assert.Equal(t, "error 1", attrMap["error_0"])
	assert.Equal(t, "error 2", attrMap["error_1"])
}

func TestLogArticleProcessing_NoInheritedAttrs(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	ctx := context.Background()

	LogArticleProcessing(ctx, "article processing completed", nil)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, int64(http.StatusOK), attrMap["status"])
}

func TestLogArticleProcessing_WithLogRecordAttrs(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	logRecord := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	ctx := context.WithValue(context.Background(), logRecordKey, &LogRecord{&logRecord})

	AddString(ctx, "fetcher_type", "browserless")
	AddInt(ctx, "response_time_ms", 250)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, int64(http.StatusOK), attrMap["status"])
	assert.Equal(t, "browserless", attrMap["fetcher_type"])
	assert.Equal(t, int64(250), attrMap["response_time_ms"])
}
