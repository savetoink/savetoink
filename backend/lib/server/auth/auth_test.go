package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
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

func (m *MockService) Fetch(_ context.Context, _ string) ([]byte, content.FetcherType, error) {
	panic("not implemented")
}

func (m *MockService) Extract(_ context.Context, _ []byte) (*model.Article, error) {
	panic("not implemented")
}

func (m *MockService) GenerateEPUB(_ *model.Article) ([]byte, error) {
	panic("not implemented")
}

func (m *MockService) SendArticle(_ context.Context, _ string, _ []byte, _ string) (*email.SendEmailResponse, error) {
	panic("not implemented")
}

func (m *MockService) CreateArticle(_ context.Context, _, _ string) (*model.Article, error) {
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

func (m *MockService) DeleteAllArticles(_ context.Context, _ string) (*servicetypes.DeleteArticleResult, error) {
	panic("not implemented")
}

func (m *MockService) GetDBError() error {
	panic("not implemented")
}

func (m *MockService) GetUserDeviceEmail(_ context.Context, _ string) (deviceEmail string, autoSend bool, err error) {
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
			setupCtx: func() context.Context { return addAccountIDToContext(context.Background(), "test-account") },
			expected: "test-account",
		},
		{
			name:     "no account ID in context",
			setupCtx: context.Background,
			expected: "",
		},
		{
			name:     "wrong type in context",
			setupCtx: func() context.Context { return context.WithValue(context.Background(), auth.AccountIDKey, 123) },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := auth.GetAccountID(ctx)
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
		req.Header.Set(authHeader, authHeaderPrefix+"invalid.jwt.token")
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		err := auth.GetAuthError(capturedContext)
		if err == nil {
			t.Error("expected auth error in context")
		}
	})

	t.Run("JWT with invalid signature adds error to context", func(t *testing.T) {
		jwksServer := setupMockJWKSServer()
		defer jwksServer.Close()

		middleware := auth0Middleware(jwksServer.URL[7:], "test-audience")

		var capturedContext context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InRlc3Qta2V5LWlkIn0." +
			"eyJzdWIiOiJ0ZXN0LXVzZXItMTIzIiwiaXNzIjoiaHR0cHM6Ly90ZXN0LmF1dGgwLmNvbS8iLCJhdWQiOiJ0" +
			"ZXN0LWF1ZGllbmNlIiwiZXhwIjoxOTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDB9.invalid_signature"
		req.Header.Set(authHeader, authHeaderPrefix+token)
		w := httptest.NewRecorder()

		middleware(next).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		err := auth.GetAuthError(capturedContext)
		if err == nil {
			t.Error("expected auth error for invalid signature")
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

			err := auth.GetAuthError(capturedContext)
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
				return context.WithValue(context.Background(), auth.AuthErrorKey, "authentication failed")
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

	ctx := addAccountIDToContext(context.Background(), "test-account")
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

func TestGetSendsCount(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected int
	}{
		{
			name:     "sends count in context",
			setupCtx: func() context.Context { return context.WithValue(context.Background(), auth.SendsCountKey, 5) },
			expected: 5,
		},
		{
			name:     "no sends count in context",
			setupCtx: context.Background,
			expected: 0,
		},
		{
			name: "wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), auth.SendsCountKey, "not-an-int")
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := auth.GetSendsCount(ctx)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestHasSendsCount(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected bool
	}{
		{
			name:     "sends count in context",
			setupCtx: func() context.Context { return context.WithValue(context.Background(), auth.SendsCountKey, 5) },
			expected: true,
		},
		{
			name:     "no sends count in context",
			setupCtx: context.Background,
			expected: false,
		},
		{
			name: "wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), auth.SendsCountKey, "not-an-int")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := auth.HasSendsCount(ctx)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestSharedAPIKeyMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		apiKeySecret   string
		authHeader     string
		expectedStatus int
		expectedAccID  string
		expectedError  string
	}{
		{
			name:           "valid API key",
			apiKeySecret:   "test-secret-key",
			authHeader:     "Bearer test-secret-key",
			expectedStatus: http.StatusOK,
			expectedAccID:  adminAccountID,
			expectedError:  "",
		},
		{
			name:           "invalid API key",
			apiKeySecret:   "test-secret-key",
			authHeader:     "Bearer wrong-key",
			expectedStatus: http.StatusOK,
			expectedAccID:  "",
			expectedError:  "invalid API key",
		},
		{
			name:           "missing authorization header",
			apiKeySecret:   "test-secret-key",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedAccID:  "",
			expectedError:  "missing or malformed auth header",
		},
		{
			name:           "malformed authorization header (no Bearer prefix)",
			apiKeySecret:   "test-secret-key",
			authHeader:     "test-secret-key",
			expectedStatus: http.StatusOK,
			expectedAccID:  "",
			expectedError:  "missing or malformed auth header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := sharedAPIKeyMiddleware(tt.apiKeySecret)

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

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			accountID := auth.GetAccountID(capturedContext)
			if accountID != tt.expectedAccID {
				t.Errorf("expected account ID '%s', got '%s'", tt.expectedAccID, accountID)
			}

			authErr := auth.GetAuthError(capturedContext)
			if tt.expectedError == "" && authErr != nil {
				t.Errorf("expected no auth error, got '%v'", authErr)
			}
			if tt.expectedError != "" && (authErr == nil || authErr.Error() != tt.expectedError) {
				t.Errorf("expected auth error '%s', got '%v'", tt.expectedError, authErr)
			}
		})
	}
}

func TestNewAccountIDMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		authBackend   consts.AuthBackend
		apiKeySecret  string
		auth0Domain   string
		auth0Audience string
		authHeader    string
	}{
		{
			name:          "shared API key backend",
			authBackend:   consts.AuthBackendSharedAPIKey,
			apiKeySecret:  "test-secret",
			auth0Domain:   "",
			auth0Audience: "",
			authHeader:    "Bearer test-secret",
		},
		{
			name:          "Auth0 backend",
			authBackend:   consts.AuthBackendAuth0,
			apiKeySecret:  "",
			auth0Domain:   "test.auth0.com",
			auth0Audience: "test-audience",
			authHeader:    "",
		},
		{
			name:          "unknown backend defaults to shared API key",
			authBackend:   "unknown",
			apiKeySecret:  "test-secret",
			auth0Domain:   "",
			auth0Audience: "",
			authHeader:    "Bearer test-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AuthBackend:   tt.authBackend,
				APIKeySecret:  tt.apiKeySecret,
				Auth0Domain:   tt.auth0Domain,
				Auth0Audience: tt.auth0Audience,
			}

			middleware := NewAccountIDMiddleware(cfg)

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

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if tt.authBackend == consts.AuthBackendSharedAPIKey && tt.authHeader != "" {
				accountID := auth.GetAccountID(capturedContext)
				if accountID != adminAccountID {
					t.Errorf("expected account ID '%s', got '%s'", adminAccountID, accountID)
				}
			}
		})
	}
}

func TestNewBouncingEmailMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		deviceEmail    string
		autoSend       bool
		emailErr       error
		isBouncing     bool
		bouncingErr    error
		userProfile    *model.UserProfile
		profileErr     error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no account ID in context",
			accountID:      "",
			deviceEmail:    "test@example.com",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "no device email set",
			accountID:      "test-account",
			deviceEmail:    "",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "device email not bouncing",
			accountID:      "test-account",
			deviceEmail:    "test@example.com",
			autoSend:       true,
			isBouncing:     false,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:        "device email bouncing with bounce info",
			accountID:   "test-account",
			deviceEmail: "test@example.com",
			autoSend:    true,
			isBouncing:  true,
			userProfile: &model.UserProfile{
				BouncedEmails: map[string]model.BounceInfo{
					"test@example.com": {Error: "bounced due to invalid mailbox"},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "device email test@example.com is blocked due to previous bounce: bounced due to invalid mailbox",
		},
		{
			name:           "device email bouncing without bounce info",
			accountID:      "test-account",
			deviceEmail:    "test@example.com",
			autoSend:       true,
			isBouncing:     true,
			userProfile:    &model.UserProfile{BouncedEmails: map[string]model.BounceInfo{}},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "device email test@example.com is blocked due to previous bounce",
		},
		{
			name:           "service error getting device email",
			accountID:      "test-account",
			emailErr:       errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get user device email: database error",
		},
		{
			name:           "service error checking bounce",
			accountID:      "test-account",
			deviceEmail:    "test@example.com",
			autoSend:       true,
			bouncingErr:    errors.New("bounce check failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to check if email is bouncing: bounce check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				deviceEmail:    tt.deviceEmail,
				autoSend:       tt.autoSend,
				emailErr:       tt.emailErr,
				isBouncing:     tt.isBouncing,
				bouncingErr:    tt.bouncingErr,
				userProfile:    tt.userProfile,
				userProfileErr: tt.profileErr,
			}

			middleware := NewBouncingEmailMiddleware(mockSvc)

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			ctx := context.Background()
			if tt.accountID != "" {
				ctx = addAccountIDToContext(ctx, tt.accountID)
			}
			req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error != tt.expectedError {
					t.Errorf("expected error '%s', got '%s'", tt.expectedError, resp.Error)
				}
			}
		})
	}
}

