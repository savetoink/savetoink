package logging

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRequestError_NoError(t *testing.T) {
	ctx := context.Background()
	err := GetRequestError(ctx)
	assert.Nil(t, err)
}

func TestGetRequestError_WithNilError(t *testing.T) {
	var nilErr error
	ctx := context.WithValue(context.Background(), RequestErrorKey, &nilErr)
	err := GetRequestError(ctx)
	assert.Nil(t, err)
}

func TestGetRequestError_WithError(t *testing.T) {
	testErr := errors.New("test error")
	ctx := context.WithValue(context.Background(), RequestErrorKey, &testErr)
	err := GetRequestError(ctx)
	assert.Equal(t, testErr, err)
}

func TestGetRequestError_JoinedErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	joinedErr := errors.Join(err1, err2)
	ctx := context.WithValue(context.Background(), RequestErrorKey, &joinedErr)
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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), RequestErrorKey, nilErr)

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
	ctx := context.WithValue(context.Background(), RequestErrorKey, &errPtr)

	testErr1 := errors.New("first error")
	AddRequestError(ctx, testErr1)

	retrievedErr := GetRequestError(ctx)
	assert.Equal(t, testErr1, retrievedErr)
}

func TestAddRequestError_MultipleErrors(t *testing.T) {
	var errPtr error
	ctx := context.WithValue(context.Background(), RequestErrorKey, &errPtr)

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
	ctx := context.WithValue(context.Background(), RequestErrorKey, &errPtr)

	newErr := errors.New("new error")
	AddRequestError(ctx, newErr)

	retrievedErr := GetRequestError(ctx)
	require.Error(t, retrievedErr)
	assert.Contains(t, retrievedErr.Error(), "initial")
	assert.Contains(t, retrievedErr.Error(), "new error")
}

func TestContextKeys(t *testing.T) {
	assert.Equal(t, contextKey("log_record"), LogRecordKey)
	assert.Equal(t, contextKey("request_error"), RequestErrorKey)
	assert.Equal(t, contextKey("request_id"), RequestIDKey)
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
	ctx := context.WithValue(context.Background(), LogRecordKey, logRecord)

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
	ctx := context.WithValue(context.Background(), RequestIDKey, requestID)

	result := GetRequestID(ctx)

	assert.Equal(t, requestID, result)
}

func TestGetRequestID_Missing(t *testing.T) {
	ctx := context.Background()

	result := GetRequestID(ctx)

	assert.Empty(t, result)
}

func TestGetRequestID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, 12345)

	result := GetRequestID(ctx)

	assert.Empty(t, result)
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
		slog.String("account_id", "acc-456"),
	}
	extraAttr := slog.String("status", "success")

	ctx := context.Background()

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs, extraAttr)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
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
	assert.NotContains(t, attrMap, "error")
}

func TestLogArticleProcessing_SingleError(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	testErr := errors.New("fetch failed")
	ctx := context.WithValue(context.Background(), RequestErrorKey, &testErr)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}
	extraAttr := slog.String("status", "failed")

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs, extraAttr)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, "failed", attrMap["status"])
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
	ctx := context.WithValue(context.Background(), RequestErrorKey, &joinedErr)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}
	extraAttr := slog.String("status", "failed")

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs, extraAttr)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelError, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, "failed", attrMap["status"])
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
	extraAttr := slog.String("status", "success")

	LogArticleProcessing(ctx, "article processing completed", nil, extraAttr)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "success", attrMap["status"])
}

func TestLogArticleProcessing_WithLogRecordAttrs(t *testing.T) {
	capture := &logCaptureHandler{records: make([]slog.Record, 0)}
	logger := slog.New(capture)
	defaultLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(defaultLogger)

	logRecord := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	ctx := context.WithValue(context.Background(), LogRecordKey, &LogRecord{&logRecord})

	AddString(ctx, "fetcher_type", "browserless")
	AddInt(ctx, "response_time_ms", 250)

	inheritedAttrs := []slog.Attr{
		slog.String("request_id", "req-123"),
	}
	extraAttr := slog.String("status", "success")

	LogArticleProcessing(ctx, "article processing completed", inheritedAttrs, extraAttr)

	require.Len(t, capture.records, 1)
	record := capture.records[0]

	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Equal(t, "article processing completed", record.Message)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "req-123", attrMap["request_id"])
	assert.Equal(t, "success", attrMap["status"])
	assert.Equal(t, "browserless", attrMap["fetcher_type"])
	assert.Equal(t, "250", attrMap["response_time_ms"])
}
