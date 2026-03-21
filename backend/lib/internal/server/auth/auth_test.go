package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/auth"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

type MockService struct {
	deviceEmail               string
	autoSend                  bool
	emailErr                  error
	isBouncing                bool
	bouncingErr               error
	sendCount                 int
	sendCountErr              error
	userProfile               *model.UserProfile
	userProfileErr            error
	accountIDByDeviceEmail    string
	accountIDByDeviceEmailErr error
}

func (m *MockService) Fetch(_ context.Context, _ *url.URL) (*content.FetchedContent, error) {
	panic("not implemented")
}

func (m *MockService) ParseHTML(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
	panic("not implemented")
}

func (m *MockService) Clean(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
	panic("not implemented")
}

func (m *MockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	panic("not implemented")
}

func (m *MockService) SendArticle(
	_ context.Context, _ string, _ io.ReadCloser, _ string) (*email.SendEmailResponse, error) {
	panic("not implemented")
}

func (m *MockService) SendArticleByID(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
	panic("not implemented")
}

func (m *MockService) CreateArticle(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
	panic("not implemented")
}

func (m *MockService) UpdateArticle(_ context.Context, _ *model.Article) error {
	return nil
}

func (m *MockService) GetArticle(_ context.Context, _, _ string) (*model.Article, error) {
	panic("not implemented")
}

func (m *MockService) GetArticlesMetadata(
	_ context.Context,
	_ string,
	_, _ int,
	_ *bool,
) (*servicetypes.GetArticlesResult, error) {
	panic("not implemented")
}

func (m *MockService) DeleteArticle(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
	panic("not implemented")
}

func (m *MockService) GetDBError() error {
	panic("not implemented")
}

func (m *MockService) GetUserDeviceEmailAndAutoSend(
	_ context.Context, _ string,
) (deviceEmail string, autoSend bool, err error) {
	if m.emailErr != nil {
		return "", false, m.emailErr
	}
	return m.deviceEmail, m.autoSend, nil
}

func (m *MockService) SetUserDeviceEmailWithAutoSend(_ context.Context, _, _ string, _ bool) error {
	panic("not implemented")
}

func (m *MockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	panic("not implemented")
}

func (m *MockService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	if m.userProfileErr != nil {
		return nil, m.userProfileErr
	}
	return m.userProfile, nil
}

func (m *MockService) SetUserEmail(_ context.Context, _, _ string) error {
	panic("not implemented")
}

func (m *MockService) DeleteUserProfile(_ context.Context, _ string) error {
	panic("not implemented")
}

func (m *MockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	panic("not implemented")
}

func (m *MockService) CountSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	if m.sendCountErr != nil {
		return 0, m.sendCountErr
	}
	return m.sendCount, nil
}

func (m *MockService) HandleBounce(_ context.Context, _, _ string) error {
	panic("not implemented")
}

func (m *MockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	if m.bouncingErr != nil {
		return false, m.bouncingErr
	}
	return m.isBouncing, nil
}

func (m *MockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	if m.accountIDByDeviceEmailErr != nil {
		return "", m.accountIDByDeviceEmailErr
	}
	return m.accountIDByDeviceEmail, nil
}

func TestGetAccountID(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected string
	}{
		{
			name:     "account ID in context",
			setupCtx: func() context.Context { return auth.AddAccountIDToCtx(context.Background(), "test-account") },
			expected: "test-account",
		},
		{
			name:     "no account ID in context",
			setupCtx: context.Background,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := auth.GetAccountIDFromCtx(ctx)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func setupMockJWKSServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jwksResponse := `{
			"keys": [{
				"kty": "RSA",
				"kid": "test-key-id",
				"n": "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1` +
			`L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-` +
			`65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08q` +
			`NLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awap` +
			`JzKnqDKgw",
				"e": "AQAB",
				"alg": "RS256"
			}]
		}`
		if _, err := w.Write([]byte(jwksResponse)); err != nil {
			panic(fmt.Sprintf("failed to write JWKS response: %v", err))
		}
	}))
}

