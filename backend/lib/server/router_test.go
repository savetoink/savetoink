package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
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
			assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"),
				"CORS header should be set")
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

		var resp healthResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})
}

func TestSetupRoutes_ArticleRoutes(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := newRouterWithClient(cfg, http.DefaultClient)

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

	t.Run("POST /v1/articles with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/v1/articles",
			strings.NewReader(`{"url":"https://example.com"}`),
		)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("DELETE /v1/articles with authentication", func(t *testing.T) {
		req := httptest.NewRequestWithContext(
			context.Background(),
			"DELETE",
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

		router := newRouterWithClient(cfg, http.DefaultClient)
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

		router := newRouterWithClient(cfg, http.DefaultClient)
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
	router := newRouterWithClient(cfg, http.DefaultClient)
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
	router := newRouterWithClient(cfg, http.DefaultClient)
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
		cfg := setupMinimalConfig(t)
		cfg.AuthBackend = consts.AuthBackendAuth0
		cfg.Auth0Domain = "test.auth0.com"
		cfg.Auth0Audience = "test-audience"
		cfg.Auth0ClientID = "test-client-id"
		cfg.Auth0ClientSecret = "test-client-secret"

		router := newRouterWithClient(cfg, http.DefaultClient)

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

		router := newRouterWithClient(cfg, http.DefaultClient)

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

		router := newRouterWithClient(cfg, http.DefaultClient)

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

		router := newRouterWithClient(cfg, http.DefaultClient)

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

		_ = newRouterWithClient(cfg, http.DefaultClient)
	})

	t.Run("sets debug level when debug is true", func(t *testing.T) {
		cfg := setupMinimalConfig(t)
		cfg.Debug = true
		cfg.LoggingProvider = consts.LoggingBackendNone

		_ = newRouterWithClient(cfg, http.DefaultClient)
	})
}

func TestSetupLogging_WithoutSentry(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.LoggingProvider = consts.LoggingBackendNone

	_ = newRouterWithClient(cfg, http.DefaultClient)
}

func TestSetupLogging_WithSentry(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.LoggingProvider = consts.LoggingBackendSentry
	cfg.SentryDSN = testSentryDSN
	cfg.SentryEnvironment = testSentryEnvironment
	cfg.SentrySampleRate = 1.0

	_ = newRouterWithClient(cfg, http.DefaultClient)
}

func TestSetupLogging_WithSentryAndDebug(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.Debug = true
	cfg.LoggingProvider = consts.LoggingBackendSentry
	cfg.SentryDSN = testSentryDSN
	cfg.SentryEnvironment = testSentryEnvironment
	cfg.SentrySampleRate = 1.0

	_ = newRouterWithClient(cfg, http.DefaultClient)
}

func TestSetupLogging_SentryInitializationError(t *testing.T) {
	t.Cleanup(cleanupLogging)

	cfg := setupMinimalConfig(t)
	cfg.LoggingProvider = consts.LoggingBackendSentry
	cfg.SentryDSN = "invalid-dsn"
	cfg.SentryEnvironment = "test"
	cfg.SentrySampleRate = 1.0

	_ = newRouterWithClient(cfg, http.DefaultClient)
}

func TestIntegration_HealthEndpointAccessible(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := NewRouter(cfg)

	req := httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/v1/health",
		http.NoBody,
	)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp healthResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}

func TestIntegration_AuthenticatedRouteRequiresAuth(t *testing.T) {
	cfg := setupMinimalConfig(t)
	router := NewRouter(cfg)

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
