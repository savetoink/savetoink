package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorsMiddleware(t *testing.T) {
	t.Run("OPTIONS request returns 204 with CORS headers", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/test", http.NoBody)
		req.Header.Set("origin", "https://example.com")
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.False(t, nextCalled, "next handler should not be called for OPTIONS")
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "POST, GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("GET request passes through to next handler", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("origin", "https://example.com")
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called for GET")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("POST request passes through to next handler", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusCreated)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", http.NoBody)
		req.Header.Set("origin", "https://example.com")
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called for POST")
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("uses origin from request header", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("origin", "https://myapp.com")
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "https://myapp.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("defaults to wildcard when origin header is missing", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("all CORS headers are set", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("DELETE request passes through to next handler", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/test", http.NoBody)
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called for DELETE")
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("PUT request passes through to next handler", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/test", http.NoBody)
		w := httptest.NewRecorder()

		corsMiddleware(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called for PUT")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("uses lambda context AWS request ID as highest priority", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.Equal(t, "lambda-request-123", requestID)
			w.WriteHeader(http.StatusOK)
		})

		lc := &lambdacontext.LambdaContext{
			AwsRequestID: "lambda-request-123",
		}
		ctx := lambdacontext.NewContext(context.Background(), lc)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req = req.WithContext(ctx)
		req.Header.Set("X-Request-ID", "header-request-id")
		req.Header.Set("x-amzn-request-id", "amzn-request-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "lambda-request-123", w.Header().Get("X-Request-ID"))
	})

	t.Run("uses X-Request-ID header when no lambda context", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.Equal(t, "header-request-id", requestID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "header-request-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "header-request-id", w.Header().Get("X-Request-ID"))
	})

	t.Run("uses x-amzn-request-id header as fallback", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.NotNil(t, requestID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("x-amzn-request-id", "amzn-request-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "amzn-request-id", w.Header().Get("X-Request-ID"))
	})

	t.Run("generates request ID when no sources available", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.NotNil(t, requestID)
			assert.NotEmpty(t, requestID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		assert.NotContains(t, requestID, ".", "request ID should not contain dots")
	})

	t.Run("sets X-Request-ID header in response", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "test-request-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "test-request-id", w.Header().Get("X-Request-ID"))
	})

	t.Run("adds request ID to request context", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.NotNil(t, requestID, "request ID should be in context")
			assert.Equal(t, "test-id", requestID, "request ID should match header value")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "test-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)
	})

	t.Run("lambda context takes precedence over x-amzn-request-id", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.Equal(t, "lambda-id", requestID)
			w.WriteHeader(http.StatusOK)
		})

		lc := &lambdacontext.LambdaContext{
			AwsRequestID: "lambda-id",
		}
		ctx := lambdacontext.NewContext(context.Background(), lc)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req = req.WithContext(ctx)
		req.Header.Set("x-amzn-request-id", "amzn-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "lambda-id", w.Header().Get("X-Request-ID"))
	})

	t.Run("X-Request-ID takes precedence over x-amzn-request-id", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.Equal(t, "custom-id", requestID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "custom-id")
		req.Header.Set("x-amzn-request-id", "amzn-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "custom-id", w.Header().Get("X-Request-ID"))
	})

	t.Run("handles lambda context with empty AwsRequestID", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.NotNil(t, requestID)
			w.WriteHeader(http.StatusOK)
		})

		lc := &lambdacontext.LambdaContext{
			AwsRequestID: "",
		}
		ctx := lambdacontext.NewContext(context.Background(), lc)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req = req.WithContext(ctx)
		req.Header.Set("X-Request-ID", "fallback-id")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "fallback-id", w.Header().Get("X-Request-ID"))
	})

	t.Run("handles empty X-Request-ID header", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logging.RequestIDKey)
			assert.NotNil(t, requestID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		req.Header.Set("X-Request-ID", "")
		w := httptest.NewRecorder()

		requestIDMiddleware(next).ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})
}

