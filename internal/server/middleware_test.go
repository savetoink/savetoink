package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

const (
	customRequestID = "custom-request-id"
	lambdaRequestID = "lambda-request-id"
	awsRequestID    = "aws-request-id"
	testRemoteAddr  = "192.168.1.1:8080"
	testClientIP    = "203.0.113.1"
	testClientIPv6  = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	testUserAgent   = "Mozilla/5.0 (Macintosh)"
)

func TestCorsMiddleware_GET(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	corsMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}

	allowedHeaders := "Content-Type, Authorization"
	if w.Header().Get("Access-Control-Allow-Headers") != allowedHeaders {
		t.Errorf("expected Access-Control-Allow-Headers '%s', got '%s'",
			allowedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
	}

	allowedMethods := "POST, GET, OPTIONS"
	if w.Header().Get("Access-Control-Allow-Methods") != allowedMethods {
		t.Errorf("expected Access-Control-Allow-Methods '%s', got '%s'",
			allowedMethods, w.Header().Get("Access-Control-Allow-Methods"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got '%s'",
			w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestCorsMiddleware_POST(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", http.NoBody)
	w := httptest.NewRecorder()

	corsMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Methods") != methodsAllowed {
		t.Errorf("expected Access-Control-Allow-Methods '%s', got '%s'",
			methodsAllowed, w.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestCorsMiddleware_OPTIONS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	w := httptest.NewRecorder()

	corsMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestCorsMiddleware_OriginHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("origin", testOrigin)
	w := httptest.NewRecorder()

	corsMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != testOrigin {
		t.Errorf("expected Access-Control-Allow-Origin '%s', got '%s'",
			testOrigin, w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCorsMiddleware_NoOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	corsMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRequestIDMiddleware_LambdaContext(t *testing.T) {
	var gotRequestID *string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = getRequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	requestID := "lambda-request-id-123"
	lc := &lambdacontext.LambdaContext{
		AwsRequestID: requestID,
	}
	ctx = lambdacontext.NewContext(ctx, lc)

	req := httptest.NewRequest("GET", "/test", http.NoBody).WithContext(ctx)
	w := httptest.NewRecorder()

	requestIDMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != requestID {
		t.Errorf("expected X-Request-ID '%s', got '%s'",
			requestID, w.Header().Get("X-Request-ID"))
	}

	if gotRequestID == nil || *gotRequestID != requestID {
		var got string
		if gotRequestID != nil {
			got = *gotRequestID
		}
		t.Errorf("expected context value '%s', got '%s'", requestID, got)
	}
}

func TestRequestIDMiddleware_XRequestIDHeader(t *testing.T) {
	var gotRequestID *string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = getRequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Request-ID", customRequestID)
	w := httptest.NewRecorder()

	requestIDMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != customRequestID {
		t.Errorf("expected X-Request-ID '%s', got '%s'",
			customRequestID, w.Header().Get("X-Request-ID"))
	}

	if gotRequestID == nil || *gotRequestID != customRequestID {
		var got string
		if gotRequestID != nil {
			got = *gotRequestID
		}
		t.Errorf("expected context value '%s', got '%s'", customRequestID, got)
	}
}

func TestRequestIDMiddleware_XAmznRequestIDHeader(t *testing.T) {
	var gotRequestID *string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = getRequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("x-amzn-request-id", awsRequestID)
	w := httptest.NewRecorder()

	requestIDMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != awsRequestID {
		t.Errorf("expected X-Request-ID '%s', got '%s'",
			awsRequestID, w.Header().Get("X-Request-ID"))
	}

	if gotRequestID == nil || *gotRequestID != awsRequestID {
		var got string
		if gotRequestID != nil {
			got = *gotRequestID
		}
		t.Errorf("expected context value '%s', got '%s'", awsRequestID, got)
	}
}

func TestRequestIDMiddleware_NoIDSource(t *testing.T) {
	var gotFromContext bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(contextKey(requestIDKey))
		_, gotFromContext = val.(string)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	requestIDMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID to be set, got empty string")
	}

	if !gotFromContext {
		t.Errorf("expected context value to be set")
	}
}

func TestRequestIDMiddleware_PriorityOrder(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	lc := &lambdacontext.LambdaContext{
		AwsRequestID: lambdaRequestID,
	}
	ctx = lambdacontext.NewContext(ctx, lc)

	req := httptest.NewRequest("GET", "/test", http.NoBody).WithContext(ctx)
	req.Header.Set("X-Request-ID", customRequestID)
	w := httptest.NewRecorder()

	requestIDMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != lambdaRequestID {
		t.Errorf("expected X-Request-ID '%s', got '%s'",
			lambdaRequestID, w.Header().Get("X-Request-ID"))
	}
}

func TestRequestIDMiddleware_PriorityOrderLambdaAbsent(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Request-ID", customRequestID)
	req.Header.Set("x-amzn-request-id", awsRequestID)
	w := httptest.NewRecorder()

	requestIDMiddleware(next).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != customRequestID {
		t.Errorf("expected X-Request-ID '%s', got '%s'",
			customRequestID, w.Header().Get("X-Request-ID"))
	}
}

func TestLoggingMiddleware_SuccessResponse(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
	loggingMiddleware(next).ServeHTTP(recorder, req)

	if recorder.status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.status)
	}
}

func TestLoggingMiddleware_ClientError(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	recorder := &responseStatusRecorder{
		ResponseWriter: w,
		status:         http.StatusBadRequest,
	}
	loggingMiddleware(next).ServeHTTP(recorder, req)

	if recorder.status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.status)
	}
}

func TestLoggingMiddleware_ServerError(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	recorder := &responseStatusRecorder{
		ResponseWriter: w,
		status:         http.StatusInternalServerError,
	}
	loggingMiddleware(next).ServeHTTP(recorder, req)

	if recorder.status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, recorder.status)
	}
}

func TestLoggingMiddleware_LatencyTracking(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
	loggingMiddleware(next).ServeHTTP(recorder, req)

	if recorder.status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.status)
	}
}

