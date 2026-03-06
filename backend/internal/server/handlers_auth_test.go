package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/internal/config"
	"github.com/shaftoe/savetoink/backend/internal/consts"
)

const (
	oauthTokenPath = "/oauth/token" //nolint:gosec // test constant, not a credential
)

func TestAuthTokenExchange_MissingCode(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	h := newHandlers(cfg, nil, http.DefaultClient)

	body := authTokenExchangeRequest{
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthTokenExchange_MissingRedirectURI(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	h := newHandlers(cfg, nil, http.DefaultClient)

	body := authTokenExchangeRequest{
		Code: "test-code",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthTokenExchange_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	h := newHandlers(cfg, nil, http.DefaultClient)

	req := httptest.NewRequest("POST", "/v1/auth/token", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewRouter_AuthTokenRoute(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != oauthTokenPath {
			t.Errorf("Expected path /oauth/token, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(authTokenExchangeResponse{
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

	body := authTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthTokenExchange_ExplicitGrantType(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != oauthTokenPath {
			t.Errorf("Expected path /oauth/token, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(authTokenExchangeResponse{
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
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, "https://"),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	h := newHandlers(cfg, nil, client)

	body := authTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
		GrantType:   "authorization_code",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthTokenRoute_Registered(t *testing.T) {
	cfg := &config.Config{
		APIKeySecret:      "test-key",
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	r := newRouterWithClient(cfg, client)

	body := authTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Errorf("route should be registered when AuthBackend is Auth0, got status %d", w.Code)
	}
}
