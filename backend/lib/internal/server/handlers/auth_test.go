package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/paseto"
	"github.com/shaftoe/savetoink/backend/lib/internal/server/types"
	"github.com/shaftoe/savetoink/backend/lib/internal/service/servicetypes"
	internaltype "github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

//nolint:gosec // test constants, not real credentials
const (
	testEmail               = "test@example.com"
	testSubject             = "auth0|test123"
	testIDTokenPayload      = `{"email":"test@example.com","sub":"auth0|test123"}`
	testAccessTokenPayload  = `{"sub":"auth0|test123"}`
	testInvalidIDTokenParts = "invalid-token-too-many-parts"
	testClientSecret        = "test-client-secret"
	contentTypeHeader       = "Content-Type"
	contentTypeJSON         = "application/json"
	testCode                = "test-code"
	expectedStatusFormat    = "expected status %d, got %d"
	httpsScheme             = "https://"
	testAuth0TokenURL       = "https://test.auth0.com/oauth/token"
	testPasetoKey           = "QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE=" //nolint:gosec // test key
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

func (m *MockService) ReadEPUB(_ context.Context, _ *url.URL) (io.ReadCloser, string, error) {
	return nil, "", errors.New("not implemented")
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
	_ context.Context, _ string, _, _ int, _ *internaltype.ArticleFilter,
) (*servicetypes.GetArticlesResult, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) DeleteArticle(_ context.Context, _, _ string) (*servicetypes.DeleteArticleResult, error) {
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

func (m *MockService) AddArticleTags(_ context.Context, _, _ string, _ []string) error {
	return errors.New("not implemented")
}

func (m *MockService) RemoveArticleTags(_ context.Context, _, _ string, _ []string) error {
	return errors.New("not implemented")
}

func (m *MockService) SetArticleTags(_ context.Context, _, _ string, _ []string) error {
	return errors.New("not implemented")
}

func (m *MockService) GetArticleTags(_ context.Context, _, _ string) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (m *MockService) GetAllTagsForAccount(_ context.Context, _ string) ([]string, error) {
	return nil, errors.New("not implemented")
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
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, http.DefaultClient, nil, nil)

	body := types.AuthTokenExchangeRequest{
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf(expectedStatusFormat, http.StatusBadRequest, w.Code)
	}
}

func TestAuthTokenExchange_MissingRedirectURI(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, http.DefaultClient, nil, nil)

	body := types.AuthTokenExchangeRequest{
		Code: testCode,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf(expectedStatusFormat, http.StatusBadRequest, w.Code)
	}
}