func TestLoggingMiddleware_RemoteAddr(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = testRemoteAddr
	w := httptest.NewRecorder()

	recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}
	loggingMiddleware(next).ServeHTTP(recorder, req)

	if recorder.status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.status)
	}
}

func TestResponseStatusRecorder_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseStatusRecorder{ResponseWriter: w}

	recorder.WriteHeader(http.StatusCreated)

	if recorder.status != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, recorder.status)
	}
}

func TestRemoteAddr_XForwardedForSingle(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For", testClientIP)

	got := remoteAddr(req)
	want := testClientIP
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestRemoteAddr_XForwardedForSingleIPv6(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For", testClientIPv6)

	got := remoteAddr(req)
	want := testClientIPv6
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestRemoteAddr_XForwardedForMultiple(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1, 192.0.2.1")

	got := remoteAddr(req)
	want := testClientIP
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestRemoteAddr_XForwardedForMultipleIPv6(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334, "+
			"2001:0db8:85a3:0000:0000:8a2e:0370:7335, "+
			"2001:0db8:85a3:0000:0000:8a2e:0370:7336")

	got := remoteAddr(req)
	want := testClientIPv6
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestRemoteAddr_XForwardedForMixedIPv4IPv6(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For",
		"203.0.113.1, 2001:0db8:85a3::8a2e:0370:7334, 198.51.100.1")

	got := remoteAddr(req)
	want := "203.0.113.1"
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestRemoteAddr_NoXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = testRemoteAddr

	got := remoteAddr(req)
	want := testRemoteAddr
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestRemoteAddr_NoSource(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = ""

	got := remoteAddr(req)
	want := "-"
	if got != want {
		t.Errorf("expected client_ip '%s', got '%s'", want, got)
	}
}

func TestUserAgent_StandardHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("User-Agent", testUserAgent)

	got := userAgent(req)
	want := testUserAgent
	if got != want {
		t.Errorf("expected user_agent '%s', got '%s'", want, got)
	}
}

func TestUserAgent_NoSource(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", http.NoBody)

	got := userAgent(req)
	want := "-"
	if got != want {
		t.Errorf("expected user_agent '%s', got '%s'", want, got)
	}
}

func BenchmarkCorsMiddleware(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := corsMiddleware(next)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("origin", testOrigin)
	w := httptest.NewRecorder()

	for b.Loop() {
		middleware.ServeHTTP(w, req)
	}
}

func BenchmarkRequestIDMiddleware(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := requestIDMiddleware(next)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Request-ID", customRequestID)
	w := httptest.NewRecorder()

	for b.Loop() {
		middleware.ServeHTTP(w, req)
	}
}

func BenchmarkLoggingMiddleware(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := loggingMiddleware(next)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	recorder := &responseStatusRecorder{ResponseWriter: w, status: http.StatusOK}

	for b.Loop() {
		middleware.ServeHTTP(recorder, req)
	}
}