func TestAuth0Middleware(t *testing.T) {
	t.Run("missing auth header continues", func(t *testing.T) {
		middleware := auth0Middleware("test.auth0.com", "test-audience")

		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid JWT token adds error to context", func(t *testing.T) {
		jwksServer := setupMockJWKSServer()
		defer jwksServer.Close()

		middleware := auth0Middleware(jwksServer.URL[7:], "test-audience")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InRlc3Qta2V5LWlkIiwidHlwIjoiUlNBIiwiYWxnIjoiUlMyNTYifQ." +
			"eyJzdWIiOiJ0ZXN0LXVzZXItMTIzIiwiaXNzIjoiaHR0cHM6Ly90ZXN0LmF1dGgwLmNvbS8iLCJhdWQiOiJ0" +
			"ZXN0LWF1ZGllbmNlIiwiZXhwIjoxOTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDB9.invalid_signature"
		req.Header.Set(authHeader, authHeaderPrefix+token)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		err := auth.GetAuthErrorFromCtx(capturedContext)
		if err == nil {
			t.Error("expected auth error in context")
		}
	})
}

func TestHandleAuthError(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
	}{
		{
			name:     "error added to context",
			errorMsg: "test error message",
		},
		{
			name:     "empty error message",
			errorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedContext context.Context
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			handleAuthError(context.Background(), next, w, req, tt.errorMsg)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			err := auth.GetAuthErrorFromCtx(capturedContext)
			if err == nil && tt.errorMsg != "" {
				t.Error("expected auth error in context")
			}
			if err != nil && err.Error() != tt.errorMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestEnsureAuthenticatedMiddleware_WithAuthError(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
	}{
		{
			name: "auth error in context",
			setupCtx: func() context.Context {
				return auth.AddAuthErrorToCtx(context.Background(), "authentication failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			ctx := tt.setupCtx()
			req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			middlewareChain := EnsureAutheticatedMiddleware(next)
			middlewareChain.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}

			var resp model.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			expectedMsg := "authentication failed"
			if resp.Error != expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
			}
		})
	}
}

func TestEnsureAuthenticatedMiddleware_WithAccountID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := auth.AddAccountIDToCtx(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middlewareChain := EnsureAutheticatedMiddleware(next)
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestEnsureAuthenticatedMiddleware_NoAccountID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middlewareChain := EnsureAutheticatedMiddleware(next)
	middlewareChain.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := "unauthorized"
	if resp.Error != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestAddAuthErrorToCtx(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
	}{
		{
			name:     "add error message",
			errorMsg: "invalid credentials",
		},
		{
			name:     "empty error message",
			errorMsg: "",
		},
		{
			name:     "long error message",
			errorMsg: strings.Repeat("error ", 100),
		},
		{
			name:     "error with special characters",
			errorMsg: "error: authentication failed (code: 401)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = auth.AddAuthErrorToCtx(ctx, tt.errorMsg)

			err := auth.GetAuthErrorFromCtx(ctx)
			require.Error(t, err)
			assert.Equal(t, tt.errorMsg, err.Error())

			accountID := auth.GetAccountIDFromCtx(ctx)
			assert.Empty(t, accountID)
		})
	}
}

func TestAddAuthErrorToCtx_PreservesExistingAccountID(t *testing.T) {
	ctx := context.Background()
	ctx = auth.AddAccountIDToCtx(ctx, "test-account")
	ctx = auth.AddAuthErrorToCtx(ctx, "auth failed")

	accountID := auth.GetAccountIDFromCtx(ctx)
	assert.Equal(t, "test-account", accountID)

	err := auth.GetAuthErrorFromCtx(ctx)
	require.Error(t, err)
	assert.Equal(t, "auth failed", err.Error())
}

func TestNewAccountIDMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		authBackend consts.AuthBackend
		apiKey      string
		auth0Domain string
		auth0Aud    string
	}{
		{
			name:        "shared API key backend",
			authBackend: consts.AuthBackendSharedAPIKey,
			apiKey:      "test-api-key",
		},
		{
			name:        "auth0 backend",
			authBackend: consts.AuthBackendAuth0,
			auth0Domain: "test.auth0.com",
			auth0Aud:    "test-audience",
		},
		{
			name:        "default backend (empty) falls back to shared API key",
			authBackend: "",
			apiKey:      "default-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AuthBackend:   tt.authBackend,
				APIKeySecret:  tt.apiKey,
				Auth0Domain:   tt.auth0Domain,
				Auth0Audience: tt.auth0Aud,
			}

			middleware := NewAccountIDMiddleware(cfg)

			assert.NotNil(t, middleware)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			assert.True(t, nextCalled)
		})
	}
}

func TestSharedAPIKeyMiddleware(t *testing.T) {
	t.Run("missing auth header", func(t *testing.T) {
		middleware := sharedAPIKeyMiddleware("secret-key")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK)

		err := auth.GetAuthErrorFromCtx(capturedContext)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing auth header")

		accountID := auth.GetAccountIDFromCtx(capturedContext)
		assert.Empty(t, accountID)
	})

	tests := []struct {
		name          string
		authHeader    string
		expectedError string
	}{
		{
			name:          "malformed auth header - no Bearer prefix",
			authHeader:    "secret-key",
			expectedError: "malformed auth header",
		},
		{
			name:          "invalid API key",
			authHeader:    "Bearer wrong-key",
			expectedError: "invalid API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := sharedAPIKeyMiddleware("secret-key")

			var capturedContext context.Context
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			assert.True(t, w.Code == http.StatusOK)

			err := auth.GetAuthErrorFromCtx(capturedContext)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			accountID := auth.GetAccountIDFromCtx(capturedContext)
			assert.Empty(t, accountID)
		})
	}

	t.Run("valid API key", func(t *testing.T) {
		middleware := sharedAPIKeyMiddleware("secret-key")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer secret-key")
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		err := auth.GetAuthErrorFromCtx(capturedContext)
		assert.NoError(t, err)

		accountID := auth.GetAccountIDFromCtx(capturedContext)
		assert.Equal(t, adminAccountID, accountID)
	})

	t.Run("valid API key with different case", func(t *testing.T) {
		middleware := sharedAPIKeyMiddleware("Secret-Key")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer Secret-Key")
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		err := auth.GetAuthErrorFromCtx(capturedContext)
		assert.NoError(t, err)

		accountID := auth.GetAccountIDFromCtx(capturedContext)
		assert.Equal(t, adminAccountID, accountID)
	})
}

func TestAuth0Middleware_ContextPropagation(t *testing.T) {
	jwksServer := setupMockJWKSServer()
	defer jwksServer.Close()

	t.Run("auth0 middleware continues without auth header", func(t *testing.T) {
		middleware := auth0Middleware(jwksServer.URL[7:], "test-audience")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		err := auth.GetAuthErrorFromCtx(capturedContext)
		assert.NoError(t, err)

		accountID := auth.GetAccountIDFromCtx(capturedContext)
		assert.Empty(t, accountID)
	})
}

func TestAuth0Middleware_MalformedAuthHeader(t *testing.T) {
	jwksServer := setupMockJWKSServer()
	defer jwksServer.Close()

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "auth header without Bearer prefix",
			authHeader: "invalid-token",
		},
		{
			name:       "auth header with lowercase bearer",
			authHeader: "bearer token",
		},
		{
			name:       "auth header with Bearer but no space",
			authHeader: "Bearertoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := auth0Middleware(jwksServer.URL[7:], "test-audience")

			var capturedContext context.Context
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set(authHeader, tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			err := auth.GetAuthErrorFromCtx(capturedContext)
			assert.NoError(t, err)

			accountID := auth.GetAccountIDFromCtx(capturedContext)
			assert.Empty(t, accountID)
		})
	}
}

