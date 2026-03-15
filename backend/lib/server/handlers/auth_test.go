package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/server/types"
	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/servicetypes"
	"golang.org/x/net/html"
)

//nolint:gosec // test constants, not real credentials
const (
	testEmail               = "test@example.com"
	testSubject             = "auth0|test123"
	testIDTokenPayload      = `{"email":"test@example.com","sub":"auth0|test123"}`
	testAccessTokenPayload  = `{"sub":"auth0|test123"}`
	testInvalidIDTokenParts = "invalid-token-too-many-parts"
)

type MockService struct {
	setUserEmailFunc func(ctx context.Context, accountID, email string) error
	setUserEmailErr  error
}

func (m *MockService) Fetch(_ context.Context, _ *url.URL) (*content.FetchedContent, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) ParseHTML(_ context.Context, _ *content.FetchedContent) (*html.Node, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) Clean(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) GenerateEPUB(_ *model.Article) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) SendArticle(
	_ context.Context,
	_ string,
	_ io.ReadCloser,
	_ string,
) (*email.SendEmailResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) SendArticleByID(_ context.Context, _, _ string) (*servicetypes.SendArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) CreateArticle(_ context.Context, _ *url.URL, _ string) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) UpdateArticle(_ context.Context, _ *model.Article) error {
	return errors.New("not implemented")
}

func (m *MockService) GetArticle(_ context.Context, _, _ string) (*model.Article, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) GetArticlesMetadata(
	_ context.Context, _ string, _, _ int, _ *bool,
) (*servicetypes.GetArticlesResult, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) DeleteArticle(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) DeleteAllArticles(_ context.Context, _ string) (*servicetypes.DeleteArticleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) GetDBError() error {
	return errors.New("not implemented")
}

func (m *MockService) GetUserDeviceEmailAndAutoSend(
	_ context.Context, _ string,
) (deviceEmail string, autoSend bool, err error) {
	return "", false, errors.New("not implemented")
}

func (m *MockService) SetUserDeviceEmailWithAutoSend(_ context.Context, _, _ string, _ bool) error {
	return errors.New("not implemented")
}

func (m *MockService) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *MockService) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) SetUserEmail(_ context.Context, accountID, userEmail string) error {
	if m.setUserEmailFunc != nil {
		return m.setUserEmailFunc(nil, accountID, userEmail)
	}
	return m.setUserEmailErr
}

func (m *MockService) DeleteUserProfile(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *MockService) ToggleFavorite(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockService) CountSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return 0, errors.New("not implemented")
}

func (m *MockService) HandleBounce(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

func (m *MockService) IsEmailBouncing(_ context.Context, _, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockService) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	return "", errors.New("not implemented")
}

var _ service.Interface = (*MockService)(nil)