func TestCheckDeviceEmail(t *testing.T) {
	tests := []struct {
		name         string
		deviceEmail  string
		autoSend     bool
		emailErr     error
		isBouncing   bool
		bouncingErr  error
		expectedSkip bool
		expectedErr  string
	}{
		{
			name:         "empty device email skips check",
			deviceEmail:  "",
			expectedSkip: true,
			expectedErr:  "",
		},
		{
			name:         "email not bouncing skips",
			deviceEmail:  "test@example.com",
			isBouncing:   false,
			expectedSkip: true,
			expectedErr:  "",
		},
		{
			name:         "email is bouncing doesn't skip",
			deviceEmail:  "test@example.com",
			isBouncing:   true,
			expectedSkip: false,
			expectedErr:  "",
		},
		{
			name:         "service error getting email",
			emailErr:     errors.New("database error"),
			expectedSkip: false,
			expectedErr:  "failed to get user device email: database error",
		},
		{
			name:         "service error checking bounce",
			deviceEmail:  "test@example.com",
			bouncingErr:  errors.New("bounce check failed"),
			expectedSkip: false,
			expectedErr:  "failed to check if email is bouncing: bounce check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				deviceEmail: tt.deviceEmail,
				emailErr:    tt.emailErr,
				isBouncing:  tt.isBouncing,
				bouncingErr: tt.bouncingErr,
			}

			shouldSkip, err := checkDeviceEmail(context.Background(), mockSvc, "test-account")

			if shouldSkip != tt.expectedSkip {
				t.Errorf("expected skip %t, got %t", tt.expectedSkip, shouldSkip)
			}

			if tt.expectedErr == "" && err != nil {
				t.Errorf("expected no error, got '%v'", err)
			}
			if tt.expectedErr != "" && (err == nil || err.Error() != tt.expectedErr) {
				t.Errorf("expected error '%s', got '%v'", tt.expectedErr, err)
			}
		})
	}
}

func TestHandleBouncingEmail(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		destEmail      string
		isBouncing     bool
		bouncingErr    error
		userProfile    *model.UserProfile
		profileErr     error
		expectedReturn bool
	}{
		{
			name:           "email not bouncing",
			destEmail:      "test@example.com",
			isBouncing:     false,
			expectedReturn: false,
		},
		{
			name:       "email bouncing with bounce info",
			destEmail:  "test@example.com",
			isBouncing: true,
			userProfile: &model.UserProfile{
				BouncedEmails: map[string]model.BounceInfo{
					"test@example.com": {Error: "mailbox full"},
				},
			},
			expectedReturn: true,
		},
		{
			name:           "email bouncing without bounce info",
			destEmail:      "test@example.com",
			isBouncing:     true,
			userProfile:    &model.UserProfile{BouncedEmails: map[string]model.BounceInfo{}},
			expectedReturn: true,
		},
		{
			name:           "service error getting profile",
			destEmail:      "test@example.com",
			isBouncing:     true,
			profileErr:     errors.New("database error"),
			expectedReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				isBouncing:     tt.isBouncing,
				userProfile:    tt.userProfile,
				userProfileErr: tt.profileErr,
			}

			w := httptest.NewRecorder()
			result := handleBouncingEmail(context.Background(), mockSvc, w, tt.accountID, tt.destEmail)

			if result != tt.expectedReturn {
				t.Errorf("expected return %t, got %t", tt.expectedReturn, result)
			}

			if tt.expectedReturn && tt.isBouncing {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
				}
			}
		})
	}
}

func TestIsEmailBouncing(t *testing.T) {
	tests := []struct {
		name        string
		destEmail   string
		isBouncing  bool
		bouncingErr error
		expected    bool
	}{
		{
			name:       "email is bouncing",
			destEmail:  "test@example.com",
			isBouncing: true,
			expected:   true,
		},
		{
			name:       "email is not bouncing",
			destEmail:  "test@example.com",
			isBouncing: false,
			expected:   false,
		},
		{
			name:        "service error returns false",
			destEmail:   "test@example.com",
			bouncingErr: errors.New("error"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				isBouncing:  tt.isBouncing,
				bouncingErr: tt.bouncingErr,
			}

			result := isEmailBouncing(context.Background(), mockSvc, "test-account", tt.destEmail)

			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestSendError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "error message",
			err:  errors.New("something went wrong"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			sendError(w, tt.err)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
			}

			var resp model.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Error != tt.err.Error() {
				t.Errorf("expected error '%s', got '%s'", tt.err.Error(), resp.Error)
			}
		})
	}
}