func TestGenerateRequestID(t *testing.T) {
	t.Run("generates ID in correct format", func(t *testing.T) {
		requestID := generateRequestID()
		assert.NotEmpty(t, requestID)
		assert.Len(t, requestID, 18, "expected format YYYYMMDD-HHMMSSmmm (17 chars)")
	})

	t.Run("ID contains no dots", func(t *testing.T) {
		requestID := generateRequestID()
		assert.NotContains(t, requestID, ".", "dots should be removed")
	})

	t.Run("ID contains date and time separator", func(t *testing.T) {
		requestID := generateRequestID()
		assert.Contains(t, requestID, "-", "should contain date-time separator")
	})

	t.Run("generates unique IDs over time", func(t *testing.T) {
		id1 := generateRequestID()
		time.Sleep(1 * time.Millisecond)
		id2 := generateRequestID()
		assert.NotEqual(t, id1, id2, "IDs should be unique")
	})

	t.Run("generated ID format matches expected pattern", func(t *testing.T) {
		requestID := generateRequestID()
		parts := strings.Split(requestID, "-")
		assert.Len(t, parts, 2, "should have date and time parts")
		assert.Len(t, parts[0], 8, "date part should be 8 characters (YYYYMMDD)")
		assert.Len(t, parts[1], 9, "time part should be 9 characters (HHMMSSmmm)")
	})

	t.Run("ID contains no dots", func(t *testing.T) {
		requestID := generateRequestID()
		assert.NotContains(t, requestID, ".", "dots should be removed")
	})

	t.Run("ID contains date and time separator", func(t *testing.T) {
		requestID := generateRequestID()
		assert.Contains(t, requestID, "-", "should contain date-time separator")
	})

	t.Run("generates unique IDs over time", func(t *testing.T) {
		id1 := generateRequestID()
		time.Sleep(1 * time.Millisecond)
		id2 := generateRequestID()
		assert.NotEqual(t, id1, id2, "IDs should be unique")
	})

	t.Run("generated ID format matches expected pattern", func(t *testing.T) {
		requestID := generateRequestID()
		parts := strings.Split(requestID, "-")
		assert.Len(t, parts, 2, "should have date and time parts")
		assert.Len(t, parts[0], 8, "date part should be 8 characters (YYYYMMDD)")
		assert.Len(t, parts[1], 9, "time part should be 9 characters (HHMMSSmmm)")
	})
}

func TestJsonContentTypeMiddleware(t *testing.T) {
	t.Run("sets Content-Type header to application/json", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		jsonContentTypeMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("passes request to next handler", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		jsonContentTypeMiddleware(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called")
	})

	t.Run("next handler response is not modified", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		jsonContentTypeMiddleware(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, `{"status":"ok"}`, w.Body.String())
	})

	t.Run("works with all HTTP methods", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				req := httptest.NewRequestWithContext(context.Background(), method, "/test", http.NoBody)
				w := httptest.NewRecorder()

				jsonContentTypeMiddleware(next).ServeHTTP(w, req)

				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			})
		}
	})
}