func TestAuthTokenExchange_MissingCode(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, http.DefaultClient, nil)

	body := types.AuthTokenExchangeRequest{
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

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
	h := New(cfg, nil, http.DefaultClient, nil)

	body := types.AuthTokenExchangeRequest{
		Code: "test-code",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

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
	h := New(cfg, nil, http.DefaultClient, nil)

	req := httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/v1/auth/token",
		bytes.NewReader([]byte("invalid json")),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthTokenExchange_ExplicitGrantType(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != OauthTokenPath {
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
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, "https://"),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
		GrantType:   "authorization_code",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExtractEmailFromIDToken(t *testing.T) {
	tests := []struct {
		name    string
		idToken string
		want    string
		wantErr bool
	}{
		{
			name:    "valid id token with email",
			idToken: "header." + base64.RawURLEncoding.EncodeToString([]byte(testIDTokenPayload)) + ".signature",
			want:    testEmail,
			wantErr: false,
		},
		{
			name:    "invalid format - missing parts",
			idToken: "header.payload",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			idToken: testInvalidIDTokenParts,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid base64 in payload",
			idToken: "header.!!!notbase64!!!.signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid json in payload",
			idToken: "header." + base64Encode(`not valid json`) + ".signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "missing email claim",
			idToken: "header." + base64Encode(`{"sub":"test123"}`) + ".signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty email claim",
			idToken: "header." + base64Encode(`{"email":""}`) + ".signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - missing parts",
			idToken: "header.payload",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			idToken: testInvalidIDTokenParts,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid base64 in payload",
			idToken: "header.!!!notbase64!!!.signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid json in payload",
			idToken: "header." + base64Encode(`not valid json`) + ".signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			idToken: testInvalidIDTokenParts,
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty email claim",
			idToken: "header." + base64Encode(`{"email":""}`) + ".signature",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractEmailFromIDToken(tt.idToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractEmailFromIDToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractEmailFromIDToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSubjectFromIDToken(t *testing.T) {
	tests := []struct {
		name    string
		idToken string
		want    string
		wantErr bool
	}{
		{
			name:    "valid id token with sub",
			idToken: "header." + base64Encode(`{"sub":"auth0|test123"}`) + ".signature",
			want:    testSubject,
			wantErr: false,
		},
		{
			name:    "invalid format - missing parts",
			idToken: "header.payload",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			idToken: testInvalidIDTokenParts,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid base64 in payload",
			idToken: "header.!!!notbase64!!!.signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid json in payload",
			idToken: "header." + base64Encode(`not valid json`) + ".signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "missing sub claim",
			idToken: "header." + base64Encode(`{"email":"test@example.com"}`) + ".signature",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty sub claim",
			idToken: "header." + base64Encode(`{"sub":""}`) + ".signature",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSubjectFromIDToken(tt.idToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSubjectFromIDToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSubjectFromIDToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleAuth0Error(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "error with description",
			body:           []byte(`{"error":"invalid_grant","error_description":"Invalid authorization code"}`),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "error without description",
			body:           []byte(`{"error":"invalid_request"}`),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid json body",
			body:           []byte(`not valid json`),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty body",
			body:           []byte(`{}`),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := context.Background()
			h := &Handlers{}
			h.handleAuth0Error(ctx, w, tt.body)

			if w.Code != tt.expectedStatus {
				t.Errorf("handleAuth0Error() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			var resp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			if resp.Error == "" {
				t.Errorf("handleAuth0Error() response should contain error message")
			}
		})
	}
}

func TestStoreUserEmail(t *testing.T) {
	tests := []struct {
		name        string
		service     service.Interface
		accessToken string
		email       string
		wantErr     bool
	}{
		{
			name:    "service not configured",
			service: nil,
			wantErr: true,
		},
		{
			name:        "valid service and token",
			service:     &MockService{setUserEmailErr: nil},
			accessToken: "header." + base64Encode(`{"sub":"auth0|test123"}`) + ".signature",
			email:       "test@example.com",
			wantErr:     false,
		},
		{
			name:        "service returns error",
			service:     &MockService{setUserEmailErr: errors.New("service error")},
			accessToken: "header." + base64Encode(`{"sub":"auth0|test123"}`) + ".signature",
			email:       "test@example.com",
			wantErr:     true,
		},
		{
			name:        "invalid access token format",
			service:     &MockService{},
			accessToken: "invalid.token",
			email:       "test@example.com",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			h := &Handlers{
				cfg:     cfg,
				service: tt.service,
			}
			ctx := context.Background()

			err := h.storeUserEmail(ctx, tt.email, tt.accessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("storeUserEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthTokenExchange_Auth0Error(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "Invalid authorization code",
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
	h := New(cfg, nil, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Error != "Invalid authorization code" {
		t.Errorf("expected error 'Invalid authorization code', got '%s'", resp.Error)
	}
}

func TestAuthTokenExchange_WithIDToken(t *testing.T) {
	var storedAccountID, storedEmail string
	mockService := &MockService{
		setUserEmailFunc: func(_ context.Context, accountID, email string) error {
			storedAccountID = accountID
			storedEmail = email
			return nil
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(types.AuthTokenExchangeResponse{ //nolint:gosec // test mock token, not a real secret
			AccessToken: "header." + base64.RawURLEncoding.EncodeToString([]byte(testAccessTokenPayload)) + ".signature",
			IDToken:     "header." + base64.RawURLEncoding.EncodeToString([]byte(testIDTokenPayload)) + ".signature",
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
	h := New(cfg, mockService, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp types.AuthTokenExchangeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Email != testEmail {
		t.Errorf("expected email '%s', got '%s'", testEmail, resp.Email)
	}

	if storedAccountID != testSubject {
		t.Errorf("expected stored account ID '%s', got '%s'", testSubject, storedAccountID)
	}

	if storedEmail != testEmail {
		t.Errorf("expected stored email '%s', got '%s'", testEmail, storedEmail)
	}
}

func TestAuthTokenExchange_InvalidIDToken(t *testing.T) {
	mockService := &MockService{}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(types.AuthTokenExchangeResponse{ //nolint:gosec // test mock token, not a real secret
			AccessToken: "test-access-token",
			IDToken:     "invalid.token",
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
	h := New(cfg, mockService, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp types.AuthTokenExchangeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Email != "" {
		t.Errorf("expected empty email when id_token is invalid, got '%s'", resp.Email)
	}
}

func TestAuthTokenExchange_ServiceError(t *testing.T) {
	mockService := &MockService{
		setUserEmailErr: errors.New("database error"),
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(types.AuthTokenExchangeResponse{ //nolint:gosec // test mock token, not a real secret
			AccessToken: "header." + base64.RawURLEncoding.EncodeToString([]byte(testAccessTokenPayload)) + ".signature",
			IDToken:     "header." + base64.RawURLEncoding.EncodeToString([]byte(testIDTokenPayload)) + ".signature",
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
	h := New(cfg, mockService, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp types.AuthTokenExchangeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Email != testEmail {
		t.Errorf("expected email '%s' even with service error, got '%s'", testEmail, resp.Email)
	}
}

func TestAuthTokenExchange_Auth0InvalidJSON(t *testing.T) { //nolint:dupl // tests Auth0 400 error with invalid JSON
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
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
	h := New(cfg, nil, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthTokenExchange_Auth0SuccessInvalidJSON(t *testing.T) { //nolint:dupl // tests Auth0 200 with invalid JSON
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
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
	h := New(cfg, nil, client, nil)

	body := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestBuildTokenRequest(t *testing.T) {
	cfg := &config.Config{
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
	}
	h := &Handlers{
		cfg: cfg,
	}

	req := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
		GrantType:   "authorization_code",
	}

	httpReq := h.buildTokenRequest(req)

	if httpReq.Method != "POST" {
		t.Errorf("expected method POST, got %s", httpReq.Method)
	}

	if httpReq.URL.String() != "https://test.auth0.com/oauth/token" {
		t.Errorf("expected URL 'https://test.auth0.com/oauth/token', got '%s'", httpReq.URL.String())
	}

	if httpReq.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("expected Content-Type 'application/x-www-form-urlencoded', got '%s'", httpReq.Header.Get("Content-Type"))
	}

	bodyBytes, _ := io.ReadAll(httpReq.Body)
	bodyStr := string(bodyBytes)

	if !strings.Contains(bodyStr, "code=test-code") {
		t.Errorf("expected body to contain 'code=test-code', got '%s'", bodyStr)
	}

	if !strings.Contains(bodyStr, "client_id=test-client-id") {
		t.Errorf("expected body to contain 'client_id=test-client-id', got '%s'", bodyStr)
	}

	if !strings.Contains(bodyStr, "redirect_uri=http%3A%2F%2Flocalhost%2Fcallback") {
		t.Errorf("expected body to contain encoded redirect_uri, got '%s'", bodyStr)
	}

	if !strings.Contains(bodyStr, "grant_type=authorization_code") {
		t.Errorf("expected body to contain 'grant_type=authorization_code', got '%s'", bodyStr)
	}
}

func TestBuildTokenRequest_DefaultGrantType(t *testing.T) {
	cfg := &config.Config{
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: "test-client-secret",
	}
	h := &Handlers{
		cfg: cfg,
	}

	req := types.AuthTokenExchangeRequest{
		Code:        "test-code",
		RedirectURI: "http://localhost/callback",
		GrantType:   "",
	}

	httpReq := h.buildTokenRequest(req)

	bodyBytes, _ := io.ReadAll(httpReq.Body)
	bodyStr := string(bodyBytes)

	if !strings.Contains(bodyStr, "grant_type=") {
		t.Errorf("expected body to contain grant_type, got '%s'", bodyStr)
	}

	if !strings.Contains(bodyStr, "grant_type=&") {
		t.Errorf("expected grant_type to be empty when not provided, got '%s'", bodyStr)
	}
}

func base64Encode(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
