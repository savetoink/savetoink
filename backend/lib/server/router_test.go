package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/server/auth"
	"github.com/shaftoe/savetoink/backend/lib/server/handlers"
	"github.com/shaftoe/savetoink/backend/lib/server/types"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"golang.org/x/net/html"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKeySecret         = "test-api-key-secret" //nolint:gosec // test credential
	testMailjetAPIKey        = "test-key"
	testMailjetAPISecret     = "test-secret"
	testMailjetWebhookSecret = "test-webhook"
	testSenderEmail          = "sender@example.com"
	testSentryDSN            = "https://test@sentry.io/123"
	testSentryEnvironment    = "test"
	testDeviceEmail          = "test@kindle.com"
)

type mockService struct{}

func (m *mockService) Fetch(_ context.Context, _ *url.URL) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockService) ParseHTML(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockService) Clean(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("epub data"))), nil
}

func (m *mockService) SendArticle(
	_ context.Context,
	_ string,
	_ io.ReadCloser,
	_ string,
) (*email.SendEmailResponse, error) {
	return &email.SendEmailResponse{
		MessageID: "test-message-id",
	}, nil
}

func (m *mockService) SendArticleByID(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockService) CreateArticle(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockService) UpdateArticle(_ context.Context, _ *model.Article) error {
	return nil
}

func (m *mockService) GetArticle(_ context.Context, _, articleID string) (*model.Article, error) {
	if articleID == "test-id" {
		return &model.Article{
			ID:    "test-id",
			Title: "Test Article",
			URL:   "https://example.com",
		}, nil
	}
	return nil, errors.New("article not found")
}

func (m *mockService) GetArticlesMetadata(
	_ context.Context, _ string, _, _ int, _ *bool,
) (*servicetypes.GetArticlesResult, error) {
	return &servicetypes.GetArticlesResult{
		Articles: []*model.Article{},
		Page:     1,
		PageSize: 10,
		Total:    0,
		HasMore:  false,
	}, nil
}

func (m *mockService) DeleteArticle(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockService) GetDBError() error {
	return nil
}

func (m *mockService) GetUserDeviceEmailAndAutoSend(
	_ context.Context, _ string,
) (deviceEmail string, autoSend bool, err error) {
	return "", false, nil
}

func (m *mockService) SetUserDeviceEmailWithAutoSend(_ context.Context, _, _ string, _ bool) error {
	return errors.New("not implemented in mock")
}

func (m *mockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	return errors.New("not implemented in mock")
}

func (m *mockService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	return &model.UserProfile{Email: "test@example.com"}, nil
}

func (m *mockService) SetUserEmail(_ context.Context, _, _ string) error {
	return errors.New("not implemented in mock")
}

func (m *mockService) DeleteUserProfile(_ context.Context, _ string) error {
	return errors.New("not implemented in mock")
}

func (m *mockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented in mock")
}

func (m *mockService) CountSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return 0, nil
}

func (m *mockService) HandleBounce(_ context.Context, _, _ string) error {
	return errors.New("not implemented in mock")
}

func (m *mockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (m *mockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented in mock")
}

func newTestRouter(cfg *config.Config, client *http.Client) *chi.Mux {
	r := chi.NewRouter()
	svc := &mockService{}

	h := handlers.New(
		cfg,
		svc,
		client,
		nil,
	)

	r.Use(auth.NewAccountIDMiddleware(cfg))
	r.Use(requestIDMiddleware)
	r.Use(logging.Middleware)
	r.Use(newCorsMiddleware(cfg))
	r.Use(jsonContentTypeMiddleware)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "not_found"})
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: "method_not_allowed"})
	})

	setupRoutes(r, h, cfg, svc)

	return r
}

func setupMinimalConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		APIKeySecret:         testAPIKeySecret,
		AuthBackend:          consts.AuthBackendSharedAPIKey,
		Debug:                false,
		ArticlesTable:        "test-articles",
		UserProfileTable:     "test-profiles",
		SendsTable:           "test-sends",
		AppURL:               "https://test.com",
		CorsAllowOrigin:      "",
		EmailProvider:        "",
		LoggingProvider:      consts.LoggingBackendNone,
		SentryDSN:            "",
		SentryEnvironment:    "",
		SentrySampleRate:     0.0,
		Auth0Domain:          "",
		Auth0Audience:        "",
		Auth0ClientID:        "",
		Auth0ClientSecret:    "",
		MailjetAPIKey:        "",
		MailjetAPISecret:     "",
		MailjetWebhookSecret: "",
		SenderEmail:          "",
	}
}

