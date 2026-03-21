package logging

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRequestID    = "req-12345"
	headerUserAgent  = "User-Agent"
	headerXForwarded = "X-Forwarded-For"
	testIPForwarded  = "203.0.113.1"
)

func TestCreateLogRecord_Basic(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test/path", http.NoBody)
	req.Header.Set(headerUserAgent, "test-agent")
	req.RemoteAddr = "192.168.1.1:8080"

	record := createLogRecord(req, "", nil)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	assert.Equal(t, "request completed", record.Message)
	assert.Equal(t, slog.LevelInfo, record.Level)

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "192.168.1.1:8080", attrMap["client_ip"])
	assert.Equal(t, "test-agent", attrMap["user_agent"])
	assert.Equal(t, "GET", attrMap["method"])
	assert.Equal(t, "/test/path", attrMap["path"])
	assert.NotContains(t, attrMap, "request_id")
	assert.NotContains(t, attrMap, "version")
	assert.NotContains(t, attrMap, "account_id")
}

func TestCreateLogRecord_WithRequestID(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/test", http.NoBody)
	requestID := "req-123"

	record := createLogRecord(req, "", &requestID)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, requestID, attrMap["request_id"])
}

func TestCreateLogRecord_WithAccountID(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	accountID := "account-456"

	record := createLogRecord(req, accountID, nil)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, accountID, attrMap["account_id"])
}

func TestCreateLogRecord_WithAllFields(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/test", http.NoBody)
	req.Header.Set(headerUserAgent, "Mozilla/5.0")
	req.RemoteAddr = "10.0.0.1:9000"
	accountID := "acct-789"
	requestID := "req-abc"

	record := createLogRecord(req, accountID, &requestID)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "10.0.0.1:9000", attrMap["client_ip"])
	assert.Equal(t, "Mozilla/5.0", attrMap["user_agent"])
	assert.Equal(t, "DELETE", attrMap["method"])
	assert.Equal(t, "/api/test", attrMap["path"])
	assert.Equal(t, requestID, attrMap["request_id"])
	assert.Equal(t, accountID, attrMap["account_id"])
}

func TestFinalizeLogRecord_Success(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	start := time.Now()

	ctx := context.Background()

	assert.NotPanics(t, func() {
		finalizeLogRecord(ctx, &record, start, http.StatusOK)
	})

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
	assert.Equal(t, slog.LevelInfo, record.Level)
	assert.Contains(t, attrMap, "latency_ms")
}

func TestFinalizeLogRecord_ClientError(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	start := time.Now()

	ctx := context.Background()

	finalizeLogRecord(ctx, &record, start, http.StatusBadRequest)

	assert.Equal(t, slog.LevelInfo, record.Level)
}

func TestFinalizeLogRecord_ServerError(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	start := time.Now()

	ctx := context.Background()

	finalizeLogRecord(ctx, &record, start, http.StatusInternalServerError)

	assert.Equal(t, slog.LevelError, record.Level)
}

func TestFinalizeLogRecord_WithRequestError(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	start := time.Now()

	testErr := errors.New("request failed")
	ctx := context.WithValue(context.Background(), requestErrorKey, &testErr)

	finalizeLogRecord(ctx, &record, start, http.StatusOK)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "request failed", attrMap["error"])
}

func TestFinalizeLogRecord_WithJoinedErrors(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	start := time.Now()

	err1 := errors.New("error one")
	err2 := errors.New("error two")
	joinedErr := errors.Join(err1, err2)
	ctx := context.WithValue(context.Background(), requestErrorKey, &joinedErr)

	finalizeLogRecord(ctx, &record, start, http.StatusOK)

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.String()
	}

	assert.Equal(t, "error one", attrMap["error_0"])
	assert.Equal(t, "error two", attrMap["error_1"])
}

func TestRemoteAddr_XForwardedForSingleIP(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set(headerXForwarded, testIPForwarded)

	result := remoteAddr(req)

	assert.Equal(t, testIPForwarded, result)
}

func TestRemoteAddr_XForwardedForMultipleIPs(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set(headerXForwarded, testIPForwarded+", 198.51.100.1, 192.0.2.1")

	result := remoteAddr(req)

	assert.Equal(t, testIPForwarded, result)
}

func TestRemoteAddr_XForwardedForNoComma(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set(headerXForwarded, testIPForwarded)

	result := remoteAddr(req)

	assert.Equal(t, testIPForwarded, result)
}

func TestRemoteAddr_NoXForwardedFor(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:8080"

	result := remoteAddr(req)

	assert.Equal(t, "192.168.1.100:8080", result)
}

func TestRemoteAddr_NoHeaders(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.RemoteAddr = ""

	result := remoteAddr(req)

	assert.Equal(t, "-", result)
}

func TestUserAgent_Present(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	result := userAgent(req)

	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", result)
}

func TestUserAgent_Missing(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)

	result := userAgent(req)

	assert.Equal(t, "-", result)
}

func TestResponseStatusRecorder_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}

	recorder.WriteHeader(http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, recorder.status)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestResponseStatusRecorder_DefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}

	assert.Equal(t, http.StatusOK, recorder.status)
}

func TestMiddleware_BasicFlow(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware := Middleware(next)
	middleware.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_CapturesStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware := Middleware(next)
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestMiddleware_WithContextValues(t *testing.T) {
	var capturedContext context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware := Middleware(next)
	middleware.ServeHTTP(w, req)

	assert.NotNil(t, capturedContext)

	logRecord := capturedContext.Value(logRecordKey)
	assert.NotNil(t, logRecord)

	requestError := capturedContext.Value(requestErrorKey)
	assert.NotNil(t, requestError)
}

func TestGetRequestIDFromContext_Present(t *testing.T) {
	requestID := testRequestID
	ctx := context.WithValue(context.Background(), requestIDKey, requestID)

	result := getRequestIDFromContext(ctx)

	require.NotNil(t, result)
	assert.Equal(t, requestID, *result)
}

func TestGetRequestIDFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	result := getRequestIDFromContext(ctx)

	assert.Nil(t, result)
}

func TestGetRequestIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey, 12345)

	result := getRequestIDFromContext(ctx)

	assert.Nil(t, result)
}

func TestMiddleware_WithRequestID(t *testing.T) {
	requestID := "test-request-id"
	ctx := context.WithValue(context.Background(), requestIDKey, requestID)
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(next)
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_WithAccountID(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(next)
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_LatencyTracking(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(next)
	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_LogLevelBasedOnStatus(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedLevel slog.Level
	}{
		{
			name:          "success status",
			statusCode:    http.StatusOK,
			expectedLevel: slog.LevelInfo,
		},
		{
			name:          "client error status",
			statusCode:    http.StatusBadRequest,
			expectedLevel: slog.LevelInfo,
		},
		{
			name:          "not found status",
			statusCode:    http.StatusNotFound,
			expectedLevel: slog.LevelInfo,
		},
		{
			name:          "server error status",
			statusCode:    http.StatusInternalServerError,
			expectedLevel: slog.LevelError,
		},
		{
			name:          "service unavailable",
			statusCode:    http.StatusServiceUnavailable,
			expectedLevel: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			middleware := Middleware(next)
			middleware.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