func TestAuthTokenExchange_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, http.DefaultClient, nil, nil)

	req := httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/v1/auth/token",
		bytes.NewReader([]byte("invalid json")),
	)
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf(expectedStatusFormat, http.StatusBadRequest, w.Code)
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
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
		GrantType:   "authorization_code",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(expectedStatusFormat, http.StatusOK, w.Code)
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
		name        string
		body        []byte
		wantErr     bool
		errContains string
	}{
		{
			name:        "error with description",
			body:        []byte(`{"error":"invalid_grant","error_description":"Invalid authorization code"}`),
			wantErr:     true,
			errContains: "Invalid authorization code",
		},
		{
			name:        "error without description",
			body:        []byte(`{"error":"invalid_request"}`),
			wantErr:     true,
			errContains: "invalid_request",
		},
		{
			name:    "invalid json body",
			body:    []byte(`not valid json`),
			wantErr: true,
		},
		{
			name:    "empty body",
			body:    []byte(`{}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handlers{}
			ctx := context.Background()

			err := h.auth0Error(ctx, tt.body)

			if (err != nil) != tt.wantErr {
				t.Errorf("auth0Error() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("auth0Error() error = %v, want containing %v", err, tt.errContains)
				}
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
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf(expectedStatusFormat, http.StatusUnauthorized, w.Code)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Error != "auth0 rejected the token exchange: Invalid authorization code" {
		t.Errorf("expected wrapped error, got '%s'", resp.Error)
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
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, mockService, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(expectedStatusFormat, http.StatusOK, w.Code)
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
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, mockService, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(expectedStatusFormat, http.StatusOK, w.Code)
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
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, mockService, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(expectedStatusFormat, http.StatusOK, w.Code)
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
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf(expectedStatusFormat, http.StatusUnauthorized, w.Code)
	}
}

func TestAuthTokenExchange_Auth0SuccessInvalidJSON(t *testing.T) { //nolint:dupl // tests Auth0 200 with invalid JSON
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(contentTypeHeader, contentTypeJSON)
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
		Auth0Domain:       strings.TrimPrefix(parsedURL.Host, httpsScheme),
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	h := New(cfg, nil, client, nil, newTestPasetoKeyStore(t))

	body := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/token", bytes.NewReader(bodyBytes))
	req.Header.Set(contentTypeHeader, contentTypeJSON)
	w := httptest.NewRecorder()

	h.HandleAuthTokenExchange(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf(expectedStatusFormat, http.StatusInternalServerError, w.Code)
	}
}

func TestBuildTokenRequest(t *testing.T) {
	cfg := &config.Config{
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
	}
	h := &Handlers{
		cfg: cfg,
	}

	req := types.AuthTokenExchangeRequest{
		Code:        testCode,
		RedirectURI: "http://localhost/callback",
		GrantType:   "authorization_code",
	}

	httpReq := h.buildTokenRequest(req)

	if httpReq.Method != "POST" {
		t.Errorf("expected method POST, got %s", httpReq.Method)
	}

	if httpReq.URL.String() != testAuth0TokenURL {
		t.Errorf("expected URL '%s', got '%s'", testAuth0TokenURL, httpReq.URL.String())
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
		Auth0ClientSecret: testClientSecret,
	}
	h := &Handlers{
		cfg: cfg,
	}

	req := types.AuthTokenExchangeRequest{
		Code:        testCode,
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

func newTestPasetoKeyStore(t *testing.T) *paseto.KeyStore {
	t.Helper()
	ks, err := paseto.NewKeyStore(paseto.KeyStoreConfig{
		SymmetricKey: testPasetoKey,
		KeyVersion:   "v1",
	})
	if err != nil {
		t.Fatalf("failed to create test paseto keystore: %v", err)
	}
	return ks
}

func TestHandleAuthRefresh(t *testing.T) {
	cfg := &config.Config{
		AuthBackend:       consts.AuthBackendAuth0,
		Auth0Domain:       "test.auth0.com",
		Auth0ClientID:     "test-client-id",
		Auth0ClientSecret: testClientSecret,
		Auth0Audience:     "test-audience",
	}
	ks := newTestPasetoKeyStore(t)
	h := New(cfg, nil, http.DefaultClient, nil, ks)

	t.Run("valid refresh token returns new pair", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(paseto.TokenClaims{
			Subject: "auth0|123",
			Email:   "user@example.com",
		})
		require.NoError(t, err)

		body := fmt.Sprintf(`{"refresh_token":%q}`, pair.RefreshToken)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/refresh", strings.NewReader(body))
		req.Header.Set(contentTypeHeader, contentTypeJSON)
		w := httptest.NewRecorder()

		h.HandleAuthRefresh(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp types.AuthTokenExchangeResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.Equal(t, int(consts.DefaultAccessTTL.Seconds()), resp.ExpiresIn)
		assert.True(t, strings.HasPrefix(resp.AccessToken, "v4.local."))
		assert.True(t, strings.HasPrefix(resp.RefreshToken, "v4.local."))
		assert.NotEqual(t, pair.AccessToken, resp.AccessToken)
	})

	t.Run("missing refresh token returns 400", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/refresh", strings.NewReader(body))
		req.Header.Set(contentTypeHeader, contentTypeJSON)
		w := httptest.NewRecorder()

		h.HandleAuthRefresh(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid refresh token returns 401", func(t *testing.T) {
		body := `{"refresh_token":"v4.local.invalidtoken"}`
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/refresh", strings.NewReader(body))
		req.Header.Set(contentTypeHeader, contentTypeJSON)
		w := httptest.NewRecorder()

		h.HandleAuthRefresh(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("access token used as refresh returns 401", func(t *testing.T) {
		pair, err := ks.GenerateTokenPair(paseto.TokenClaims{
			Subject: "auth0|123",
			Email:   "user@example.com",
		})
		require.NoError(t, err)

		body := fmt.Sprintf(`{"refresh_token":%q}`, pair.AccessToken)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/refresh", strings.NewReader(body))
		req.Header.Set(contentTypeHeader, contentTypeJSON)
		w := httptest.NewRecorder()

		h.HandleAuthRefresh(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("nil key store returns 500", func(t *testing.T) {
		hNoKS := New(cfg, nil, http.DefaultClient, nil, nil)

		body := `{"refresh_token":"sometoken"}`
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/refresh", strings.NewReader(body))
		req.Header.Set(contentTypeHeader, contentTypeJSON)
		w := httptest.NewRecorder()

		hNoKS.HandleAuthRefresh(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