func TestProcessorInfoMiddleware(t *testing.T) {
	t.Run("adds processed_by attribute when ProcessArticleLambda is set", func(t *testing.T) {
		cfg := &config.Config{
			ProcessArticleLambda: "article-processor-lambda",
		}

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
		logRecord := &logging.LogRecord{&record}
		ctx := context.WithValue(context.Background(), logging.LogRecordKey, logRecord)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", http.NoBody)
		w := httptest.NewRecorder()

		processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called")

		var attrs []slog.Attr
		record.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})

		require.Len(t, attrs, 1)
		assert.Equal(t, "processed_by", attrs[0].Key)
		assert.Equal(t, "article-processor-lambda", attrs[0].Value.String())
	})

	t.Run("does not add processed_by attribute when ProcessArticleLambda is empty", func(t *testing.T) {
		cfg := &config.Config{
			ProcessArticleLambda: "",
		}

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
		logRecord := &logging.LogRecord{&record}
		ctx := context.WithValue(context.Background(), logging.LogRecordKey, logRecord)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", http.NoBody)
		w := httptest.NewRecorder()

		processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

		assert.True(t, nextCalled, "next handler should be called")

		var attrs []slog.Attr
		record.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})

		assert.Len(t, attrs, 0, "no attributes should be added when ProcessArticleLambda is empty")
	})

	t.Run("always calls next handler regardless of config", func(t *testing.T) {
		testCases := []struct {
			name                 string
			processArticleLambda string
		}{
			{
				name:                 "with lambda configured",
				processArticleLambda: "article-processor-lambda",
			},
			{
				name:                 "without lambda configured",
				processArticleLambda: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					ProcessArticleLambda: tc.processArticleLambda,
				}

				nextCalled := false
				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					nextCalled = true
					w.WriteHeader(http.StatusOK)
				})

				req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/test", http.NoBody)
				w := httptest.NewRecorder()

				processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

				assert.True(t, nextCalled, "next handler should always be called")
			})
		}
	})

	t.Run("works with different HTTP methods", func(t *testing.T) {
		cfg := &config.Config{
			ProcessArticleLambda: "article-processor-lambda",
		}

		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
				logRecord := &logging.LogRecord{&record}
				ctx := context.WithValue(context.Background(), logging.LogRecordKey, logRecord)

				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				req := httptest.NewRequestWithContext(ctx, method, "/test", http.NoBody)
				w := httptest.NewRecorder()

				processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

				var attrs []slog.Attr
				record.Attrs(func(a slog.Attr) bool {
					attrs = append(attrs, a)
					return true
				})

				require.Len(t, attrs, 1)
				assert.Equal(t, "processed_by", attrs[0].Key)
				assert.Equal(t, "article-processor-lambda", attrs[0].Value.String())
			})
		}
	})

	t.Run("gracefully handles request without LogRecord in context", func(t *testing.T) {
		cfg := &config.Config{
			ProcessArticleLambda: "article-processor-lambda",
		}

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		ctx := context.Background()
		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", http.NoBody)
		w := httptest.NewRecorder()

		assert.NotPanics(t, func() {
			processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)
		})

		assert.True(t, nextCalled, "next handler should be called even without LogRecord")
	})

	t.Run("preserves existing attributes in LogRecord", func(t *testing.T) {
		cfg := &config.Config{
			ProcessArticleLambda: "article-processor-lambda",
		}

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
		logRecord := &logging.LogRecord{&record}
		ctx := context.WithValue(context.Background(), logging.LogRecordKey, logRecord)

		logging.AddString(ctx, "existing_key", "existing_value")

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", http.NoBody)
		w := httptest.NewRecorder()

		processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

		var attrs []slog.Attr
		record.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})

		require.Len(t, attrs, 2)

		attrMap := make(map[string]string)
		for _, attr := range attrs {
			attrMap[attr.Key] = attr.Value.String()
		}

		assert.Equal(t, "existing_value", attrMap["existing_key"])
		assert.Equal(t, "article-processor-lambda", attrMap["processed_by"])
	})

	t.Run("uses correct lambda function name from config", func(t *testing.T) {
		testCases := []struct {
			name       string
			lambdaName string
		}{
			{
				name:       "standard lambda name",
				lambdaName: "article-processor-lambda",
			},
			{
				name:       "lambda name with version",
				lambdaName: "article-processor-lambda:prod",
			},
			{
				name:       "lambda name with ARN",
				lambdaName: "arn:aws:lambda:us-east-1:123456789012:function:article-processor",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					ProcessArticleLambda: tc.lambdaName,
				}

				record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
				logRecord := &logging.LogRecord{&record}
				ctx := context.WithValue(context.Background(), logging.LogRecordKey, logRecord)

				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", http.NoBody)
				w := httptest.NewRecorder()

				processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

				var attrs []slog.Attr
				record.Attrs(func(a slog.Attr) bool {
					attrs = append(attrs, a)
					return true
				})

				require.Len(t, attrs, 1)
				assert.Equal(t, tc.lambdaName, attrs[0].Value.String())
			})
		}
	})

	t.Run("preserves request context and response", func(t *testing.T) {
		cfg := &config.Config{
			ProcessArticleLambda: "article-processor-lambda",
		}

		record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
		logRecord := &logging.LogRecord{&record}
		ctx := context.WithValue(context.Background(), logging.LogRecordKey, logRecord)

		customHeaderValue := "custom-request-value"
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", customHeaderValue)
			w.WriteHeader(http.StatusCreated)
		})

		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", http.NoBody)
		w := httptest.NewRecorder()

		processorInfoMiddleware(cfg)(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, customHeaderValue, w.Header().Get("X-Custom-Header"))
	})
}
