package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
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

func (m *MockService) Process(_ context.Context, _ string) (*servicetypes.ProcessResult, error) {
	panic("not implemented")
}

func (m *MockService) SendArticle(_ context.Context, _ *model.Article, _, _ string) (*email.SendEmailResponse, error) {
	panic("not implemented")
}

func (m *MockService) WriteToFile(_ *servicetypes.ProcessResult, _ string) error {
	panic("not implemented")
}

func (m *MockService) CreateArticle(_ context.Context, _, _ string) (*model.Article, error) {
	panic("not implemented")
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
			setupCtx: func() context.Context { return context.WithValue(context.Background(), accountIDKey, 123) },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := GetAccountID(ctx)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
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
			setupCtx: func() context.Context { return context.WithValue(context.Background(), sendsCountKey, 5) },
			expected: 5,
		},
		{
			name:     "no sends count in context",
			setupCtx: context.Background,
			expected: 0,
		},
		{
			name:     "wrong type in context",
			setupCtx: func() context.Context { return context.WithValue(context.Background(), sendsCountKey, "five") },
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := GetSendsCount(ctx)
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
			setupCtx: func() context.Context { return context.WithValue(context.Background(), sendsCountKey, 5) },
			expected: true,
		},
		{
			name:     "no sends count in context",
			setupCtx: context.Background,
			expected: false,
		},
		{
			name:     "wrong type in context",
			setupCtx: func() context.Context { return context.WithValue(context.Background(), sendsCountKey, "five") },
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := HasSendsCount(ctx)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
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

	if resp.Error != "unauthorized" {
		t.Errorf("expected error message 'unauthorized', got '%s'", resp.Error)
	}
}

func TestBouncingEmailMiddleware_NoAccountID(t *testing.T) {
	mockSvc := &MockService{}
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBouncingEmailMiddleware_NoDeviceEmail(t *testing.T) {
	mockSvc := &MockService{
		deviceEmail: "",
		autoSend:    false,
	}
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBouncingEmailMiddleware_GetDeviceEmailError(t *testing.T) {
	mockSvc := &MockService{
		emailErr: errors.New("database error"),
	}
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := "failed to get user device email: database error"
	if resp.Error != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestBouncingEmailMiddleware_CheckBouncingError(t *testing.T) {
	mockSvc := &MockService{
		deviceEmail: "test@kindle.com",
		isBouncing:  false,
		bouncingErr: errors.New("bounce check failed"),
	}
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := "failed to check if email is bouncing: bounce check failed"
	if resp.Error != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestBouncingEmailMiddleware_NotBouncing(t *testing.T) {
	mockSvc := &MockService{
		deviceEmail: "test@kindle.com",
		isBouncing:  false,
	}
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBouncingEmailMiddleware_BouncingWithError(t *testing.T) {
	mockSvc := &MockService{
		deviceEmail: "test@kindle.com",
		isBouncing:  true,
		userProfile: &model.UserProfile{
			BouncedEmails: map[string]model.BounceInfo{
				"test@kindle.com": {
					Timestamp: time.Now(),
					Error:     "mailbox full",
				},
			},
		},
	}
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedMsg := "device email test@kindle.com is blocked due to previous bounce: mailbox full"
	if resp.Error != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestBouncingEmailMiddleware_BouncingNoError(t *testing.T) {
	testBouncingEmail(t, &MockService{
		deviceEmail: "test@kindle.com",
		isBouncing:  true,
		userProfile: &model.UserProfile{
			BouncedEmails: map[string]model.BounceInfo{},
		},
	}, http.StatusBadRequest, "device email test@kindle.com is blocked due to previous bounce")
}

func TestBouncingEmailMiddleware_BouncingProfileError(t *testing.T) {
	testBouncingEmail(t, &MockService{
		deviceEmail:    "test@kindle.com",
		isBouncing:     true,
		userProfileErr: errors.New("profile not found"),
	}, http.StatusBadRequest, "device email test@kindle.com is blocked due to previous bounce")
}

func testBouncingEmail(t *testing.T, mockSvc *MockService, expectedStatus int, expectedMsg string) {
	t.Helper()
	middleware := NewBouncingEmailMiddleware(mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestEmailBackendMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		provider     consts.EmailProvider
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "mailjet provider",
			provider:     consts.EmailBackendMailjet,
			expectedCode: http.StatusOK,
			expectedMsg:  "",
		},
		{
			name:         "empty provider",
			provider:     "",
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "email backend not configured",
		},
		{
			name:         "wrong provider",
			provider:     "wrong-provider",
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "email backend not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				EmailProvider: tt.provider,
			}
			middleware := NewEmailBackendEnabledMiddleware(cfg)

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
			w := httptest.NewRecorder()

			middleware(next).ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.expectedMsg != "" {
				var resp model.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.Error != tt.expectedMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, resp.Error)
				}
			}
		})
	}
}

func TestActiveSubscriptionMiddleware_SharedAPIKey(t *testing.T) {
	cfg := &config.Config{
		AuthBackend: consts.AuthBackendSharedAPIKey,
	}
	mockSvc := &MockService{}
	middleware := NewActiveSubscriptionMiddleware(cfg, mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestActiveSubscriptionMiddleware_Auth0_ValidSubscription(t *testing.T) {
	cfg := &config.Config{
		AuthBackend: consts.AuthBackendAuth0,
	}
	mockSvc := &MockService{
		sendCount: 5,
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

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	count := GetSendsCount(capturedContext)
	if count != 5 {
		t.Errorf("expected sends count 5, got %d", count)
	}
}

func TestActiveSubscriptionMiddleware_Auth0_LimitExceeded(t *testing.T) {
	cfg := &config.Config{
		AuthBackend: consts.AuthBackendAuth0,
	}
	mockSvc := &MockService{
		sendCount: 11,
	}
	middleware := NewActiveSubscriptionMiddleware(cfg, mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	var resp model.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != "free tier limit exceeded" {
		t.Errorf("expected error message 'free tier limit exceeded', got '%s'", resp.Error)
	}
}

func TestActiveSubscriptionMiddleware_Auth0_ServiceError(t *testing.T) {
	cfg := &config.Config{
		AuthBackend: consts.AuthBackendAuth0,
	}
	mockSvc := &MockService{
		sendCountErr: errors.New("database error"),
	}
	middleware := NewActiveSubscriptionMiddleware(cfg, mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := addAccountIDToContext(context.Background(), "test-account")
	req := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var capturedContext context.Context
	nextWithCapture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req2 := httptest.NewRequestWithContext(ctx, "GET", "/test", http.NoBody)
	w2 := httptest.NewRecorder()
	middleware(nextWithCapture).ServeHTTP(w2, req2)

	err := GetAuthError(capturedContext)
	if err == nil {
		t.Error("expected auth error in context")
	}
	expectedMsg := "failed to check subscription limit: database error"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestActiveSubscriptionMiddleware_Auth0_NoAccountID(t *testing.T) {
	cfg := &config.Config{
		AuthBackend: consts.AuthBackendAuth0,
	}
	mockSvc := &MockService{
		sendCount: 5,
	}
	middleware := NewActiveSubscriptionMiddleware(cfg, mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestActiveSubscriptionMiddleware_DefaultBackend(t *testing.T) {
	cfg := &config.Config{
		AuthBackend: "",
	}
	mockSvc := &MockService{}
	middleware := NewActiveSubscriptionMiddleware(cfg, mockSvc)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	middleware(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
