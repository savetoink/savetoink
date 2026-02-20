package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

const (
	errorMsgMissingOrMalformedHeader = "missing or malformed auth header"
	errorMsgInvalidKey               = "invalid API key"
)

func TestNewMiddleware_SharedAPIKey(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:  consts.AuthBackendSharedAPIKey,
		APIKeySecret: "valid-key",
	}

	middleware := NewAccountIDMiddleware(cfg)

	if middleware == nil {
		t.Fatal("expected middleware to not be nil")
	}
}

func TestNewMiddleware_Auth0(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:   consts.AuthBackendAuth0,
		Auth0Domain:   "example.auth0.com",
		Auth0Audience: "test-audience",
	}

	middleware := NewAccountIDMiddleware(cfg)

	if middleware == nil {
		t.Fatal("expected middleware to not be nil")
	}
}

func TestSharedAPIKeyMiddleware_ValidAPIKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-key")
	w := httptest.NewRecorder()

	sharedAPIKeyMiddleware("valid-key")(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSharedAPIKeyMiddleware_MissingAPIKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middlewareChain := EnsureAutheticatedMiddleware(sharedAPIKeyMiddleware("valid-key")(next))
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSharedAPIKeyMiddleware_WrongAPIKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()

	middlewareChain := EnsureAutheticatedMiddleware(sharedAPIKeyMiddleware("valid-key")(next))
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSharedAPIKeyMiddleware_EmptyAPIKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	middlewareChain := EnsureAutheticatedMiddleware(sharedAPIKeyMiddleware("valid-key")(next))
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSharedAPIKeyMiddleware_InvalidBearerFormat(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "valid-key")
	w := httptest.NewRecorder()

	middlewareChain := EnsureAutheticatedMiddleware(sharedAPIKeyMiddleware("valid-key")(next))
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestNewMiddleware_DefaultBackend(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:  "",
		APIKeySecret: "valid-key",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewAccountIDMiddleware(cfg)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-key")
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetAuthError_NoError(t *testing.T) {
	ctx := context.Background()
	err := GetAuthError(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetAuthError_HasError(t *testing.T) {
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		err := GetAuthError(r.Context())
		if err == nil {
			t.Error("expected auth error in context")
		}
		if err.Error() != errorMsgMissingOrMalformedHeader {
			t.Errorf("expected error message '%s', got '%s'", errorMsgMissingOrMalformedHeader, err.Error())
		}
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	sharedAPIKeyMiddleware("valid-key")(next).ServeHTTP(w, req)
}

func TestSharedAPIKeyMiddleware_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectedMsg string
	}{
		{
			name:        "missing header",
			authHeader:  "",
			expectedMsg: errorMsgMissingOrMalformedHeader,
		},
		{
			name:        "invalid format",
			authHeader:  "valid-key",
			expectedMsg: errorMsgMissingOrMalformedHeader,
		},
		{
			name:        "empty token",
			authHeader:  "Bearer ",
			expectedMsg: errorMsgInvalidKey,
		},
		{
			name:        "wrong key",
			authHeader:  "Bearer wrong-key",
			expectedMsg: errorMsgInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMsg string
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				err := GetAuthError(r.Context())
				if err != nil {
					gotMsg = err.Error()
				}
			})

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			sharedAPIKeyMiddleware("valid-key")(next).ServeHTTP(w, req)

			if gotMsg != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, gotMsg)
			}
		})
	}
}

func TestSharedAPIKeyMiddleware_ErrorMessageInResponse(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectedMsg string
	}{
		{
			name:        "missing header",
			authHeader:  "",
			expectedMsg: errorMsgMissingOrMalformedHeader,
		},
		{
			name:        "invalid format",
			authHeader:  "valid-key",
			expectedMsg: errorMsgMissingOrMalformedHeader,
		},
		{
			name:        "empty token",
			authHeader:  "Bearer ",
			expectedMsg: errorMsgInvalidKey,
		},
		{
			name:        "wrong key",
			authHeader:  "Bearer wrong-key",
			expectedMsg: errorMsgInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middlewareChain := sharedAPIKeyMiddleware("valid-key")(EnsureAutheticatedMiddleware(next))
			middlewareChain.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}

			var resp model.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Error != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, resp.Error)
			}
		})
	}
}

func TestEnsureAutheticatedMiddleware_AuthErrorInContext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middlewareChain := sharedAPIKeyMiddleware("valid-key")(EnsureAutheticatedMiddleware(next))
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := errorMsgMissingOrMalformedHeader
	if resp.Error != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
	}
}
