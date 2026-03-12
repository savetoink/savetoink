package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/auth"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
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
		req.Header.Set(authHeader, authHeaderPrefix+"invalid.jwt.token")
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

		err := auth.GetAuthErrorFromCtx(capturedContext)
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