func cleanupLogging() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
}

func TestNewRouter(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := NewRouter(cfg)

	assert.NotNil(t, router, "router should not be nil")
}

func TestNewRouterWithClient(t *testing.T) {
	t.Run("creates router with all middleware", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		client := &http.Client{Timeout: 5 * time.Second}
		router := newRouterWithClient(cfg, client)

		assert.NotNil(t, router, "router should not be nil")

		t.Run("middleware are registered", func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				context.Background(),
				"GET",
				"/v1/health",
				http.NoBody,
			)
			req.Header.Set("Authorization", "Bearer "+testAPIKeySecret)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEmpty(t, w.Header().Get("Content-Type"),
				"Content-Type header should be set")
			assert.NotEmpty(t, w.Header().Get("X-Request-ID"),
				"X-Request-ID header should be set")
		})
	})

	t.Run("not found handler is registered", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		router := newRouterWithClient(cfg, http.DefaultClient)

		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/nonexistent",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp model.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "not_found", resp.Error)
	})

	t.Run("method not allowed handler is registered", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		router := newRouterWithClient(cfg, http.DefaultClient)

		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/health",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		var resp model.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "method_not_allowed", resp.Error)
	})
}

func TestSetupRoutes_BaseRoutes(t *testing.T) {
	t.Run("health endpoint is registered", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		router := newRouterWithClient(cfg, http.DefaultClient)

		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/health",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp types.HealthResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})
}

func TestSetupRoutes_ArticleRoutes(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := newTestRouter(cfg, http.DefaultClient)

	authHeader := "Bearer " + testAPIKeySecret

	t.Run("GET /v1/articles requires authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/articles",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GET /v1/articles with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/articles",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET /v1/articles/{id} with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/articles/test-id",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("DELETE /v1/articles/{id} with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"DELETE",
			"/v1/articles/test-id",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("PUT /v1/articles/{id}/favorite with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"PUT",
			"/v1/articles/test-id/favorite",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestSetupRoutes_SendArticleRoute(t *testing.T) {
	t.Run("send route registered when mailjet is configured", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.EmailProvider = consts.EmailBackendMailjet
		cfg.MailjetAPIKey = testMailjetAPIKey
		cfg.MailjetAPISecret = testMailjetAPISecret
		cfg.MailjetWebhookSecret = testMailjetWebhookSecret
		cfg.SenderEmail = testSenderEmail

		router := newTestRouter(cfg, http.DefaultClient)
		authHeader := "Bearer " + testAPIKeySecret

		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/articles/test-id/send",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("send route not registered when mailjet is not configured", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.EmailProvider = ""

		router := newTestRouter(cfg, http.DefaultClient)
		authHeader := "Bearer " + testAPIKeySecret

		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/articles/test-id/send",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var resp model.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSetupRoutes_UserProfileRoutes(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := newTestRouter(cfg, http.DefaultClient)
	authHeader := "Bearer " + testAPIKeySecret

	t.Run("GET /v1/user/profile requires authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/user/profile",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GET /v1/user/profile with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/user/profile",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestSetupRoutes_DeviceRoutes(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := newTestRouter(cfg, http.DefaultClient)
	authHeader := "Bearer " + testAPIKeySecret

	t.Run("PUT /v1/devices requires authentication", func(t *testing.T) {
		body := strings.NewReader(
			`{"device_email":"` + testDeviceEmail + `","auto_send":true}`,
		)
		req := httptest.NewRequestWithContext(
			context.Background(),
			"PUT",
			"/v1/devices",
			body,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("PUT /v1/devices with authentication", func(t *testing.T) {
		body := strings.NewReader(
			`{"device_email":"` + testDeviceEmail + `","auto_send":true}`,
		)
		req := httptest.NewRequestWithContext(
			context.Background(),
			"PUT",
			"/v1/devices",
			body,
		)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("DELETE /v1/devices requires authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"DELETE",
			"/v1/devices",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("DELETE /v1/devices with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"DELETE",
			"/v1/devices",
			http.NoBody,
		)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestSetupRoutes_AuthRoute(t *testing.T) {
	t.Run("auth/token route registered when auth0 is configured", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   "3600",
			})
		}))
		defer server.Close()

		parsedURL, _ := url.Parse(server.URL)
		client := server.Client()
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}

		cfg := setupMinimalConfig(t)
		cfg.AuthBackend = consts.AuthBackendAuth0
		cfg.Auth0Domain = strings.TrimPrefix(parsedURL.Host, "https://")
		cfg.Auth0Audience = "test-audience"
		cfg.Auth0ClientID = "test-client-id"
		cfg.Auth0ClientSecret = "test-client-secret"

		router := newTestRouter(cfg, client)

		body := strings.NewReader(
			`{"code":"test","redirect_uri":"http://localhost"}`,
		)
		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/auth/token",
			body,
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("auth/token route not registered when shared api key is configured", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.AuthBackend = consts.AuthBackendSharedAPIKey

		router := newTestRouter(cfg, http.DefaultClient)

		body := strings.NewReader(
			`{"code":"test","redirect_uri":"http://localhost"}`,
		)
		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/auth/token",
			body,
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSetupRoutes_WebhookRoute(t *testing.T) {
	t.Run("webhooks/mailjet route registered when mailjet is configured", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.EmailProvider = consts.EmailBackendMailjet
		cfg.MailjetAPIKey = testMailjetAPIKey
		cfg.MailjetAPISecret = testMailjetAPISecret
		cfg.MailjetWebhookSecret = testMailjetWebhookSecret
		cfg.SenderEmail = testSenderEmail

		router := newTestRouter(cfg, http.DefaultClient)

		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/webhooks/mailjet",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("webhooks/mailjet route not registered when mailjet is not configured", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.EmailProvider = ""

		router := newTestRouter(cfg, http.DefaultClient)

		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/webhooks/mailjet",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSetupLogging_LogLevel(t *testing.T) {
	t.Cleanup(cleanupLogging)

	t.Run("sets info level when debug is false", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.Debug = false
		cfg.LoggingProvider = consts.LoggingBackendNone

		_ = newTestRouter(cfg, http.DefaultClient)
	})

	t.Run("sets debug level when debug is true", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.Debug = true
		cfg.LoggingProvider = consts.LoggingBackendNone

		_ = newTestRouter(cfg, http.DefaultClient)
	})
}

func TestSetupLogging_WithoutSentry(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.LoggingProvider = consts.LoggingBackendNone

	_ = newTestRouter(cfg, http.DefaultClient)
}

func TestSetupLogging_WithSentry(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.LoggingProvider = consts.LoggingBackendSentry
	cfg.SentryDSN = testSentryDSN
	cfg.SentryEnvironment = testSentryEnvironment
	cfg.SentrySampleRate = 1.0

	_ = newTestRouter(cfg, http.DefaultClient)
}

func TestSetupLogging_WithSentryAndDebug(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.Debug = true
	cfg.LoggingProvider = consts.LoggingBackendSentry
	cfg.SentryDSN = testSentryDSN
	cfg.SentryEnvironment = testSentryEnvironment
	cfg.SentrySampleRate = 1.0

	_ = newTestRouter(cfg, http.DefaultClient)
}

func TestSetupLogging_SentryInitializationError(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.LoggingProvider = consts.LoggingBackendSentry
	cfg.SentryDSN = "invalid-dsn"
	cfg.SentryEnvironment = "test"
	cfg.SentrySampleRate = 1.0

	_ = newTestRouter(cfg, http.DefaultClient)
}

func TestIntegration_HealthEndpointAccessible(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := newTestRouter(cfg, http.DefaultClient)

	req := httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/v1/health",
		http.NoBody,
	)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}

func TestIntegration_AuthenticatedRouteRequiresAuth(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := newTestRouter(cfg, http.DefaultClient)

	t.Run("articles endpoint without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/articles",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user profile endpoint without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/v1/user/profile",
			http.NoBody,
		)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("devices endpoint without auth returns 401", func(t *testing.T) {
		body := strings.NewReader(
			`{"device_email":"` + testDeviceEmail + `"}`,
		)
		req := httptest.NewRequestWithContext(
			context.Background(),
			"PUT",
			"/v1/devices",
			body,
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestNewRouter_AuthTokenRoute(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != handlers.OauthTokenPath {
			t.Errorf("Expected path /oauth/token, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(types.AuthTokenExchangeResponse{ //nolint:gosec // test mock token, not a real secret
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	client := server.Client()
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	cfg := &config.Config{
		APIKeySecret:      "test-key",
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, "https://"),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	r := newRouterWithClient(cfg, client)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthTokenRoute_Registered(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(types.AuthTokenExchangeResponse{ //nolint:gosec // test mock token, not a real secret
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	client := server.Client()
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	cfg := &config.Config{
		APIKeySecret:      "test-key",
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, "https://"),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	r := newRouterWithClient(cfg, client)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Errorf("route should be registered when AuthBackend is Auth0, got status %d", w.Code)
	}
}