func TestAuth0Middleware_InvalidJWTScenarios(t *testing.T) {
	jwksServer := setupMockJWKSServer()
	defer jwksServer.Close()

	tests := []struct {
		name       string
		token      string
		wantErrMsg string
	}{
		{
			name:       "empty token",
			token:      "",
			wantErrMsg: "invalid JWT token",
		},
		{
			name:       "malformed JWT - missing header",
			token:      "invalid.jwt",
			wantErrMsg: "invalid JWT token",
		},
		{
			name:       "malformed JWT - invalid base64",
			token:      "invalid.token.format@#",
			wantErrMsg: "invalid JWT token",
		},
		{
			name:       "malformed JWT - wrong separator",
			token:      "invalid|token|format",
			wantErrMsg: "invalid JWT token",
		},
		{
			name: "JWT with invalid signature",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InRlc3Qta2V5LWlkIn0." +
				"eyJzdWIiOiJ0ZXN0LXVzZXItMTIzIiwiaXNzIjoiaHR0cHM6Ly90ZXN0LmF1dGgwLmNvbS8iLCJhdWQiOiJ0" +
				"ZXN0LWF1ZGllbmNlIiwiZXhwIjoxOTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDB9.invalid_signature",
			wantErrMsg: "invalid JWT token",
		},
		{
			name: "expired JWT token",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InRlc3Qta2V5LWlkIn0." +
				"eyJzdWIiOiJ0ZXN0LXVzZXItMTIzIiwiaXNzIjoiaHR0cHM6Ly90ZXN0LmF1dGgwLmNvbS8iLCJhdWQiOiJ0" +
				"ZXN0LWF1ZGllbmNlIiwiZXhwIjoxNjAwMDAwMDAwLCJpYXQiOjE1OTk5OTk5OTl9.invalid_signature",
			wantErrMsg: "invalid JWT token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := auth0Middleware(jwksServer.URL[7:], "test-audience")

			var capturedContext context.Context
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			req.Header.Set(authHeader, authHeaderPrefix+tt.token)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			err := auth.GetAuthErrorFromCtx(capturedContext)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)

			accountID := auth.GetAccountIDFromCtx(capturedContext)
			assert.Empty(t, accountID)
		})
	}
}

func TestAuth0Middleware_ClaimsParsingFailure(t *testing.T) {
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jwksResponse := `{
			"keys": [{
				"kty": "RSA",
				"kid": "test-key-id",
				"n": "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1` +
			`L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-` +
			`65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08q` +
			`NLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awap` +
			`JzKnqDKgw",
				"e": "AQAB",
				"alg": "RS256"
			}]
		}`
		if _, err := w.Write([]byte(jwksResponse)); err != nil {
			panic(fmt.Sprintf("failed to write JWKS response: %v", err))
		}
	}))
	defer jwksServer.Close()

	t.Run("claims type assertion failure", func(t *testing.T) {
		middleware := auth0Middleware(jwksServer.URL[7:], "test-audience")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		req.Header.Set(authHeader, authHeaderPrefix+
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InRlc3Qta2V5LWlkIn0."+
			"eyJzdWIiOiJ0ZXN0LXVzZXItMTIzIiwiaXNzIjoiaHR0cHM6Ly90ZXN0LmF1dGgwLmNvbS8iLCJhdWQiOiJ0"+
			"ZXN0LWF1ZGllbmNlIiwiZXhwIjoxOTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDB9.invalid_signature")
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		err := auth.GetAuthErrorFromCtx(capturedContext)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT token")

		accountID := auth.GetAccountIDFromCtx(capturedContext)
		assert.Empty(t, accountID)
	})
}