func TestSendBounceError(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		errorMessage  string
		expectedError string
	}{
		{
			name:          "with error message",
			email:         "test@example.com",
			errorMessage:  "mailbox full",
			expectedError: "device email test@example.com is blocked due to previous bounce: mailbox full",
		},
		{
			name:          "without error message",
			email:         "test@example.com",
			errorMessage:  "",
			expectedError: "device email test@example.com is blocked due to previous bounce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			sendBounceError(w, tt.email, tt.errorMessage)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			var resp model.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Error != tt.expectedError {
				t.Errorf("expected error '%s', got '%s'", tt.expectedError, resp.Error)
			}
		})
	}
}

func TestNewEmailBackendEnabledMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		emailProvider  consts.EmailProvider
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Mailjet backend configured",
			emailProvider:  consts.EmailBackendMailjet,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "empty email provider",
			emailProvider:  "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email backend not configured",
		},
		{
			name:           "different provider configured",
			emailProvider:  "other-provider",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email backend not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				EmailProvider: tt.emailProvider,
			}

			middleware := NewEmailBackendEnabledMiddleware(cfg)

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error != tt.expectedError {
					t.Errorf("expected error '%s', got '%s'", tt.expectedError, resp.Error)
				}
			}
		})
	}
}

func TestNoOpMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("next called")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	noOpMiddleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "next called" {
		t.Errorf("expected response 'next called', got '%s'", w.Body.String())
	}
}

func TestNewActiveSubscriptionMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authBackend    consts.AuthBackend
		sendCount      int
		sendCountErr   error
		expectedStatus int
		expectedCount  int
		expectedError  string
	}{
		{
			name:           "shared API key backend uses no-op",
			authBackend:    consts.AuthBackendSharedAPIKey,
			sendCount:      5,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedError:  "",
		},
		{
			name:           "Auth0 with count under limit",
			authBackend:    consts.AuthBackendAuth0,
			sendCount:      5,
			expectedStatus: http.StatusOK,
			expectedCount:  5,
			expectedError:  "",
		},
		{
			name:           "Auth0 with count at limit",
			authBackend:    consts.AuthBackendAuth0,
			sendCount:      consts.MaxFreeTierSendsPerPeriod,
			expectedStatus: http.StatusTooManyRequests,
			expectedCount:  0,
			expectedError:  "free tier limit exceeded",
		},
		{
			name:           "Auth0 with count over limit",
			authBackend:    consts.AuthBackendAuth0,
			sendCount:      consts.MaxFreeTierSendsPerPeriod + 1,
			expectedStatus: http.StatusTooManyRequests,
			expectedCount:  0,
			expectedError:  "free tier limit exceeded",
		},
		{
			name:           "Auth0 with service error",
			authBackend:    consts.AuthBackendAuth0,
			sendCount:      5,
			sendCountErr:   errors.New("database error"),
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedError:  "failed to check subscription limit: database error",
		},
		{
			name:           "unknown backend defaults to no-op",
			authBackend:    "unknown",
			sendCount:      5,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{
				sendCount:    tt.sendCount,
				sendCountErr: tt.sendCountErr,
			}

			cfg := &config.Config{
				AuthBackend: tt.authBackend,
			}

			middleware := NewActiveSubscriptionMiddleware(cfg, mockSvc)

			var capturedContext context.Context
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			ctx := addAccountIDToContext(context.Background(), "test-account")
			req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				count := auth.GetSendsCount(capturedContext)
				if count != tt.expectedCount {
					t.Errorf("expected sends count %d, got %d", tt.expectedCount, count)
				}

				if tt.expectedError != "" {
					authErr := auth.GetAuthError(capturedContext)
					if authErr == nil || authErr.Error() != tt.expectedError {
						t.Errorf("expected auth error '%s', got '%v'", tt.expectedError, authErr)
					}
				}
			}

			if tt.expectedStatus == http.StatusTooManyRequests {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Error != tt.expectedError {
					t.Errorf("expected error '%s', got '%s'", tt.expectedError, resp.Error)
				}
			}
		})
	}
}
